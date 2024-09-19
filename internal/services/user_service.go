package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"weblineBackend/internal/app_errors"

	"log"
	"strings"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepository              *repository.UserRepository
	roleRepository              *repository.RoleRepository
	userRoleRepository          *repository.UserRoleRepository
	tokenRepository             *repository.TokenRepository
	verificationTokenRepository repository.VerificationTokenRepository
	passwordResetRepository     repository.PasswordResetRepository
	adminRequestRepository      repository.AdminRequestRepository
	guestCheckoutRepository     *repository.GuestCheckoutRepository
	config                      *appconfig.Config
	logger                      *zap.Logger
	s3Client                    *s3.Client
}

func NewUserService(
	userRepository *repository.UserRepository,
	roleRepository *repository.RoleRepository,
	userRoleRepository *repository.UserRoleRepository,
	verificationTokenRepository repository.VerificationTokenRepository,
	passwordResetRepository repository.PasswordResetRepository,
	adminRequestRepository repository.AdminRequestRepository,
	tokenRepository *repository.TokenRepository,
	guestCheckoutRepository *repository.GuestCheckoutRepository,
	config *appconfig.Config,
	logger *zap.Logger,
	s3Client *s3.Client,
) *UserService {
	return &UserService{
		userRepository:              userRepository,
		roleRepository:              roleRepository,
		userRoleRepository:          userRoleRepository,
		verificationTokenRepository: verificationTokenRepository,
		passwordResetRepository:     passwordResetRepository,
		adminRequestRepository:      adminRequestRepository,
		tokenRepository:             tokenRepository,
		guestCheckoutRepository:     guestCheckoutRepository,
		config:                      config,
		logger:                      logger,
		s3Client:                    s3Client,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, registerUserParams model.RegisterUserParams) error {
	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerUserParams.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return app_errors.NewInternalError("Failed to process password", err)
	}

	// create user
	newUser := database.CreateUserParams{
		Email:          registerUserParams.Email,
		HashedPassword: sql.NullString{String: string(hashedPassword), Valid: true},
	}

	createdUser, err := s.userRepository.CreateUser(ctx, newUser)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return app_errors.NewInternalError("Failed to create user", err)
	}

	// get or create customer role
	role, err := s.getOrCreateCustomerRole(ctx)
	if err != nil {
		s.logger.Error("Failed to get or create customer role", zap.Error(err))
		return err
	}

	// Assign role to user
	err = s.userRoleRepository.AssignRoleToUser(ctx, createdUser.ID, role.ID)
	if err != nil {
		s.logger.Error("Failed to assign role to user", zap.Error(err))
		return app_errors.NewInternalError("Failed to assign role to user", err)
	}

	return nil
}

// CreateUserFromOrder creates a new user from an order
func (s *UserService) CreateUserFromOrder(ctx context.Context, userParams model.CreateUserParams) (uuid.UUID, error) {
	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userParams.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return uuid.UUID{}, app_errors.NewInternalError("Failed to process password", err)
	}

	// create user
	newUser := database.CreateUserParams{
		Email:          userParams.Email,
		HashedPassword: sql.NullString{String: string(hashedPassword), Valid: true},
	}

	createdUser, err := s.userRepository.CreateUser(ctx, newUser)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return uuid.UUID{}, app_errors.NewInternalError("Failed to create user", err)
	}

	// get or create customer role
	role, err := s.getOrCreateCustomerRole(ctx)
	if err != nil {
		s.logger.Error("Failed to get or create customer role", zap.Error(err))
		return uuid.UUID{}, err
	}

	// Assign role to user
	err = s.userRoleRepository.AssignRoleToUser(ctx, createdUser.ID, role.ID)
	if err != nil {
		s.logger.Error("Failed to assign role to user", zap.Error(err))
		return uuid.UUID{}, app_errors.NewInternalError("Failed to assign role to user", err)
	}

	// update user details
	err = s.userRepository.UpdateUser(ctx, model.UpdateUserParams{
		ID:          createdUser.ID,
		FirstName:   userParams.FirstName,
		LastName:    userParams.LastName,
		PhoneNumber: userParams.PhoneNumber,
	})
	if err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return uuid.UUID{}, app_errors.NewInternalError("Failed to update user", err)
	}

	return createdUser.ID, nil
}

