package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sync"
	"weblineBackend/internal/app_errors"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"
)

type PromotionService struct {
	logger           *zap.Logger
	config           *appconfig.Config
	s3Client         *s3.Client
	promotionRepo    *repository.PromotionRepository
	productRepo      *repository.ProductRepository
	productImageRepo *repository.ProductImageRepository
	discountRepo     *repository.DiscountRepository
	userRepo         *repository.UserRepository
}

func NewPromotionService(
	logger *zap.Logger,
	config *appconfig.Config,
	s3Client *s3.Client,
	promotionRepo *repository.PromotionRepository,
	productRepo *repository.ProductRepository,
	productImageRepo *repository.ProductImageRepository,
	discountRepo *repository.DiscountRepository,
	userRepo *repository.UserRepository) *PromotionService {
	return &PromotionService{
		logger:           logger,
		config:           config,
		s3Client:         s3Client,
		promotionRepo:    promotionRepo,
		productRepo:      productRepo,
		productImageRepo: productImageRepo,
		discountRepo:     discountRepo,
		userRepo:         userRepo,
	}
}

// constructS3URL constructs the S3 URL for a given file path
func (s *PromotionService) constructS3URL(filePath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.AWSBucketName, s.config.AWSRegion, filePath)
}

// GetPromotions returns all promotions
func (s *PromotionService) GetPromotions(ctx context.Context) ([]*model.Promotion, error) {
	// get promotions
	promotions, err := s.promotionRepo.GetPromotions(ctx)
	if err != nil {
		s.logger.Error("failed to get promotions", zap.Error(err))
		return nil, err
	}

	for _, promotion := range promotions {
		// update the image urls
		if promotion.ImageUrl != "" {
			promotion.ImageUrl = s.constructS3URL(promotion.ImageUrl)
		}
	}

	return promotions, nil
}

// GetV2Promotions returns all promotions for dashboard
func (s *PromotionService) GetV2Promotions(ctx context.Context) ([]*model.V2Promotion, error) {
	// get promotions
	promotions, err := s.promotionRepo.GetV2Promotions(ctx)
	if err != nil {
		s.logger.Error("failed to get promotions", zap.Error(err))
		return nil, err
	}

	for _, promotion := range promotions {
		// update the image urls
		if promotion.ImageUrl != "" {
			promotion.ImageUrl = s.constructS3URL(promotion.ImageUrl)
		}
	}

	return promotions, nil
}

// CreateOrEditV2Promotion creates or edits a promotion
func (s *PromotionService) CreateOrEditV2Promotion(ctx context.Context, params *model.CreatePromotionParams, image *model.ImageFile) error {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	if err := s.verifyAdminStatus(ctx, userID); err != nil {
		return err
	}

	if params.Slug == "" {
		return s.createPromotion(ctx, params, image)
	}

	existingPromotion, err := s.promotionRepo.GetPromotionBySlug(ctx, params.Slug)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return s.logAndReturnError("failed to check if promotion exists", err)
	}

	if existingPromotion != nil {
		return s.updatePromotion(ctx, existingPromotion, params, image)
	}

	return s.createPromotion(ctx, params, image)
}

func (s *PromotionService) createPromotion(ctx context.Context, params *model.CreatePromotionParams, image *model.ImageFile) error {
	filePath, err := s.handlePromotionImage(ctx, image, "")
	if err != nil {
		return s.logAndReturnError("failed to upload promotion image", err)
	}

	promotion, err := s.promotionRepo.CreatePromotion(ctx, &database.CreatePromotionParams{
		Title:       params.Name,
		Description: optionalString(params.Description),
		ImageUrl:    optionalString(filePath),
		StartDate:   params.StartDate,
		EndDate:     params.EndDate,
	})
	if err != nil {
		return s.logAndReturnError("failed to create promotion", err)
	}

	return s.addProductsToPromotion(ctx, promotion.ID, params.ProductSlugs)
}

func (s *PromotionService) updatePromotion(ctx context.Context, promotion *model.PromotionSchema, params *model.CreatePromotionParams, image *model.ImageFile) error {
	filePath, err := s.handlePromotionImage(ctx, image, promotion.ImageUrl)
	if err != nil {
		return s.logAndReturnError("failed to upload promotion image", err)
	}

	updateParams := &database.UpdatePromotionParams{
		ID:          promotion.ID,
		Title:       params.Name,
		Description: optionalString(params.Description),
		StartDate:   params.StartDate,
		EndDate:     params.EndDate,
		ImageUrl:    optionalString(filePath),
	}

	if err := s.promotionRepo.UpdatePromotion(ctx, updateParams); err != nil {
		return s.logAndReturnError("failed to update promotion", err)
	}

	return s.addProductsToPromotion(ctx, promotion.ID, params.ProductSlugs)
}

