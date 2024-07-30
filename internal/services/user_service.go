package services

import (
	"context"
	"database/sql"
	"errors"
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
	userRepository     *repository.UserRepository
	roleRepository     *repository.RoleRepository
	userRoleRepository *repository.UserRoleRepository
	tokenRepository    *repository.TokenRepository
	config             *appconfig.Config
	logger             *zap.Logger
}

func NewUserService(
	userRepository *repository.UserRepository,
	roleRepository *repository.RoleRepository,
	userRoleRepository *repository.UserRoleRepository,
	tokenRepository *repository.TokenRepository,
	config *appconfig.Config,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		userRepository:     userRepository,
		roleRepository:     roleRepository,
		userRoleRepository: userRoleRepository,
		tokenRepository:    tokenRepository,
		config:             config,
		logger:             logger,
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
		HashedPassword: sql.NullString{String: string(hashedPassword), Valid: true},
	}

	createdUser, err := s.userRepository.CreateUser(ctx, newUser)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	// assign user role
	role, err := s.roleRepository.GetRoleByName(ctx, "customer")
	if err != nil {
		s.logger.Error("Failed to get role", zap.Error(err))
		return nil, err
	}

	// Assign role to user
	err = s.userRoleRepository.AssignRoleToUser(ctx, createdUser.ID, role.ID)
	if err != nil {
		s.logger.Error("Failed to assign role to user", zap.Error(err))
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

	roles, err := s.userRoleRepository.GetRolesForUser(ctx, createdUser.ID)
	if err != nil {
		s.logger.Error("Failed to get roles for user", zap.Error(err))
		return nil, err

	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: model.User{
			ID:    createdUser.ID,
			Email: createdUser.Email,
			Roles: roles,
		},
	}, nil
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
		return model.LoginResponse{}, errors.New("user does not have a password")
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
		ID: userId,
		HashedPassword: sql.NullString{
			String: string(hashedPassword),
			Valid:  true,
		},
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
func (s *UserService) GetUserProfile(ctx context.Context, userId string) (*model.User, error) {
	userIdUUID, err := uuid.Parse(userId)
	if err != nil {
		s.logger.Error("Failed to parse user id", zap.Error(err))
		return nil, err
	}

	var user *model.User
	user, err = s.userRepository.GetUserByID(ctx, userIdUUID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			user, err = s.userRepository.GetUserByProvider(ctx, "google", userId)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil, err
				}

				s.logger.Error("Failed to get user by provider", zap.Error(err))
				return nil, err
			}
		default:
			s.logger.Error("Failed to get user by id", zap.Error(err))
			return nil, err
		}
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