// CreateGuestUser creates a new guest user
func (s *UserService) CreateGuestUser(ctx context.Context, userParams model.CreateGuestUserParams) (*uuid.UUID, error) {
	existingGuest, err := s.guestCheckoutRepository.GetGuestCheckoutByEmail(ctx, userParams.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("failed to check if guest exists", zap.Error(err))
		return nil, fmt.Errorf("failed to check if guest exists: %w", err)
	}

	if existingGuest != nil {
		return &existingGuest.ID, nil
	}

	guestParams := &database.CreateGuestCheckoutParams{
		Email:         userParams.Email,
		FirstName:     userParams.FirstName,
		LastName:      userParams.LastName,
		Phone:         sql.NullString{String: userParams.Phone, Valid: true},
		StreetAddress: "",
		City:          userParams.City,
		State:         userParams.County,
		Country:       userParams.Country,
	}
	newGuestID, err := s.guestCheckoutRepository.CreateGuestCheckout(ctx, guestParams)
	if err != nil {
		s.logger.Error("failed to create guest checkout", zap.Error(err))
		return nil, fmt.Errorf("failed to create guest checkout: %w", err)
	}
	return newGuestID, nil

}

// getOrCreateCustomerRole gets the customer role or creates it if it doesn't exist
func (s *UserService) getOrCreateCustomerRole(ctx context.Context) (*database.GetRoleByNameRow, error) {
	role, err := s.roleRepository.GetRoleByName(ctx, "customer")
	if err != nil {
		if err == sql.ErrNoRows {
			// Role doesn't exist, create it
			createdRole, createErr := s.roleRepository.CreateRole(ctx, "customer", "Default customer role")
			if createErr != nil {
				s.logger.Error("Failed to create customer role", zap.Error(createErr))
				return nil, app_errors.NewInternalError("Failed to create customer role", createErr)
			}
			return &database.GetRoleByNameRow{
				ID:       createdRole.ID,
				RoleName: createdRole.Name,
			}, nil
		}
		s.logger.Error("Failed to get customer role", zap.Error(err))
		return nil, app_errors.NewInternalError("Failed to get customer role", err)
	}
	return role, nil
}

// GetUserByEmail returns the user with the given email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// LoginUser logs in a user
func (s *UserService) LoginUser(ctx context.Context, params model.LoginParams) (model.LoginResponse, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, params.Email)
	if err != nil {
		return model.LoginResponse{}, err
	}

	if user.Password == "" {
		return model.LoginResponse{}, errors.New("use google to login")
	}

	if user.EmailVerified == nil {
		emailNotVerified := app_errors.NewEmailNotVerifiedError()

		return model.LoginResponse{}, emailNotVerified
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(params.Password))
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

	roles, err := s.userRoleRepository.GetRolesForUser(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to get roles for user", zap.Error(err))
		return model.LoginResponse{}, err

	}

	return model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: model.User{
			ID:    user.ID,
			Email: user.Email,
			Roles: roles,
		},
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

// SendPasswordResetEmail sends a password reset email to the user
func (s *UserService) SendPasswordResetEmail(ctx context.Context, email string) error {
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return app_errors.NewUserNotFoundError()
		}
		return err
	}

	// Delete existing reset password tokens
	err = s.passwordResetRepository.DeletePasswordResetToken(ctx, user.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("Failed to delete password reset tokens", zap.Error(err))
		return err
	}

	// Generate password reset token
	passwordResetToken, expiresAt, err := utils.GeneratePasswordResetToken(user.Email, 1*time.Hour)
	if err != nil {
		s.logger.Error("Failed to generate password reset token", zap.Error(err))
		return err
	}

	// Store password reset token
	err = s.passwordResetRepository.StorePasswordResetToken(ctx, user.Email, passwordResetToken, expiresAt)
	if err != nil {
		s.logger.Error("Failed to store password reset token", zap.Error(err))
		return err
	}

	// Send password reset email
	err = utils.SendPasswordResetEmail(s.config, user.Email, passwordResetToken)
	if err != nil {
		s.logger.Error("Failed to send password reset email", zap.Error(err))
		return err
	}

	return nil
}

// UpdateUserPassword updates a user's password
func (s *UserService) UpdateUserPassword(ctx context.Context, token string, newPassword string) error {
	claims, err := utils.ParsePasswordResetToken(token)
	if err != nil {
		return err
	}

	// check if the token has expired
	if expired, _ := utils.IsPasswordResetTokenExpired(token); expired {
		return app_errors.NewUserNotFoundError()
	}

	// get user by email
	user, err := s.userRepository.GetUserByEmail(ctx, claims.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return app_errors.NewUserNotFoundError()
		}
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.userRepository.UpdateUserPassword(ctx, database.UpdateUserPasswordParams{
		ID: user.ID,
		HashedPassword: sql.NullString{
			String: string(hashedPassword),
			Valid:  true,
		},
	})
	if err != nil {
		return err
	}

	// Delete password reset token
	err = s.passwordResetRepository.DeletePasswordResetToken(ctx, user.Email)
	if err != nil {
		return err
	}

	return nil
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
func (s *UserService) GetUserProfile(ctx context.Context, userId string) (*model.User, error) {
	userIdUUID, err := uuid.Parse(userId)
	if err != nil {
		s.logger.Error("Failed to parse user id", zap.Error(err))
		return nil, err
	}

	var user *model.User
	user, err = s.userRepository.GetUserByID(ctx, userIdUUID)
	if errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("User not found", zap.Error(err))
		user, err = s.userRepository.GetUserByProvider(ctx, "google", userId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}

			s.logger.Error("Failed to get user by provider", zap.Error(err))
			return nil, err
		}

		s.logger.Info("User found by provider")
	}
	if err != nil {
		s.logger.Error("Failed to get user by id", zap.Error(err))
		return nil, err
	}

	roles, err := s.userRoleRepository.GetRolesForUser(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to get roles for user", zap.Error(err))
		return nil, err
	}

	user.Roles = roles

	return user, nil
}

