package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	"weblineBackend/internal/app_errors"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/database"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Simple in-memory cache implementation
type categoryCache struct {
	sync.RWMutex
	categories          map[string][]database.ListCategoriesRow
	popularCategories   []database.GetPopularCategoriesRow
	hierarchyCache      map[string][]*model.V2CategoryHierarchy
	categoryDetailsMap  map[string]*model.V2CategoryDetail
	expiry              map[string]time.Time
	statsCache          map[string]*model.CategoryHierarchyStats    // Add a cache for hierarchy stats
	directChildrenCache map[string][]model.CategoryWithProductCount // Cache for direct children with stats
}

func newCategoryCache() *categoryCache {
	return &categoryCache{
		categories:          make(map[string][]database.ListCategoriesRow),
		hierarchyCache:      make(map[string][]*model.V2CategoryHierarchy),
		categoryDetailsMap:  make(map[string]*model.V2CategoryDetail),
		expiry:              make(map[string]time.Time),
		statsCache:          make(map[string]*model.CategoryHierarchyStats),
		directChildrenCache: make(map[string][]model.CategoryWithProductCount),
	}
}

func (c *categoryCache) set(key string, value interface{}, duration time.Duration) {
	c.Lock()
	defer c.Unlock()

	expiryTime := time.Now().Add(duration)
	c.expiry[key] = expiryTime

	switch v := value.(type) {
	case []database.ListCategoriesRow:
		c.categories[key] = v
	case []*model.V2CategoryHierarchy:
		c.hierarchyCache[key] = v
	case *model.V2CategoryDetail:
		c.categoryDetailsMap[key] = v
	case []database.GetPopularCategoriesRow:
		c.popularCategories = v
	case *model.CategoryHierarchyStats:
		c.statsCache[key] = v
	case []model.CategoryWithProductCount:
		c.directChildrenCache[key] = v
	}
}

func (c *categoryCache) get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()

	expiry, exists := c.expiry[key]
	if !exists || time.Now().After(expiry) {
		return nil, false
	}

	// Check different cache types based on key prefix
	if key == "popular" {
		return c.popularCategories, true
	}

	if strings.HasPrefix(key, "direct_children_") && len(c.directChildrenCache[key]) > 0 {
		return c.directChildrenCache[key], true
	}

	if strings.HasPrefix(key, "stats_") && c.statsCache[key] != nil {
		return c.statsCache[key], true
	}

	if strings.HasPrefix(key, "details_") && c.categoryDetailsMap[key] != nil {
		return c.categoryDetailsMap[key], true
	}

	if strings.HasPrefix(key, "hierarchy_") && c.hierarchyCache[key] != nil {
		return c.hierarchyCache[key], true
	}

	if c.categories[key] != nil {
		return c.categories[key], true
	}

	return nil, false
}

func (c *categoryCache) invalidate(keyPrefix string) {
	c.Lock()
	defer c.Unlock()

	keysToDelete := make([]string, 0, len(c.expiry)/2) // Preallocate with estimated size

	// Identify keys to delete
	for k := range c.expiry {
		if keyPrefix == "all" || strings.HasPrefix(k, keyPrefix) {
			keysToDelete = append(keysToDelete, k)
		}
	}

	// Delete identified keys
	for _, k := range keysToDelete {
		delete(c.expiry, k)
	}

	// Clear specific maps based on prefix
	if keyPrefix == "categories" || keyPrefix == "all" {
		c.categories = make(map[string][]database.ListCategoriesRow)
	}
	if keyPrefix == "hierarchy" || keyPrefix == "all" {
		c.hierarchyCache = make(map[string][]*model.V2CategoryHierarchy)
	}
	if keyPrefix == "details" || keyPrefix == "all" {
		c.categoryDetailsMap = make(map[string]*model.V2CategoryDetail)
	}
	if keyPrefix == "stats" || keyPrefix == "all" {
		c.statsCache = make(map[string]*model.CategoryHierarchyStats)
	}
	if keyPrefix == "direct_children" || keyPrefix == "all" {
		c.directChildrenCache = make(map[string][]model.CategoryWithProductCount)
	}
	if keyPrefix == "popular" || keyPrefix == "all" {
		c.popularCategories = nil
	}
}

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
	userRepo     *repository.UserRepository
	logger       *zap.Logger
	config       *appconfig.Config
	s3Client     *s3.Client
	cache        *categoryCache
}

func NewCategoryService(categoryRepo *repository.CategoryRepository, userRepo *repository.UserRepository, logger *zap.Logger, cfg *appconfig.Config, s3Client *s3.Client) *CategoryService {
	service := &CategoryService{
		categoryRepo: categoryRepo,
		userRepo:     userRepo,
		logger:       logger,
		config:       cfg,
		s3Client:     s3Client,
		cache:        newCategoryCache(),
	}

	// Start periodic cache cleanup
	service.schedulePeriodicCacheCleanup()

	// Prefetch popular categories on startup
	go func() {
		// Wait a short time for application to fully initialize
		time.Sleep(5 * time.Second)
		service.prefetchPopularCategories(context.Background())
	}()

	return service
}

