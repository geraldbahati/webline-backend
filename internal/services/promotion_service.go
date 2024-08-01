package services

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
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
}

func NewPromotionService(logger *zap.Logger, config *appconfig.Config, s3Client *s3.Client, promotionRepo *repository.PromotionRepository, productRepo *repository.ProductRepository, productImageRepo *repository.ProductImageRepository, discountRepo *repository.DiscountRepository) *PromotionService {
	return &PromotionService{
		logger:           logger,
		config:           config,
		s3Client:         s3Client,
		promotionRepo:    promotionRepo,
		productRepo:      productRepo,
		productImageRepo: productImageRepo,
		discountRepo:     discountRepo,
	}
}

// CreatePromotion creates a new promotion
func (s *PromotionService) CreatePromotion(ctx context.Context, r *http.Request, tagline, mainTitle, subTitle, title, description string, discount float64, productID uuid.UUID) (*model.Promotion, error) {
	// check if the product exists
	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		s.logger.Error("failed to get product by ID", zap.Error(err))
		return nil, err
	}

	// upload image to S3
	filePath, err := utils.UploadFileToS3(ctx, r, s.s3Client, s.config.AWSBucketName, "promotions")
	if err != nil {
		s.logger.Error("failed to upload file to S3", zap.Error(err))
		return nil, err
	}

	// create promotion
	promotion, err := s.promotionRepo.CreatePromotion(ctx, &database.CreatePromotionParams{
		Title: title,
		Description: sql.NullString{
			String: description,
			Valid:  true,
		},
		ImageUrl: sql.NullString{
			String: filePath,
			Valid:  true,
		},
		StartDate: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		EndDate: sql.NullTime{
			Time:  time.Now().AddDate(0, 0, 7),
			Valid: true,
		},
		Tagline: sql.NullString{
			String: tagline,
			Valid:  true,
		},
		MainTitle: mainTitle,
		Subtitle:  subTitle,
	})
	if err != nil {
		s.logger.Error("failed to create promotion", zap.Error(err))
		return nil, err
	}

	// if discount is not 0, create discount
	var discountPercentage float64

	if discount != 0 {
		discount, err := s.discountRepo.CreateDiscount(ctx, &database.CreateDiscountParams{
			ProductID:          uuid.NullUUID{UUID: product.ID, Valid: true},
			DiscountPercentage: fmt.Sprintf("%f", discount),
			StartDate: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
			EndDate: sql.NullTime{
				Time:  time.Now().AddDate(0, 0, 7),
				Valid: true,
			},
		})
		if err != nil {
			s.logger.Error("failed to create discount", zap.Error(err))
			return nil, err
		}

		discountPercentage, err = strconv.ParseFloat(discount.DiscountPercentage, 64)
		if err != nil {
			s.logger.Error("failed to parse discount percentage", zap.Error(err))
			return nil, err
		}

	}

	// add product to promotion
	if err := s.promotionRepo.AddProductToPromotion(ctx, promotion.ID, productID); err != nil {
		s.logger.Error("failed to add product to promotion", zap.Error(err))
		return nil, err
	}

	return &model.Promotion{
		Tagline:           promotion.Tagline,
		MainTitle:         promotion.MainTitle,
		SubTitle:          promotion.SubTitle,
		Title:             promotion.Title,
		Description:       promotion.Description,
		Discount:          discountPercentage,
		PromotionImageUrl: s.constructS3URL(promotion.ImageUrl),
	}, nil

}

// constructS3URL constructs the S3 URL for a given file path
func (s *PromotionService) constructS3URL(filePath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.AWSBucketName, s.config.AWSRegion, filePath)
}

func (s *PromotionService) getProductDiscountPercentage(ctx context.Context, productID uuid.UUID) (float64, error) {
	discount, err := s.discountRepo.GetDiscountByProductID(ctx, &productID)
	if err != nil {
		s.logger.Error("failed to get product discount", zap.Error(err))
		return 0, fmt.Errorf("failed to get product discount: %w", err)
	}

	if discount != nil {
		discountPercentage, err := strconv.ParseFloat(discount.DiscountPercentage, 64)
		if err != nil {
			s.logger.Error("failed to parse discount percentage", zap.Error(err))
			return 0, fmt.Errorf("failed to parse discount percentage: %w", err)
		}
		return discountPercentage, nil
	}

	return 0, nil
}

// GetPromotions returns all promotions
func (s *PromotionService) GetPromotions(ctx context.Context) ([]*model.Promotion, error) {
	// get promotions
	promotions, err := s.promotionRepo.GetPromotionsWithProducts(ctx)
	if err != nil {
		s.logger.Error("failed to get promotions", zap.Error(err))
		return nil, err
	}

	for _, promotion := range promotions {
		// update the image urls
		promotion.PromotionImageUrl = s.constructS3URL(promotion.PromotionImageUrl)
		promotion.ProductImageUrl = s.constructS3URL(promotion.ProductImageUrl)
	}

	return promotions, nil
}
