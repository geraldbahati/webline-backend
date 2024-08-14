package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
	logger       *zap.Logger
	cfg          *appconfig.Config
	s3Client     *s3.Client
}

func NewCategoryService(categoryRepo *repository.CategoryRepository, logger *zap.Logger, cfg *appconfig.Config, s3Client *s3.Client) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,

		logger:   logger,
		cfg:      cfg,
		s3Client: s3Client,
	}
}

// CreateCategoryService creates a new category
func (s *CategoryService) CreateCategoryService(
	ctx context.Context,
	name string,
	parentID string,
	position int32,
) (database.Category, error) {

	// parse parentID to null uuid
	log.Printf("parentID: %v", parentID)
	var parentIDValue uuid.NullUUID
	if parentID != "" {
		id, err := uuid.Parse(parentID)
		if err != nil {
			s.logger.Error("failed to parse parent ID", zap.Error(err))
			return database.Category{}, err
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

	categoryParams := database.CreateCategoryParams{
		Name:     name,
		ParentID: parentIDValue,
		Position: position,
	}

	category, err := s.categoryRepo.CreateCategory(ctx, categoryParams)
	if err != nil {
		s.logger.Error("failed to create category", zap.Error(err))
		return database.Category{}, err
	}

	return category, nil
}

// GetCategoryByIDService retrieves a category by its ID
func (s *CategoryService) GetCategoryByIDService(ctx context.Context, id string) (database.Category, error) {
	// parse id to uuid
	categoryID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("failed to parse category ID", zap.Error(err))
		return database.Category{}, err
	}

	category, err := s.categoryRepo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to get category by ID", zap.Error(err))
		return database.Category{}, err
	}

	return category, nil
}

// GetCategoriesService retrieves all categories
func (s *CategoryService) GetCategoriesService(ctx context.Context) ([]database.Category, error) {
	categories, err := s.categoryRepo.GetCategories(ctx)
	if err != nil {
		s.logger.Error("failed to get categories", zap.Error(err))
		return nil, err
	}

	return categories, nil
}

// UpdateCategoryService updates a category in the database and returns the updated category
func (s *CategoryService) UpdateCategoryService(
	ctx context.Context,
	id string,
	name string,
	parentID string,
	position int32,
) (*database.Category, error) {
	// parse id to uuid
	categoryID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("failed to parse category ID", zap.Error(err))
		return nil, err
	}

	// get category by ID
	existingCategory, err := s.categoryRepo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		s.logger.Error("the category does not exist", zap.Error(err))
		return nil, err
	}

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
		parentIDValue = existingCategory.ParentID
	}

	// if the name is the same as the current name, return the category
	if name == existingCategory.Name && parentIDValue == existingCategory.ParentID && position == existingCategory.Position {
		return nil, nil
	}

	// if name is not provided, use the current name
	if name == "" {
		name = existingCategory.Name
	}

	// if position is not provided, use the current position
	if position == 0 {
		position = existingCategory.Position
	}

	// update category
	category, err := s.categoryRepo.UpdateCategory(ctx, categoryID, name, parentIDValue, position)
	if err != nil {
		s.logger.Error("failed to update category", zap.Error(err))
		return nil, err
	}

	return &category, nil
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

//
//// GetCategoryHierarchyService retrieves the category hierarchy
//func (s *CategoryService) GetCategoryHierarchyService(ctx context.Context) ([]*Category, error) {
//	// get category tree
//	categories, err := s.categoryRepo.GetCategoryHierarchy(ctx)
//	if err != nil {
//		s.logger.Error("failed to get category tree", zap.Error(err))
//		return nil, err
//	}
//
//	// build the category hierarchy
//	categoryHierarchy := s.buildCategoryHierarchy(categories)
//
//	return categoryHierarchy, nil
//
//}

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

//func (s *CategoryService) buildCategoryHierarchy(categories []database.GetCategoryHierarchyRow) []*Category {
//	categoryMap := make(map[uuid.UUID]*Category)
//	childParentMap := make(map[uuid.UUID]uuid.UUID)
//
//	for _, row := range categories {
//		if _, exists := categoryMap[row.CategoryID]; !exists {
//			categoryMap[row.CategoryID] = &Category{
//				CategoryID:       row.CategoryID.String(),
//				CategoryName:     row.CategoryName,
//				FeaturedProducts: []Product{},
//				Processors:       []string{},
//				Size:             []string{},
//				Storage:          []string{},
//				Children:         []*Category{},
//				Position:         int(row.Position),
//				ParentID:         row.ParentID.UUID,
//			}
//		}
//
//		if row.ParentID.UUID != uuid.Nil {
//			childParentMap[row.CategoryID] = row.ParentID.UUID
//		}
//
//		if row.ProductID.Valid {
//			category := categoryMap[row.CategoryID]
//
//			product := Product{
//				ProductID:   row.ProductID.UUID.String(),
//				ProductName: row.ProductName.String,
//				ImageURL:    s.constructS3URL(row.ImageUrl.String),
//			}
//
//			// Find the immediate parent category to place the product in the second layer
//			var parentCategory *Category
//			if parentID, exists := childParentMap[row.CategoryID]; exists {
//				if _, exists := childParentMap[parentID]; exists {
//					parentCategory = categoryMap[parentID]
//				}
//			}
//
//			if parentCategory != nil {
//				parentCategory.FeaturedProducts = append(parentCategory.FeaturedProducts, product)
//				if row.Processor.Valid {
//					parentCategory.Processors = append(parentCategory.Processors, row.Processor.String)
//				}
//				if row.Size.Valid {
//					parentCategory.Size = append(parentCategory.Size, row.Size.String)
//				}
//				if row.Storage.Valid {
//					parentCategory.Storage = append(parentCategory.Storage, row.Storage.String)
//				}
//			} else {
//				category.FeaturedProducts = append(category.FeaturedProducts, product)
//				if row.Processor.Valid {
//					category.Processors = append(category.Processors, row.Processor.String)
//				}
//				if row.Size.Valid {
//					category.Size = append(category.Size, row.Size.String)
//				}
//				if row.Storage.Valid {
//					category.Storage = append(category.Storage, row.Storage.String)
//				}
//			}
//		}
//	}
//
//	// Remove duplicate processors, sizes, and storages
//	for _, category := range categoryMap {
//		category.Processors = uniqueStrings(category.Processors)
//		category.Size = uniqueStrings(category.Size)
//		category.Storage = uniqueStrings(category.Storage)
//	}
//
//	// Organize the hierarchy
//	rootCategories := make([]*Category, 0)
//	for id, category := range categoryMap {
//		if parentID, exists := childParentMap[id]; exists {
//			if parentCategory, ok := categoryMap[parentID]; ok {
//				parentCategory.Children = append(parentCategory.Children, category)
//			}
//		} else {
//			rootCategories = append(rootCategories, category)
//		}
//	}
//
//	// Sort root categories by position
//	sort.SliceStable(rootCategories, func(i, j int) bool {
//		return rootCategories[i].Position < rootCategories[j].Position
//	})
//
//	// Sort children categories by position
//	for _, category := range categoryMap {
//		sort.SliceStable(category.Children, func(i, j int) bool {
//			return category.Children[i].Position < category.Children[j].Position
//		})
//	}
//
//	return rootCategories
//}

// constructS3URL constructs the S3 URL for a given file path
func (s *CategoryService) constructS3URL(filePath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.cfg.AWSBucketName, s.cfg.AWSRegion, filePath)
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
	filePath, err := utils.UploadFileToS3(ctx, r, s.s3Client, s.cfg.AWSBucketName, "category_images")
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