// CreateCategoryService creates a new category
func (s *CategoryService) CreateCategoryService(
	ctx context.Context,
	params *model.CreateCategoryParams,
	image *model.ImageFile,
) error {
	// Validate inputs before processing
	if params == nil {
		return fmt.Errorf("category parameters cannot be nil")
	}

	if params.Name == "" {
		return fmt.Errorf("category name cannot be empty")
	}

	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	if err := s.verifyAdminStatus(ctx, userID); err != nil {
		return err
	}

	categoryID, err := s.categoryRepo.GetCategoryBySlug(ctx, params.Slug)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return s.logAndReturnError("failed to get category by slug", err)
	}

	var result error
	if categoryID != nil {
		result = s.updateExistingCategory(ctx, params, image, categoryID)
	} else {
		result = s.createNewCategory(ctx, params, image)
	}

	// Invalidate cache after creation/update
	if result == nil {
		s.cache.invalidate("all")
	}

	return result
}

func (s *CategoryService) createNewCategory(ctx context.Context, params *model.CreateCategoryParams, image *model.ImageFile) error {
	// parse parentID to null uuid
	var parentIDValue uuid.NullUUID
	if params.ParentID != "" {
		id, err := uuid.Parse(params.ParentID)
		if err != nil {
			s.logger.Error("failed to parse parent ID", zap.Error(err))
			return err
		}
		parentIDValue = uuid.NullUUID{
			UUID:  id,
			Valid: true,
		}
	} else {
		parentIDValue = uuid.NullUUID{
			UUID:  uuid.Nil,
			Valid: false,
		}
	}

	// upload image to S3
	filePath, err := s.handleCategoryImage(ctx, image, "")
	if err != nil {
		return s.logAndReturnError("failed to handle category image", err)
	}

	categoryParams := database.CreateCategoryParams{
		Name:            params.Name,
		Description:     optionalString(params.Description),
		MetaTitle:       optionalString(params.MetaTitle),
		MetaDescription: optionalString(params.MetaDescription),
		ParentID:        parentIDValue,
		ImageUrl:        optionalString(filePath),
	}

	err = s.categoryRepo.CreateCategory(ctx, categoryParams)
	if err != nil {
		s.logger.Error("failed to create category", zap.Error(err))
		return err
	}

	return nil
}

func (s *CategoryService) updateExistingCategory(ctx context.Context, params *model.CreateCategoryParams, image *model.ImageFile, categoryID *uuid.UUID) error {
	// updated by
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// parse parentID to null uuid
	var parentIDValue uuid.NullUUID
	if params.ParentID != "" {
		id, err := uuid.Parse(params.ParentID)
		if err != nil {
			s.logger.Error("failed to parse parent ID", zap.Error(err))
			return err
		}
		parentIDValue = uuid.NullUUID{
			UUID:  id,
			Valid: true,
		}
	} else {
		parentIDValue = uuid.NullUUID{
			UUID:  uuid.Nil,
			Valid: false,
		}
	}

	// get existing category
	existingCategory, err := s.categoryRepo.GetCategoryByID(ctx, *categoryID)
	if err != nil {
		s.logger.Error("failed to get category by ID", zap.Error(err))
		return err
	}

	// upload image to S3
	filePath, err := s.handleCategoryImage(ctx, image, existingCategory.ImageUrl.String)
	if err != nil {
		return s.logAndReturnError("failed to handle category image", err)
	}

	// update category
	categoryParams := database.UpdateCategoryParams{
		ID:              *categoryID,
		Name:            params.Name,
		ParentID:        parentIDValue,
		ImageUrl:        optionalString(filePath),
		MetaTitle:       optionalString(params.MetaTitle),
		MetaDescription: optionalString(params.MetaDescription),
		LastUpdatedBy: uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
	}

	err = s.categoryRepo.UpdateCategory(ctx, categoryParams)
	if err != nil {
		s.logger.Error("failed to update category", zap.Error(err))
		return err
	}

	return nil
}

func (s *CategoryService) handleCategoryImage(ctx context.Context, image *model.ImageFile, existingImageUrl string) (string, error) {
	if image == nil {
		return existingImageUrl, nil
	}

	// Check file size before processing
	if image.FileHeader.Size > 10*1024*1024 { // 10MB limit
		return "", fmt.Errorf("image size exceeds 10MB limit")
	}

	// Check file type
	fileType := image.FileHeader.Header.Get("Content-Type")
	if fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/webp" {
		return "", fmt.Errorf("unsupported image format: %s. Only JPEG, PNG and WebP are supported", fileType)
	}

	// Upload the image to S3 with optimization
	filePath, err := utils.UploadCustomFileToS3(ctx, image.File, image.FileHeader, s.s3Client, s.config.AWSBucketName, "category_images")
	if err != nil {
		return "", s.logAndReturnError("failed to upload category image", err)
	}

	// Delete old image if it exists and is different
	if existingImageUrl != "" && existingImageUrl != filePath {
		// Run deletion in a goroutine to not block the main flow
		go func(ctx context.Context, oldPath string) {
			if err := utils.DeleteFileFromS3(ctx, s.s3Client, s.config.AWSBucketName, oldPath); err != nil {
				s.logger.Warn("failed to delete old category image",
					zap.Error(err),
					zap.String("path", oldPath))
			}
		}(context.Background(), existingImageUrl)
	}

	return filePath, nil
}

