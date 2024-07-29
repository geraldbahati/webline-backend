package services

import (
	"context"
	"database/sql"
	"go.uber.org/zap"
	"log"
	"time"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepository  *repository.UserRepository
	tokenRepository *repository.TokenRepository
	config          *appconfig.Config
	logger          *zap.Logger
}

func NewUserService(
	userRepository *repository.UserRepository,
	tokenRepository *repository.TokenRepository,
	config *appconfig.Config,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		userRepository:  userRepository,
		tokenRepository: tokenRepository,
		config:          config,
		logger:          logger,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, registerUserParams model.RegisterUserParams) (*model.LoginResponse, error) {
	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerUserParams.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	// create user
	newUser := database.CreateUserParams{
		Email:          registerUserParams.Email,
		HashedPassword: string(hashedPassword),
	}

	createdUser, err := s.userRepository.CreateUser(ctx, newUser)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	// assign user role
	role, err := s.userRepository.GetRoleByName(ctx, "user")

	err = s.userRepository.AssignUserRole(ctx, createdUser.ID)
	if err != nil {
		s.logger.Error("Failed to assign user role", zap.Error(err))
		return nil, err
	}

	// login the user
	accessToken, refreshToken, expireTime, err := utils.GenerateTokens(createdUser.ID, createdUser.Email)
	if err != nil {
		s.logger.Error("Failed to generate tokens", zap.Error(err))
		return nil, err
	}

	err = s.tokenRepository.StoreRefreshToken(ctx, database.StoreRefreshTokenParams{
		UserID:    createdUser.ID,
		Token:     refreshToken,
		ExpiresAt: expireTime,
	})
	if err != nil {
		s.logger.Error("Failed to store refresh token", zap.Error(err))
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetUserByEmail returns the user with the given email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return database.User{}, err
	}

	return user, nil
}

// LoginUser logs in a user
func (s *UserService) LoginUser(ctx context.Context, params model.LoginParams) (model.LoginResponse, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, params.Email)
	if err != nil {
		return model.LoginResponse{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(params.Password))
	if err != nil {
		return model.LoginResponse{}, err
	}

	accessToken, refreshToken, expireTime, err := utils.GenerateTokens(user.ID, user.Email)
	if err != nil {
		return model.LoginResponse{}, err
	}

	err = s.tokenRepository.StoreRefreshToken(ctx, database.StoreRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: expireTime,
	})
	if err != nil {
		return model.LoginResponse{}, err
	}

	_, err = s.userRepository.UpdateUserLastLogin(ctx, user.ID)
	if err != nil {
		log.Printf("Failed to update last login for user with id %s: %v", user.ID.String(), err)
	}

	return model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken refreshes a user's access token
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (model.LoginResponse, error) {
	newAccessToken, err := utils.RefreshToken(refreshToken)
	if err != nil {
		return model.LoginResponse{}, err
	}

	return model.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken,
	}, nil
}

// UpdateUserProfile updates a user's profile
func (s *UserService) UpdateUserProfile(ctx context.Context, params model.UpdateUserProfileParams) (database.User, error) {
	userId := ctx.Value("userId").(uuid.UUID)
	user, err := s.userRepository.GetUserByID(ctx, userId)
	if err != nil {
		return database.User{}, err
	}

	if params.FirstName != "" {
		user.FirstName = sql.NullString{String: params.FirstName, Valid: params.FirstName != ""}
	}

	if params.LastName != "" {
		user.LastName = sql.NullString{String: params.LastName, Valid: params.LastName != ""}
	}

	if params.PhoneNumber != "" {
		user.PhoneNumber = sql.NullString{String: params.PhoneNumber, Valid: params.PhoneNumber != ""}
	}

	if params.ProfileImageUrl != "" {
		user.ProfileImageUrl = sql.NullString{String: params.ProfileImageUrl, Valid: params.ProfileImageUrl != ""}
	}

	if params.DateOfBirth != "" {
		dateOfBirth, err := time.Parse("02-01-2006", params.DateOfBirth)
		if err != nil {
			dateOfBirth = time.Time{}
		}
		user.DateOfBirth = sql.NullTime{Time: dateOfBirth, Valid: !dateOfBirth.IsZero()}
	}

	updatedUser, err := s.userRepository.UpdateUserProfile(ctx, database.UpdateUserProfileParams{
		ID:              user.ID,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		PhoneNumber:     user.PhoneNumber,
		ProfileImageUrl: user.ProfileImageUrl,
		DateOfBirth:     user.DateOfBirth,
	})
	if err != nil {
		return database.User{}, err
	}

	return updatedUser, nil
}

// SendPasswordResetEmail sends a password reset email to the user
func (s *UserService) SendPasswordResetEmail(ctx context.Context, email string) error {
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	return utils.SendResetPasswordEmail(user.ID, user.Email)
}

// UpdateUserPassword updates a user's password
func (s *UserService) UpdateUserPassword(ctx context.Context, token string, newPassword string) error {
	userId, err := utils.VerifyResetPasswordToken(token)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.userRepository.UpdateUserPassword(ctx, database.UpdateUserPasswordParams{
		ID:             userId,
		HashedPassword: string(hashedPassword),
	})
	if err != nil {
		return err
	}

	return nil
}

// VerifyResetPasswordToken verifies a reset password token
func (s *UserService) VerifyResetPasswordToken(token string) (uuid.UUID, error) {
	return utils.VerifyResetPasswordToken(token)
}

// DeactivateUser deactivates a user's account
func (s *UserService) DeactivateUser(ctx context.Context, userId string) error {
	userIdUUID, err := uuid.Parse(userId)
	if err != nil {
		return err
	}

	_, err = s.userRepository.DeactivateUser(ctx, userIdUUID)
	if err != nil {
		return err
	}

	return nil
}

// ListUsers lists all users
func (s *UserService) ListUsers(ctx context.Context, pageSize int32, page int32) (model.PaginationResult[[]database.User], error) {
	count, err := s.userRepository.CountAllUsers(ctx)
	if err != nil {
		return model.PaginationResult[[]database.User]{}, err
	}

	paginatedListUsers, err := utils.Paginate(
		s.config,
		count,
		page,
		pageSize,
		func(offset int32, limit int32) ([]database.User, error) {
			return s.userRepository.ListUsers(ctx, limit, offset)
		},
	)
	if err != nil {
		return model.PaginationResult[[]database.User]{}, err
	}

	return *paginatedListUsers, nil
}

// GetUserProfile gets a user's profile
func (s *UserService) GetUserProfile(ctx context.Context, userId string) (database.User, error) {
	userIdUUID, err := uuid.Parse(userId)
	if err != nil {
		return database.User{}, err
	}

	user, err := s.userRepository.GetUserByID(ctx, userIdUUID)
	if err != nil {
		return database.User{}, err
	}

	return user, nil
}
