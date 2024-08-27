package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"weblineBackend/internal/app_errors"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
	userRepo     *repository.UserRepository
	logger       *zap.Logger
	config       *appconfig.Config
	s3Client     *s3.Client
}

func NewCategoryService(categoryRepo *repository.CategoryRepository, userRepo *repository.UserRepository, logger *zap.Logger, cfg *appconfig.Config, s3Client *s3.Client) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
		userRepo:     userRepo,
		logger:       logger,
		config:       cfg,
		s3Client:     s3Client,
	}
}

// CreateCategoryService creates a new category
func (s *CategoryService) CreateCategoryService(
	ctx context.Context,
	params *model.CreateCategoryParams,
	image *model.ImageFile,
) error {

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

	if categoryID != nil {
		return s.updateExistingCategory(ctx, params, image, categoryID)
	}

	return s.createNewCategory(ctx, params, image)
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

	filePath, err := utils.UploadCustomFileToS3(ctx, image.File, image.FileHeader, s.s3Client, s.config.AWSBucketName, "promotions")
	if err != nil {
		return "", err
	}

	if existingImageUrl != "" {
		if err := utils.DeleteFileFromS3(ctx, s.s3Client, s.config.AWSBucketName, existingImageUrl); err != nil {
			return "", err
		}
	}

	return filePath, nil
}

func (s *CategoryService) getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("userId").(uuid.UUID)
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
	categories, err := s.categoryRepo.GetCategories(ctx)
	if err != nil {
		s.logger.Error("failed to get categories", zap.Error(err))
		return nil, err
	}

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

func uniqueStrings(input []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range input {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// constructS3URL constructs the S3 URL for a given file path
func (s *CategoryService) constructS3URL(filePath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.AWSBucketName, s.config.AWSRegion, filePath)
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
	hierarchy, err := s.categoryRepo.GetV2CategoryHierarchy(ctx)
	if err != nil {
		s.logger.Error("failed to get V2 category hierarchy", zap.Error(err))
		return nil, fmt.Errorf("failed to get V2 category hierarchy: %w", err)
	}

	return hierarchy, nil
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
	category, err := s.categoryRepo.GetCategoryDetailsBySlug(ctx, slug)
	if err != nil {
		s.logger.Error("failed to get category by slug", zap.Error(err))
		return nil, err
	}

	// update the image url to s3 url
	if category.ImageURL != "" {
		category.ImageURL = s.constructS3URL(category.ImageURL)
	}

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
