package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"weblineBackend/internal/app_errors"

	"go.uber.org/zap"
	"log"
	"strings"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"

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
	config                      *appconfig.Config
	logger                      *zap.Logger
}

func NewUserService(
	userRepository *repository.UserRepository,
	roleRepository *repository.RoleRepository,
	userRoleRepository *repository.UserRoleRepository,
	verificationTokenRepository repository.VerificationTokenRepository,
	passwordResetRepository repository.PasswordResetRepository,
	tokenRepository *repository.TokenRepository,
	config *appconfig.Config,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		userRepository:              userRepository,
		roleRepository:              roleRepository,
		userRoleRepository:          userRoleRepository,
		verificationTokenRepository: verificationTokenRepository,
		passwordResetRepository:     passwordResetRepository,
		tokenRepository:             tokenRepository,
		config:                      config,
		logger:                      logger,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, registerUserParams model.RegisterUserParams) error {
	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerUserParams.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return err
	}

	// create user
	newUser := database.CreateUserParams{
		Email:          registerUserParams.Email,
		HashedPassword: sql.NullString{String: string(hashedPassword), Valid: true},
	}

	createdUser, err := s.userRepository.CreateUser(ctx, newUser)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return err
	}

	// assign user role
	role, err := s.roleRepository.GetRoleByName(ctx, "customer")
	if err != nil {
		s.logger.Error("Failed to get role", zap.Error(err))
		return err
	}

	// Assign role to user
	err = s.userRoleRepository.AssignRoleToUser(ctx, createdUser.ID, role.ID)
	if err != nil {
		s.logger.Error("Failed to assign role to user", zap.Error(err))
		return err
	}

	return nil
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

//// UpdateUserProfile updates a user's profile
//func (s *UserService) UpdateUserProfile(ctx context.Context, params model.UpdateUserProfileParams) (database.User, error) {
//	userId := ctx.Value("userId").(uuid.UUID)
//	user, err := s.userRepository.GetUserByID(ctx, userId)
//	if err != nil {
//		return database.User{}, err
//	}
//
//	if params.FirstName != "" {
//		user.FirstName = sql.NullString{String: params.FirstName, Valid: params.FirstName != ""}
//	}
//
//	if params.LastName != "" {
//		user.LastName = sql.NullString{String: params.LastName, Valid: params.LastName != ""}
//	}
//
//	if params.PhoneNumber != "" {
//		user.PhoneNumber = sql.NullString{String: params.PhoneNumber, Valid: params.PhoneNumber != ""}
//	}
//
//	if params.ProfileImageUrl != "" {
//		user.ProfileImageUrl = sql.NullString{String: params.ProfileImageUrl, Valid: params.ProfileImageUrl != ""}
//	}
//
//	if params.DateOfBirth != "" {
//		dateOfBirth, err := time.Parse("02-01-2006", params.DateOfBirth)
//		if err != nil {
//			dateOfBirth = time.Time{}
//		}
//		user.DateOfBirth = sql.NullTime{Time: dateOfBirth, Valid: !dateOfBirth.IsZero()}
//	}
//
//	updatedUser, err := s.userRepository.UpdateUserProfile(ctx, database.UpdateUserProfileParams{
//		ID:              user.ID,
//		FirstName:       user.FirstName,
//		LastName:        user.LastName,
//		PhoneNumber:     user.PhoneNumber,
//		ProfileImageUrl: user.ProfileImageUrl,
//		DateOfBirth:     user.DateOfBirth,
//	})
//	if err != nil {
//		return database.User{}, err
//	}
//
//	return updatedUser, nil
//}

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
	} else {
		s.logger.Info("User found")
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