func (s *CategoryService) getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return uuid.Nil, app_errors.NewUnauthorizedUserError()
	}
	return userID, nil
}

func (s *CategoryService) logAndReturnError(message string, err error) error {
	s.logger.Error(message, zap.Error(err))
	return fmt.Errorf("%s: %w", message, err)
}

func (s *CategoryService) verifyAdminStatus(ctx context.Context, userID uuid.UUID) error {
	isAdmin, err := s.userRepo.IsAdmin(ctx, userID)
	if err != nil || !isAdmin {
		return app_errors.NewUnauthorizedUserError()
	}
	return nil
}

// GetCategoryByIDService retrieves a category by its ID
func (s *CategoryService) GetCategoryByIDService(ctx context.Context, id string) (database.GetCategoryByIDRow, error) {
	// parse id to uuid
	categoryID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("failed to parse category ID", zap.Error(err))
		return database.GetCategoryByIDRow{}, err
	}

	category, err := s.categoryRepo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to get category by ID", zap.Error(err))
		return database.GetCategoryByIDRow{}, err
	}

	return category, nil
}

// GetCategoriesService retrieves all categories
func (s *CategoryService) GetCategoriesService(ctx context.Context) ([]database.ListCategoriesRow, error) {
	// Check cache first
	cacheKey := "categories_active"
	if cached, found := s.cache.get(cacheKey); found {
		if categories, ok := cached.([]database.ListCategoriesRow); ok {
			s.logger.Debug("returning categories from cache", zap.Int("count", len(categories)))
			return categories, nil
		}
	}

	// Not in cache, fetch from database
	categories, err := s.categoryRepo.GetCategories(ctx, true) // Only return active categories by default
	if err != nil {
		s.logger.Error("failed to get categories", zap.Error(err))
		return nil, err
	}

	// Cache the result for 10 minutes
	s.cache.set(cacheKey, categories, 10*time.Minute)

	return categories, nil
}

// GetCategoriesServiceWithFilter retrieves categories with filter option
func (s *CategoryService) GetCategoriesServiceWithFilter(ctx context.Context, activeOnly bool) ([]database.ListCategoriesRow, error) {
	// Add context info for debugging
	s.logger.Debug("retrieving categories with filter", zap.Bool("activeOnly", activeOnly))

	categories, err := s.categoryRepo.GetCategories(ctx, activeOnly)
	if err != nil {
		s.logger.Error("failed to get categories with filter",
			zap.Error(err),
			zap.Bool("activeOnly", activeOnly))
		return nil, err
	}

	s.logger.Debug("retrieved categories successfully",
		zap.Int("count", len(categories)),
		zap.Bool("activeOnly", activeOnly))
	return categories, nil
}

// SoftDeleteCategoryService marks a category as inactive
func (s *CategoryService) SoftDeleteCategoryService(ctx context.Context, id string) error {
	// parse id to uuid
	categoryID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("failed to parse category ID", zap.Error(err))
		return err
	}

	err = s.categoryRepo.SoftDeleteCategory(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to soft delete category", zap.Error(err))
		return err
	}

	// Invalidate caches since category structure might have changed
	s.cache.invalidate("all")

	return nil
}

// GetCategoriesByParentIDService retrieves categories by their parent ID
func (s *CategoryService) GetCategoriesByParentIDService(ctx context.Context, parentID string) ([]*model.Category, error) {
	// parse parentID to null uuid
	var parentIDValue uuid.NullUUID
	if parentID != "" {
		id, err := uuid.Parse(parentID)
		if err != nil {
			s.logger.Error("failed to parse parent ID", zap.Error(err))
			return nil, err
		}
		parentIDValue = uuid.NullUUID{
			UUID:  id,
			Valid: true,
		}
	} else {
		parentIDValue = uuid.NullUUID{
			UUID:  uuid.Nil,
			Valid: false,
		}
	}

	categories, err := s.categoryRepo.GetCategoriesByParentID(ctx, parentIDValue)
	if err != nil {
		s.logger.Error("failed to get categories by parent ID", zap.Error(err))
		return nil, err
	}

	categoriesModel := make([]*model.Category, 0, len(categories))
	for _, category := range categories {
		categoryModel := &model.Category{
			ID:   category.ID,
			Name: category.Name,
		}
		categoriesModel = append(categoriesModel, categoryModel)
	}

	return categoriesModel, nil
}