func (s *PromotionService) handlePromotionImage(ctx context.Context, image *model.ImageFile, existingImageUrl string) (string, error) {
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

func (s *PromotionService) addProductsToPromotion(ctx context.Context, promotionID uuid.UUID, productSlugs []string) error {
	const batchSize = 100
	const maxConcurrency = 10
	var (
		wg              sync.WaitGroup
		sem             = make(chan struct{}, maxConcurrency)
		mu              sync.Mutex
		errList         []error
		addedProductIDs []uuid.UUID
	)

	// Fetch the existing product IDs in the promotion
	existingProductIDs, err := s.promotionRepo.GetProductIDsByPromotionID(ctx, promotionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return s.logAndReturnError("failed to fetch existing product IDs in promotion", err)
	}

	// Process product slugs in batches to add them to the promotion
	for i := 0; i < len(productSlugs); i += batchSize {
		end := i + batchSize
		if end > len(productSlugs) {
			end = len(productSlugs)
		}
		batch := productSlugs[i:end]

		wg.Add(1)
		go func(batch []string) {
			defer wg.Done()

			// Acquire a semaphore slot
			sem <- struct{}{}
			defer func() { <-sem }() // Release the semaphore slot

			// Fetch product IDs for the current batch
			productIDs, err := s.productRepo.GetProductIDsBySlugs(ctx, batch)
			if err != nil {
				mu.Lock()
				errList = append(errList, fmt.Errorf("failed to fetch product IDs by slugs for batch %v: %w", batch, err))
				mu.Unlock()
				return
			}

			// Batch add products to promotion
			if err := s.promotionRepo.AddProductsToPromotion(ctx, promotionID, productIDs); err != nil {
				mu.Lock()
				errList = append(errList, fmt.Errorf("failed to add products to promotion for batch %v: %w", batch, err))
				mu.Unlock()
				return
			}

			// Collect added product IDs for later removal check
			mu.Lock()
			addedProductIDs = append(addedProductIDs, productIDs...)
			mu.Unlock()

		}(batch)
	}

	wg.Wait()

	// After all additions, handle removals
	if len(errList) == 0 && len(existingProductIDs) > 0 {
		toRemove := s.getProductsToRemove(existingProductIDs, addedProductIDs)

		if len(toRemove) > 0 {
			if err := s.promotionRepo.RemoveProductsFromPromotion(ctx, promotionID, toRemove); err != nil {
				s.logger.Error("failed to remove products from promotion", zap.Error(err))
				errList = append(errList, fmt.Errorf("failed to remove products from promotion: %w", err))
			}
		}
	}

	if len(errList) > 0 {
		for _, err := range errList {
			s.logger.Error("error during promotion modification", zap.Error(err))
		}
		return fmt.Errorf("one or more errors occurred during promotion modification")
	}

	return nil
}

// getProductsToRemove determines which products need to be removed from the promotion
func (s *PromotionService) getProductsToRemove(existingProductIDs, addedProductIDs []uuid.UUID) []uuid.UUID {
	addedSet := make(map[uuid.UUID]struct{}, len(addedProductIDs))
	for _, id := range addedProductIDs {
		addedSet[id] = struct{}{}
	}

	var toRemove []uuid.UUID
	for _, id := range existingProductIDs {
		if _, found := addedSet[id]; !found {
			toRemove = append(toRemove, id)
		}
	}

	return toRemove
}

func (s *PromotionService) getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("userId").(uuid.UUID)
	if !ok {
		return uuid.Nil, app_errors.NewUnauthorizedUserError()
	}
	return userID, nil
}

func (s *PromotionService) logAndReturnError(message string, err error) error {
	s.logger.Error(message, zap.Error(err))
	return fmt.Errorf("%s: %w", message, err)
}

func (s *PromotionService) verifyAdminStatus(ctx context.Context, userID uuid.UUID) error {
	isAdmin, err := s.userRepo.IsAdmin(ctx, userID)
	if err != nil || !isAdmin {
		return app_errors.NewUnauthorizedUserError()
	}
	return nil
}

func optionalString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

// GetPromotionDetails retrieves the details of a promotion
func (s *PromotionService) GetPromotionDetails(ctx context.Context, slug string) (*model.PromotionDetails, error) {
	// check if is admin
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		s.logger.Error("failed to get user ID from context", zap.Error(err))
		return nil, err
	}

	if err := s.verifyAdminStatus(ctx, userID); err != nil {
		s.logger.Error("failed to verify admin status", zap.Error(err))
		return nil, err
	}

	// get the promotion details
	promotion, err := s.promotionRepo.GetPromotionDetails(ctx, slug)
	if err != nil {
		s.logger.Error("failed to get promotion details", zap.Error(err))
		return nil, err
	}

	// update the image url
	if promotion.ImageUrl != "" {
		promotion.ImageUrl = s.constructS3URL(promotion.ImageUrl)
	}

	// update the images for the products
	for _, product := range promotion.Products {
		if product.ImageURL != "" {
			product.ImageURL = s.constructS3URL(product.ImageURL)
		}
	}

	return promotion, nil
}