// LoginWithGoogle logs in a user using Google
func (s *UserService) LoginWithGoogle(ctx context.Context, googleUser model.GoogleUser) (model.LoginResponse, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, googleUser.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			// create user
			firstName := strings.Split(*googleUser.Name, " ")[0]
			lastName := strings.Split(*googleUser.Name, " ")[1]

			newUser := database.CreateUserParams{
				Email: googleUser.Email,
				ProfileImageUrl: sql.NullString{
					String: *googleUser.Picture,
					Valid:  googleUser.Picture != nil,
				},
				FirstName: sql.NullString{
					String: firstName,
					Valid:  true,
				},
				LastName: sql.NullString{
					String: lastName,
					Valid:  true,
				},

				Provider: sql.NullString{
					String: "google",
					Valid:  true,
				},
				ProviderID: sql.NullString{
					String: googleUser.ID,
					Valid:  true,
				},
			}

			createdUser, err := s.userRepository.CreateUser(ctx, newUser)
			if err != nil {
				s.logger.Error("Failed to create user", zap.Error(err))
				return model.LoginResponse{}, err
			}

			// assign user role
			role, err := s.roleRepository.GetRoleByName(ctx, "customer")
			if err != nil {
				s.logger.Error("Failed to get role", zap.Error(err))
				return model.LoginResponse{}, err
			}

			// Assign role to user
			err = s.userRoleRepository.AssignRoleToUser(ctx, createdUser.ID, role.ID)
			if err != nil {
				s.logger.Error("Failed to assign role to user", zap.Error(err))
				return model.LoginResponse{}, err
			}

			// login the user
			accessToken, refreshToken, expireTime, err := utils.GenerateTokens(createdUser.ID, createdUser.Email)
			if err != nil {
				s.logger.Error("Failed to generate tokens", zap.Error(err))
				return model.LoginResponse{}, err
			}

			err = s.tokenRepository.StoreRefreshToken(ctx, database.StoreRefreshTokenParams{
				UserID:    createdUser.ID,
				Token:     refreshToken,
				ExpiresAt: expireTime,
			})
			if err != nil {
				s.logger.Error("Failed to store refresh token", zap.Error(err))
				return model.LoginResponse{}, err
			}

			roles, err := s.userRoleRepository.GetRolesForUser(ctx, createdUser.ID)
			if err != nil {
				s.logger.Error("Failed to get roles for user", zap.Error(err))
				return model.LoginResponse{}, err

			}

			createdUser.Roles = roles

			return model.LoginResponse{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				User:         *createdUser,
			}, nil
		}
		return model.LoginResponse{}, err
	}

	// update user with google details
	firstName := strings.Split(*googleUser.Name, " ")[0]
	lastName := strings.Split(*googleUser.Name, " ")[1]

	params := database.UpdateUserProfileParams{
		ID: user.ID,
		FirstName: sql.NullString{
			String: firstName,
			Valid:  true,
		},

		LastName: sql.NullString{
			String: lastName,
			Valid:  true,
		},

		ProfileImageUrl: sql.NullString{
			String: *googleUser.Picture,
			Valid:  googleUser.Picture != nil,
		},
		Provider: sql.NullString{
			String: "google",
			Valid:  true,
		},
		ProviderID: sql.NullString{
			String: googleUser.ID,
			Valid:  true,
		},
	}

	updatedUser, err := s.userRepository.UpdateUserProfile(ctx, params)
	if err != nil {
		s.logger.Error("Failed to update user profile", zap.Error(err))
		return model.LoginResponse{}, err
	}

	accessToken, refreshToken, expireTime, err := utils.GenerateTokens(user.ID, user.Email)
	if err != nil {
		s.logger.Error("Failed to generate tokens", zap.Error(err))
		return model.LoginResponse{}, err
	}

	err = s.tokenRepository.StoreRefreshToken(ctx, database.StoreRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: expireTime,
	})
	if err != nil {
		s.logger.Error("Failed to store refresh token", zap.Error(err))
		return model.LoginResponse{}, err
	}

	roles, err := s.userRoleRepository.GetRolesForUser(ctx, updatedUser.ID)
	if err != nil {
		s.logger.Error("Failed to get roles for user", zap.Error(err))
		return model.LoginResponse{}, err
	}

	updatedUser.Roles = roles

	return model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *updatedUser,
	}, nil

}