// GetCategoriesWithProductsCountService retrieves categories along with the count of products in each category
func (s *CategoryService) GetCategoriesWithProductsCountService(ctx context.Context) ([]database.GetCategoriesWithProductsCountRow, error) {
	categoriesWithCount, err := s.categoryRepo.GetCategoriesWithProductsCount(ctx)
	if err != nil {
		s.logger.Error("failed to get categories with products count", zap.Error(err))
		return nil, err
	}

	return categoriesWithCount, nil
}

// GetCategoryTreeService retrieves the full category hierarchy
func (s *CategoryService) GetCategoryTreeService(ctx context.Context) ([]database.GetCategoryTreeRow, error) {
	categories, err := s.categoryRepo.GetCategoryTree(ctx)
	if err != nil {
		s.logger.Error("failed to get category tree", zap.Error(err))
		return nil, err
	}

	return categories, nil
}

// CheckCategoryExistenceService checks if a category exists in the database
func (s *CategoryService) CheckCategoryExistenceService(ctx context.Context, id string) (bool, error) {
	// parse id to uuid
	categoryID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("failed to parse category ID", zap.Error(err))
		return false, err
	}

	exists, err := s.categoryRepo.CheckCategoryExistence(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to check category existence", zap.Error(err))
		return false, err
	}

	return exists, nil
}

// GetCategoriesWithSubcategoryCountService retrieves categories along with the count of subcategories for each category
func (s *CategoryService) GetCategoriesWithSubcategoryCountService(ctx context.Context) ([]database.GetCategoriesWithSubcategoryCountRow, error) {
	categoriesWithCount, err := s.categoryRepo.GetCategoriesWithSubcategoryCount(ctx)
	if err != nil {
		s.logger.Error("failed to get categories with subcategory count", zap.Error(err))
		return nil, err
	}

	return categoriesWithCount, nil
}

// GetParentCategoriesService retrieves parent categories
func (s *CategoryService) GetParentCategoriesService(ctx context.Context) ([]model.Category, error) {
	categories, err := s.categoryRepo.GetParentCategories(ctx)
	if err != nil {
		s.logger.Error("failed to get parent categories", zap.Error(err))
		return nil, err
	}

	categoriesModel := make([]model.Category, 0)
	for _, category := range categories {
		imageUrl := ""
		if category.ImageUrl.Valid {
			imageUrl = s.constructS3URL(category.ImageUrl.String)
		}
		categoryModel := model.Category{
			ID:       category.ID,
			Name:     category.Name,
			ImageUrl: imageUrl,
		}
		categoriesModel = append(categoriesModel, categoryModel)
	}

	return categoriesModel, nil
}

// GetCategoryByNameService retrieves a category by its name
func (s *CategoryService) GetCategoryByNameService(ctx context.Context, name string) (model.CategoryDetail, error) {
	// get category by name
	category, err := s.categoryRepo.GetCategoryByName(ctx, name)
	if err != nil {
		s.logger.Error("failed to get category by name", zap.Error(err))
		return model.CategoryDetail{}, err
	}

	// get categories by parent ID
	categories, err := s.categoryRepo.GetCategoriesByParentID(ctx, uuid.NullUUID{UUID: category.ID, Valid: true})
	if err != nil {
		s.logger.Error("failed to get categories by parent ID", zap.Error(err))
		return model.CategoryDetail{}, err
	}

	//// get available colors for the category
	//availableColors, err := s.productColorRepo.GetAvailableColorsByCategoryID(ctx, category.ID)
	//if err != nil {
	//	s.logger.Error("failed to get available colors by category ID", zap.Error(err))
	//	return model.CategoryDetail{}, err
	//}
	//
	//var colors []string
	//for _, color := range *availableColors {
	//	colors = append(colors, color.ColorName)
	//}

	// create a new category model
	var subCategories []model.CategoryDetail
	for _, category := range categories {
		// check if image is valid
		var imageURL string
		if category.ImageUrl.Valid {
			imageURL = s.constructS3URL(category.ImageUrl.String)
		} else {
			imageURL = ""
		}

		subCategory := model.CategoryDetail{
			ID:       category.ID,
			Name:     category.Name,
			ParentID: category.ParentID.UUID,
			ImageURL: imageURL,
			IsActive: category.IsActive,
		}
		subCategories = append(subCategories, subCategory)
	}

	categoryModel := model.CategoryDetail{
		ID:            category.ID,
		Name:          category.Name,
		ParentID:      category.ParentID.UUID,
		IsActive:      category.IsActive,
		SubCategories: subCategories,
		//AvailableColors: colors,
	}

	return categoryModel, nil
}

type Product struct {
	ProductID   string `json:"ProductID"`
	ProductName string `json:"ProductName"`
	ImageURL    string `json:"ImageURL"`
}

type Category struct {
	CategoryID       string      `json:"CategoryID"`
	CategoryName     string      `json:"CategoryName"`
	FeaturedProducts []Product   `json:"FeaturedProducts"`
	Processors       []string    `json:"Processors"`
	Size             []string    `json:"Size"`
	Storage          []string    `json:"Storage"`
	Children         []*Category `json:"Children"`
	Position         int         `json:"-"` // Exclude from JSON output
	ParentID         uuid.UUID   `json:"-"` // Exclude from JSON output, used for processing
}

