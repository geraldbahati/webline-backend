package services

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
	"weblineBackend/internal/app_errors"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"
)

type AdminRequestService struct {
	adminRequestRepo repository.AdminRequestRepository
	userRepo         *repository.UserRepository
	logger           *zap.Logger
	config           *appconfig.Config
}

func NewAdminRequestService(adminRequestRepo repository.AdminRequestRepository, userRepo *repository.UserRepository, logger *zap.Logger,
	config *appconfig.Config) *AdminRequestService {
	return &AdminRequestService{
		adminRequestRepo: adminRequestRepo,
		userRepo:         userRepo,
		logger:           logger,
		config:           config,
	}
}

// RequestAdminRole requests an admin role
func (s *AdminRequestService) RequestAdminRole(ctx context.Context, userID uuid.UUID, reason string) error {
	// get if the user is an admin
	isAdmin, err := s.userRepo.IsAdmin(ctx, userID)
	if err != nil {
		return err
	}

	if isAdmin {
		return app_errors.NewAlreadyAdminError()
	}

	// create a new admin request
	questID, err := s.adminRequestRepo.CreateAdminRequest(ctx, userID, reason)
	if err != nil {
		s.logger.Error("failed to create admin request", zap.Error(err))
		return err
	}

	// generate an approval token
	token, expiresAt, err := utils.GenerateApprovalToken(questID, 1*time.Hour)
	if err != nil {
		s.logger.Error("failed to generate approval token", zap.Error(err))
		return err
	}

	// store the approval token
	err = s.adminRequestRepo.StoreApprovalToken(ctx, token, questID, expiresAt)
	if err != nil {
		s.logger.Error("failed to store approval token", zap.Error(err))
		return err
	}

	// get user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user", zap.Error(err))
		return err
	}

	// send an email to the admin
	if err := utils.SendAdminRequestEmail(s.config, user.Email, token); err != nil {
		s.logger.Error("failed to send admin request email", zap.Error(err))
		return err
	}

	s.logger.Info("admin request created", zap.String("email", token))
	return nil
}

// ApproveAdminRequest approves an admin request
func (s *AdminRequestService) ApproveAdminRequest(ctx context.Context, token string) error {
	// get user id from context
	userID, ok := ctx.Value("userId").(uuid.UUID)
	if !ok {
		s.logger.Error("failed to get user id from context")
		return app_errors.NewUnauthorizedUserError()
	}

	// check if the user is an admin
	isAdmin, err := s.userRepo.IsAdmin(ctx, userID)
	if err != nil {
		return err
	}

	if !isAdmin {
		s.logger.Error("user is not an admin")
		return app_errors.NewUnauthorizedUserError()
	}

	// parse the approval token
	claims, err := utils.ParseApprovalToken(token)
	if err != nil {
		s.logger.Error("failed to parse approval token", zap.Error(err))
		return app_errors.NewInvalidTokenError()
	}

	// get the admin request
	adminRequest, err := s.adminRequestRepo.GetAdminRequestByID(ctx, claims.RequestID)
	if err != nil {
		s.logger.Error("failed to get admin request", zap.Error(err))
		return err
	}

	// approve the admin request
	err = s.adminRequestRepo.ApproveAdminRequest(ctx, adminRequest.ID)
	if err != nil {
		s.logger.Error("failed to approve admin request", zap.Error(err))
		return err
	}

	// delete the approval token
	err = s.adminRequestRepo.DeleteApprovalToken(ctx, token)
	if err != nil {
		s.logger.Error("failed to delete approval token", zap.Error(err))
		return err
	}

	// make the user an admin
	err = s.userRepo.MakeAdmin(ctx, adminRequest.UserID)
	if err != nil {
		s.logger.Error("failed to make user an admin", zap.Error(err))
		return err
	}

	// send an email to the user
	if err := utils.SendAdminRequestApprovedEmail(s.config, adminRequest.Email); err != nil {
		s.logger.Error("failed to send admin request approved email", zap.Error(err))
		return err
	}

	s.logger.Info("admin request approved", zap.String("email", adminRequest.Email))
	return nil
}