// EmailVerified checks if a user's email is verified
func (s *UserService) EmailVerified(ctx context.Context, email string) error {
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user.EmailVerified == nil {
		// Delete existing verification tokens
		err = s.verificationTokenRepository.DeleteVerificationTokens(ctx, user.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("Failed to delete verification tokens", zap.Error(err))
			return err
		}

		// Generate email verification token
		verificationToken, expiresAt, err := utils.GenerateVerificationToken(user.Email)
		if err != nil {
			s.logger.Error("Failed to generate email verification token", zap.Error(err))
			return err
		}

		_, err = s.verificationTokenRepository.CreateVerificationToken(ctx, user.Email, verificationToken, expiresAt)
		if err != nil {
			s.logger.Error("Failed to create verification token", zap.Error(err))
			return err
		}

		// send verification email
		err = utils.SendVerificationEmail(s.config, user.Email, verificationToken)
		if err != nil {
			s.logger.Error("Failed to send verification email", zap.Error(err))
			return err
		}

		emailNotVerified := app_errors.NewEmailNotVerifiedError()

		return emailNotVerified
	}

	return nil
}

// VerifyEmail verifies a user's email
func (s *UserService) VerifyEmail(ctx context.Context, token string) error {
	claims, err := utils.ParseEmailVerificationToken(token)
	if err != nil {
		return err
	}

	// Check if the token has expired
	if time.Now().After(claims.ExpiresAt.Time) {
		return fmt.Errorf("token has expired")
	}

	// Update user email verification status
	err = s.userRepository.UpdateUserEmailVerified(ctx, claims.Email)
	if err != nil {
		return err
	}

	// Delete verification token
	err = s.verificationTokenRepository.DeleteVerificationToken(ctx, claims.Email)
	if err != nil {
		return err
	}

	return nil
}

// GetUserInfo gets a user's info
func (s *UserService) GetUserInfo(ctx context.Context, userId string) (*model.UserProfile, error) {
	userIdUUID, err := uuid.Parse(userId)
	if err != nil {
		s.logger.Error("Failed to parse user id", zap.Error(err))
		return nil, err
	}

	user, err := s.userRepository.GetUserProfileByID(ctx, userIdUUID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserInfo updates a user's info
func (s *UserService) UpdateUserInfo(ctx context.Context, params model.UpdateUserInfoParams, image *model.ImageFile) error {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get user ID from context", zap.Error(err))
		return err
	}

	// get the user profile
	user, err := s.userRepository.GetUserProfileByID(ctx, userID)
	if err != nil {
		return err
	}

	imageUrl, err := s.handleUserProfileImage(ctx, image, user.ProfileImageUrl)
	if err != nil {
		return err
	}

	dateOfBirth, err := time.Parse("2006-01-02", params.DateOfBirth)
	if err != nil {
		return err
	}

	values := database.UpdateUserInfoParams{
		ID: userID,
		ProfileImageUrl: sql.NullString{
			String: imageUrl,
			Valid:  true,
		},
		FirstName: sql.NullString{
			String: params.FirstName,
			Valid:  true,
		},
		LastName: sql.NullString{
			String: params.LastName,
			Valid:  true,
		},
		PhoneNumber: sql.NullString{
			String: params.PhoneNumber,
			Valid:  true,
		},
		DateOfBirth: sql.NullTime{
			Time:  dateOfBirth,
			Valid: true,
		},
	}

	// Handle admin request if applicable
	if params.RequestAdmin {
		// check if the admin request is still pending
		adminRequest, err := s.adminRequestRepository.GetAdminRequestByUserID(ctx, userID)
		if err != nil {
			return err
		}

		if adminRequest.Status != "PENDING" {

			_, err = s.adminRequestRepository.CreateAdminRequest(ctx, userID, params.AdminRequestReason)
			if err != nil {
				return err
			}

			return nil
		}

		return nil

	}

	err = s.userRepository.UpdateUserInfo(ctx, values)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("userId").(uuid.UUID)
	if !ok {
		return uuid.Nil, app_errors.NewUnauthorizedUserError()
	}
	return userID, nil
}

func (s *UserService) handleUserProfileImage(ctx context.Context, image *model.ImageFile, existingImageUrl string) (string, error) {
	if image == nil {
		return existingImageUrl, nil
	}

	filePath, err := utils.UploadCustomFileToS3(ctx, image.File, image.FileHeader, s.s3Client, s.config.AWSBucketName, "users")
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