// constructS3URL constructs the S3 URL for a given file path
func (s *CategoryService) constructS3URL(filePath string) string {
	return s.optimizedConstructS3URL(filePath)
}

// UpdateCategoryImageService updates the image of a category
func (s *CategoryService) UpdateCategoryImageService(ctx context.Context, r *http.Request, categoryID string) error {
	// Check if the category exists
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		s.logger.Error("failed to parse category ID", zap.Error(err))
		return err
	}

	if exists, err := s.categoryRepo.CheckCategoryExistence(ctx, categoryUUID); err != nil {
		s.logger.Error("failed to check if category exists", zap.Error(err))
		return err
	} else if !exists {
		s.logger.Error("category does not exist")
		return fmt.Errorf("category does not exist")
	}

	// Update the image of the category
	filePath, err := utils.UploadFileToS3(ctx, r, s.s3Client, s.config.AWSBucketName, "category_images")
	if err != nil {
		s.logger.Error("failed to upload file to S3", zap.Error(err))
		return fmt.Errorf("failed to upload file to S3")
	}

	_, err = s.categoryRepo.UpdateCategoryImage(ctx, categoryUUID, filePath)
	if err != nil {
		s.logger.Error("failed to update category image", zap.Error(err))
		return err
	}

	return nil
}

// GetCollectionCategoriesService retrieves categories that are collections
func (s *CategoryService) GetCollectionCategoriesService(ctx context.Context) ([]model.Collection, error) {
	// get the parent categories
	parentCategories, err := s.categoryRepo.GetParentCategories(ctx)
	if err != nil {
		s.logger.Error("failed to get parent categories", zap.Error(err))
		return nil, err
	}

	// get the child categories
	collections := make([]model.Collection, 0)

	for _, category := range parentCategories {
		if category.Name == "Computing" {
			childCategories, err := s.categoryRepo.GetCategoriesByParentID(ctx, uuid.NullUUID{UUID: category.ID, Valid: true})
			if err != nil {
				s.logger.Error("failed to get child categories", zap.Error(err))
				return nil, err
			}

			for _, childCategory := range childCategories {
				imageUrl := ""
				if childCategory.ImageUrl.Valid {
					imageUrl = s.constructS3URL(childCategory.ImageUrl.String)
				}
				collection := model.Collection{
					ID:         childCategory.ID,
					Name:       childCategory.Name,
					ParentName: category.Name,
					ImageUrl:   imageUrl,
				}
				collections = append(collections, collection)
			}
		} else {
			imageUrl := ""
			if category.ImageUrl.Valid {
				imageUrl = s.constructS3URL(category.ImageUrl.String)
			}
			collection := model.Collection{
				ID:         category.ID,
				Name:       category.Name,
				ParentName: "",
				ImageUrl:   imageUrl,
			}
			collections = append(collections, collection)
		}

	}

	return collections, nil
}

// GetV2CategoryHierarchy retrieves the category hierarchy for the V2 API
func (s *CategoryService) GetV2CategoryHierarchy(
	ctx context.Context,
) ([]*model.V2CategoryHierarchy, error) {
	// By default, only return active categories for the public API
	hierarchy, err := s.categoryRepo.GetV2CategoryHierarchy(ctx, true)
	if err != nil {
		s.logger.Error("failed to get V2 category hierarchy", zap.Error(err))
		return nil, fmt.Errorf("failed to get V2 category hierarchy: %w", err)
	}

	return hierarchy, nil
}

// GetV2CategoryHierarchyWithFilter retrieves the category hierarchy with filter option
func (s *CategoryService) GetV2CategoryHierarchyWithFilter(
	ctx context.Context,
	activeOnly bool,
) ([]*model.V2CategoryHierarchy, error) {
	s.logger.Debug("retrieving category hierarchy", zap.Bool("activeOnly", activeOnly))

	hierarchy, err := s.categoryRepo.GetV2CategoryHierarchy(ctx, activeOnly)
	if err != nil {
		s.logger.Error("failed to get V2 category hierarchy",
			zap.Error(err),
			zap.Bool("activeOnly", activeOnly))
		return nil, fmt.Errorf("failed to get V2 category hierarchy: %w", err)
	}

	s.logger.Debug("retrieved category hierarchy successfully",
		zap.Int("rootCategories", len(hierarchy)),
		zap.Bool("activeOnly", activeOnly))
	return hierarchy, nil
}

// GetCategoryHierarchyStatsService retrieves both hierarchy and stats in a single call
func (s *CategoryService) GetCategoryHierarchyStatsService(
	ctx context.Context,
	categoryID string,
	activeOnly bool,
) (*model.CategoryHierarchyStats, error) {
	// Input validation
	if categoryID == "" {
		return nil, fmt.Errorf("category ID cannot be empty")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("stats_%s_%v", categoryID, activeOnly)
	if cached, found := s.cache.get(cacheKey); found {
		if stats, ok := cached.(*model.CategoryHierarchyStats); ok {
			s.logger.Debug("returning category hierarchy stats from cache",
				zap.String("categoryID", categoryID),
				zap.Bool("activeOnly", activeOnly))
			return stats, nil
		}
	}

	// Parse ID to UUID
	id, err := uuid.Parse(categoryID)
	if err != nil {
		s.logger.Error("failed to parse category ID",
			zap.Error(err),
			zap.String("categoryID", categoryID))
		return nil, fmt.Errorf("invalid category ID format: %w", err)
	}

	// Create context with timeout for this operation
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Get hierarchy stats
	hierarchyStats, err := s.categoryRepo.GetCategoryHierarchyStats(ctxWithTimeout, id, activeOnly)
	if err != nil {
		s.logger.Error("failed to get category hierarchy stats",
			zap.Error(err),
			zap.String("categoryID", categoryID),
			zap.Bool("activeOnly", activeOnly))
		return nil, fmt.Errorf("failed to get category hierarchy stats: %w", err)
	}

	// Update image URLs to S3 URLs if present (and limit processing to necessary items)
	if hierarchyStats.Category.ImageURL != "" {
		hierarchyStats.Category.ImageURL = s.constructS3URL(hierarchyStats.Category.ImageURL)
	}

	// Process children in parallel if there are many
	if len(hierarchyStats.Children) > 10 {
		var wg sync.WaitGroup
		// Use a buffered channel to limit concurrency
		semaphore := make(chan struct{}, 5)

		// Create a mutex to protect the children slice during concurrent updates
		var mu sync.Mutex

		// Process children in batches
		for i := range hierarchyStats.Children {
			wg.Add(1)
			semaphore <- struct{}{} // Acquire semaphore

			go func(idx int) {
				defer wg.Done()
				defer func() { <-semaphore }() // Release semaphore

				child := &hierarchyStats.Children[idx]

				// Truncate excessively long descriptions
				if len(child.Description) > 200 {
					truncated := child.Description[:200] + "..."

					// Update safely with mutex
					mu.Lock()
					child.Description = truncated
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()
	} else {
		// For small number of children, process sequentially
		for i, child := range hierarchyStats.Children {
			if child.Description != "" && len(child.Description) > 200 {
				hierarchyStats.Children[i].Description = child.Description[:200] + "..."
			}
		}
	}

	// Cache the result for 10 minutes
	s.cache.set(cacheKey, hierarchyStats, 10*time.Minute)

	s.logger.Debug("retrieved category hierarchy stats successfully",
		zap.String("categoryID", categoryID),
		zap.Int("childrenCount", len(hierarchyStats.Children)),
		zap.Int("totalProductCount", hierarchyStats.Stats.TotalProductCount))

	return hierarchyStats, nil
}

// DeleteCategoryService deletes a category
func (s *CategoryService) DeleteCategoryService(ctx context.Context, id string) error {
	// parse id to uuid
	categoryID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("failed to parse category ID", zap.Error(err))
		return err
	}

	err = s.categoryRepo.DeleteCategory(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to delete category", zap.Error(err))
		return err
	}

	return nil
}

// GetCategoryDetailsService retrieves the details of a category
func (s *CategoryService) GetCategoryDetailsService(ctx context.Context, slug string) (*model.V2CategoryDetail, error) {
	if slug == "" {
		return nil, fmt.Errorf("category slug cannot be empty")
	}

	// Check cache first
	cacheKey := "details_" + slug
	if cached, found := s.cache.get(cacheKey); found {
		if details, ok := cached.(*model.V2CategoryDetail); ok {
			s.logger.Debug("returning category details from cache", zap.String("slug", slug))
			return details, nil
		}
	}

	category, err := s.categoryRepo.GetCategoryDetailsBySlug(ctx, slug)
	if err != nil {
		s.logger.Error("failed to get category by slug", zap.Error(err), zap.String("slug", slug))
		return nil, err
	}

	// update the image url to s3 url
	if category.ImageURL != "" {
		category.ImageURL = s.constructS3URL(category.ImageURL)
	}

	// Cache the result for 1 hour
	s.cache.set(cacheKey, category, time.Hour)

	return category, nil
}

// GetCategorySEOService retrieves the SEO details of a category
func (s *CategoryService) GetCategorySEOService(ctx context.Context, slug string) (*model.CategorySEO, error) {
	seo, err := s.categoryRepo.GetCategorySEOBySlug(ctx, slug)
	if err != nil {
		s.logger.Error("failed to get category SEO by slug", zap.Error(err))
		return nil, err
	}

	return seo, nil
}

// BatchUpdateCategoryPositionsService updates positions of multiple categories in one operation
func (s *CategoryService) BatchUpdateCategoryPositionsService(
	ctx context.Context,
	updates map[string]int,
) error {
	if len(updates) == 0 {
		return nil // Nothing to update
	}

	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	if err := s.verifyAdminStatus(ctx, userID); err != nil {
		return err
	}

	// Prepare bulk update parameters
	categoryIDs := make([]uuid.UUID, 0, len(updates))
	positions := make([]int32, 0, len(updates))

	for idStr, pos := range updates {
		id, err := uuid.Parse(idStr)
		if err != nil {
			s.logger.Error("invalid category ID in batch update",
				zap.Error(err),
				zap.String("categoryID", idStr))
			return fmt.Errorf("invalid category ID format: %s", idStr)
		}

		categoryIDs = append(categoryIDs, id)
		positions = append(positions, int32(pos))
	}

	// Execute bulk update
	err = s.categoryRepo.BulkUpdateCategoryPositions(ctx, categoryIDs, positions)
	if err != nil {
		s.logger.Error("failed to batch update category positions",
			zap.Error(err),
			zap.Int("count", len(categoryIDs)))
		return fmt.Errorf("failed to batch update category positions: %w", err)
	}

	// Invalidate caches for affected categories
	for _, id := range categoryIDs {
		s.invalidateCategoryCaches(id)
	}

	// Invalidate hierarchy cache as positions affect hierarchy
	s.cache.invalidate("hierarchy")

	s.logger.Info("successfully batch updated category positions",
		zap.Int("count", len(categoryIDs)))
	return nil
}

// GetPopularCategoriesService retrieves popular categories with the most products
func (s *CategoryService) GetPopularCategoriesService(ctx context.Context, limit int32) ([]database.GetPopularCategoriesRow, error) {
	// Check cache first
	cacheKey := "popular"
	if cached, found := s.cache.get(cacheKey); found {
		if categories, ok := cached.([]database.GetPopularCategoriesRow); ok {
			// If more results than needed are cached, just return the requested amount
			if int32(len(categories)) >= limit {
				s.logger.Debug("returning popular categories from cache",
					zap.Int32("limit", limit),
					zap.Int("available", len(categories)))
				if int32(len(categories)) > limit {
					return categories[:limit], nil
				}
				return categories, nil
			}
		}
	}

	// Not in cache or not enough results, fetch from database
	categories, err := s.categoryRepo.GetPopularCategories(ctx, limit)
	if err != nil {
		s.logger.Error("failed to get popular categories",
			zap.Error(err),
			zap.Int32("limit", limit))
		return nil, fmt.Errorf("failed to get popular categories: %w", err)
	}

	// Cache with a longer expiry since this is an expensive operation
	s.cache.set(cacheKey, categories, 30*time.Minute)

	return categories, nil
}

// Prefetch popular categories in background
func (s *CategoryService) prefetchPopularCategories(ctx context.Context) {
	go func() {
		// Create a background context with timeout
		bgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// Get a larger number of popular categories to cache
		categories, err := s.categoryRepo.GetPopularCategories(bgCtx, 20)
		if err != nil {
			s.logger.Warn("failed to prefetch popular categories", zap.Error(err))
			return
		}

		// Cache the results
		s.cache.set("popular", categories, 30*time.Minute)
		s.logger.Debug("prefetched popular categories", zap.Int("count", len(categories)))
	}()
}

// GetDirectChildrenWithStatsService gets only immediate children with their stats
func (s *CategoryService) GetDirectChildrenWithStatsService(
	ctx context.Context,
	slug string,
	activeOnly bool,
) ([]model.CategoryWithProductCount, error) {
	// Input validation
	if slug == "" {
		return nil, fmt.Errorf("category slug cannot be empty")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("direct_children_%s_%v", slug, activeOnly)

	// Force a refresh if requested via context
	_, forceRefresh := ctx.Value("force_refresh").(bool)

	// Check cache first (if not forcing refresh)
	if !forceRefresh {
		if cached, found := s.cache.get(cacheKey); found {
			if children, ok := cached.([]model.CategoryWithProductCount); ok {
				s.logger.Debug("returning direct children from cache",
					zap.String("categorySlug", slug),
					zap.Bool("activeOnly", activeOnly))
				return children, nil
			}
		}
	}

	// Parse slug to UUID
	categoryID, err := s.categoryRepo.GetCategoryBySlug(ctx, slug)
	if err != nil {
		s.logger.Error("failed to get category by slug",
			zap.Error(err),
			zap.String("categorySlug", slug))
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	// Get direct children with stats
	children, err := s.categoryRepo.GetDirectChildrenWithStats(ctx, *categoryID, activeOnly)
	if err != nil {
		s.logger.Error("failed to get direct children with stats",
			zap.Error(err),
			zap.String("categorySlug", slug),
			zap.Bool("activeOnly", activeOnly))
		return nil, fmt.Errorf("failed to get direct children with stats: %w", err)
	}

	// Update image URLs efficiently with the new optimized processor
	children = s.processCategories(children)

	// Cache the result for 10 minutes
	s.cache.set(cacheKey, children, 10*time.Minute)

	s.logger.Debug("retrieved direct children with stats successfully",
		zap.String("categorySlug", slug),
		zap.Int("count", len(children)),
		zap.Bool("activeOnly", activeOnly))

	return children, nil
}

// invalidateCategoryCaches intelligently invalidates only the caches related to a specific category
func (s *CategoryService) invalidateCategoryCaches(categoryID uuid.UUID) {
	s.logger.Debug("invalidating caches for specific category",
		zap.String("categoryID", categoryID.String()))

	// Get a list of keys to invalidate
	s.cache.Lock()
	keysToInvalidate := make([]string, 0)
	idStr := categoryID.String()

	// Find all keys related to this category
	for key := range s.cache.expiry {
		if strings.Contains(key, idStr) {
			keysToInvalidate = append(keysToInvalidate, key)
		}
	}

	// Also invalidate parent hierarchies and popular categories as they might be affected
	keysToInvalidate = append(keysToInvalidate, "popular")
	keysToInvalidate = append(keysToInvalidate, "hierarchy_true")
	keysToInvalidate = append(keysToInvalidate, "hierarchy_false")
	s.cache.Unlock()

	// Perform the actual invalidation
	for _, key := range keysToInvalidate {
		s.cache.Lock()
		delete(s.cache.expiry, key)
		s.cache.Unlock()
	}
}

// schedulePeriodicCacheCleanup sets up regular cleanup of expired cache entries
func (s *CategoryService) schedulePeriodicCacheCleanup() {
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for range ticker.C {
			s.cleanupExpiredCacheEntries()
		}
	}()
}

// cleanupExpiredCacheEntries removes all expired entries from cache
func (s *CategoryService) cleanupExpiredCacheEntries() {
	s.logger.Debug("performing scheduled cache cleanup")
	now := time.Now()

	s.cache.Lock()
	defer s.cache.Unlock()

	// Find expired keys with capacity hint
	expiredKeys := make([]string, 0, len(s.cache.expiry)/4) // Assuming ~25% expired
	for key, expiry := range s.cache.expiry {
		if now.After(expiry) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	// Remove expired entries
	for _, key := range expiredKeys {
		delete(s.cache.expiry, key)

		// Use efficient map clearing for specific caches
		// Process most common prefixes first for efficiency
		if strings.HasPrefix(key, "details_") {
			delete(s.cache.categoryDetailsMap, key)
		} else if strings.HasPrefix(key, "hierarchy_") {
			delete(s.cache.hierarchyCache, key)
		} else if strings.HasPrefix(key, "stats_") {
			delete(s.cache.statsCache, key)
		} else if strings.HasPrefix(key, "direct_children_") {
			delete(s.cache.directChildrenCache, key)
		} else if strings.HasPrefix(key, "categories_") {
			delete(s.cache.categories, key)
		} else if key == "popular" {
			s.cache.popularCategories = nil
		}
	}

	s.logger.Debug("cache cleanup completed", zap.Int("entriesRemoved", len(expiredKeys)))
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// updateImageURLsInChildren updates S3 image URLs in children collection with optimized concurrent processing
func (s *CategoryService) updateImageURLsInChildren(children []model.CategoryWithProductCount) []model.CategoryWithProductCount {
	if len(children) == 0 {
		return children
	}

	// For small collections, process sequentially to avoid goroutine overhead
	if len(children) <= 5 {
		for i, child := range children {
			if child.ImageURL != "" {
				children[i].ImageURL = s.constructS3URL(child.ImageURL)
			}
		}
		return children
	}

	// For larger collections, process concurrently with a worker pool
	var wg sync.WaitGroup
	// Create a worker pool with reasonable concurrency
	concurrency := min(len(children), 10)
	semaphore := make(chan struct{}, concurrency)

	// Create mutex for safe concurrent updates
	var mu sync.Mutex

	for i := range children {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(idx int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			child := &children[idx]
			if child.ImageURL != "" {
				newURL := s.constructS3URL(child.ImageURL)

				mu.Lock()
				child.ImageURL = newURL
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	return children
}

// optimizedConstructS3URL improves S3 URL construction with a consistent format
func (s *CategoryService) optimizedConstructS3URL(filePath string) string {
	// Skip processing if the path is empty or already an HTTPS URL
	if filePath == "" || strings.HasPrefix(filePath, "https://") {
		return filePath
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		s.config.AWSBucketName,
		s.config.AWSRegion,
		strings.TrimPrefix(filePath, "/"))
}

// processCategories processes a list of categories with optimized concurrent image URL updates
func (s *CategoryService) processCategories(categories []model.CategoryWithProductCount) []model.CategoryWithProductCount {
	return s.updateImageURLsInChildren(categories)
}

// RefreshCategoryCache invalidates cache for a specific category or all categories
func (s *CategoryService) RefreshCategoryCache(ctx context.Context, categoryID string) error {
	if categoryID == "" || categoryID == "all" {
		// Invalidate all category caches
		s.cache.invalidate("all")
		s.logger.Info("invalidated all category caches")
		return nil
	}

	// Try to parse the UUID to verify it's valid
	id, err := uuid.Parse(categoryID)
	if err != nil {
		return fmt.Errorf("invalid category ID format: %w", err)
	}

	// Invalidate specific category caches
	s.invalidateCategoryCaches(id)
	s.logger.Info("invalidated cache for category", zap.String("categoryID", categoryID))
	return nil
}
