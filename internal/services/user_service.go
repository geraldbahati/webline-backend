package services

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
	"weblineBackend/internal/config"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/logger"
	"weblineBackend/pkg/utils"
)

type UserService struct {
	userRepository  *repository.UserRepository
	tokenRepository *repository.TokenRepository
}

func NewUserService(
	userRepository *repository.UserRepository,
	tokenRepository *repository.TokenRepository,
) *UserService {
	return &UserService{
		userRepository:  userRepository,
		tokenRepository: tokenRepository,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, registerUserParams model.RegisterUserParams) (database.User, error) {
	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerUserParams.Password), bcrypt.DefaultCost)
	if err != nil {
		return database.User{}, err
	}

	// generate username
	username, err := generateUsername(ctx, s.userRepository, registerUserParams.FirstName, registerUserParams.LastName)
	if err != nil {
		return database.User{}, err
	}

	// create user
	newUser := database.CreateUserParams{
		Username:       username,
		Email:          registerUserParams.Email,
		HashedPassword: string(hashedPassword),
		FirstName: sql.NullString{
			String: strings.ToLower(registerUserParams.FirstName),
			Valid:  true,
		},
		LastName: sql.NullString{
			String: strings.ToLower(registerUserParams.LastName),
			Valid:  true,
		},
	}

	createdUser, err := s.userRepository.CreateUser(ctx, newUser)
	if err != nil {
		return database.User{}, err
	}

	// return created user
	return createdUser, nil
}

// sanitizeUsername sanitizes the given username
func sanitizeUsername(username string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9_.-]+")
	return reg.ReplaceAllString(username, "")
}

// generateUsername generates a username from the given first and last name
func generateUsername(ctx context.Context, userRepo *repository.UserRepository, firstName string, lastName string) (string, error) {
	// generate username from first and last name
	baseUsername := strings.ToLower(sanitizeUsername(firstName + lastName))
	log.Printf("baseUsername: %v", baseUsername)
	username := baseUsername
	const maxUsernameLength = 20

	// trim username to max length
	if len(baseUsername) > maxUsernameLength {
		baseUsername = baseUsername[:maxUsernameLength]
		username = baseUsername
	}

	for suffix := 1; ; suffix++ {
		// check if username is available
		count, err := userRepo.CountUsersByUsername(ctx, username)
		if err != nil {
			return "", err
		}
		if count == 0 {
			return username, nil
		}

		// append suffix to username
		suffixStr := strconv.Itoa(suffix)
		cutOffLength := maxUsernameLength - len(suffixStr)
		if cutOffLength < len(baseUsername) {
			username = baseUsername[:cutOffLength]
		}

		username += suffixStr
	}
}

// GetUserByEmail returns the user with the given email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	// get user from database
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return database.User{}, err
	}

	// return user
	return user, nil
}

// LoginUser logs in a user
func (s *UserService) LoginUser(ctx context.Context, params model.LoginParams) (model.LoginResponse, error) {
	// get user by email
	user, err := s.userRepository.GetUserByEmail(ctx, params.Email)
	if err != nil {
		return model.LoginResponse{}, err
	}

	// compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(params.Password))
	if err != nil {
		return model.LoginResponse{}, err
	}

	// generate access token and refresh token
	accessToken, refreshToken, expireTime, err := utils.GenerateTokens(user.ID, user.Username, user.Email)
	if err != nil {
		return model.LoginResponse{}, err
	}

	// save refresh token
	err = s.tokenRepository.StoreRefreshToken(ctx, database.StoreRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: expireTime,
	})
	if err != nil {
		return model.LoginResponse{}, err
	}

	// update last login
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
	// generate access token
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
	// get user id from context
	userId := ctx.Value("userId").(uuid.UUID)
	user, err := s.userRepository.GetUserByID(ctx, userId)
	if err != nil {
		return database.User{}, err
	}

	// update user profile
	if params.FirstName != "" {
		firstNameValue := sql.NullString{String: params.FirstName, Valid: params.FirstName != ""}
		user.FirstName = firstNameValue
	}

	if params.LastName != "" {
		lastNameValue := sql.NullString{String: params.LastName, Valid: params.LastName != ""}
		user.LastName = lastNameValue
	}

	if params.PhoneNumber != "" {
		phoneNumberValue := sql.NullString{String: params.PhoneNumber, Valid: params.PhoneNumber != ""}
		user.PhoneNumber = phoneNumberValue
	}

	if params.ProfileImageUrl != "" {
		genderValue := sql.NullString{String: params.ProfileImageUrl, Valid: params.ProfileImageUrl != ""}
		user.ProfileImageUrl = genderValue
	}

	if params.DateOfBirth != "" {
		dateOfBirthDate, err := time.Parse("02-01-2006", params.DateOfBirth)
		if err != nil {
			dateOfBirthDate = time.Time{}
		}
		dateOfBirthValue := sql.NullTime{Time: dateOfBirthDate, Valid: dateOfBirthDate != time.Time{}}
		user.DateOfBirth = dateOfBirthValue
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

	// return updated user
	return updatedUser, nil
}

// SendPasswordResetEmail sends a password reset email to the user
func (s *UserService) SendPasswordResetEmail(ctx context.Context, email string) error {
	// get user by email
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	return utils.SendResetPasswordEmail(user.ID, user.Email)

}

// UpdateUserPassword updates a user's password
func (s *UserService) UpdateUserPassword(ctx context.Context, token string, newPassword string) error {
	// verify reset password token
	userId, err := utils.VerifyResetPasswordToken(token)
	if err != nil {
		return err
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// update user password
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
	// verify reset password token
	passwordToken, err := utils.VerifyResetPasswordToken(token)
	return passwordToken, err
}

// DeactivateUser deactivates a user's account
func (s *UserService) DeactivateUser(ctx context.Context) error {
	// get user id from context
	userId := ctx.Value("userId").(uuid.UUID)

	// deactivate user
	_, err := s.userRepository.DeactivateUser(ctx, userId)
	if err != nil {
		return err
	}

	return nil
}

// ListUsers lists all users
func (s *UserService) ListUsers(ctx context.Context, pageSize int32, page int32) (model.PaginationResult[[]database.User], error) {
	// get total user count
	count, err := s.userRepository.CountAllUsers(ctx)
	if err != nil {
		return model.PaginationResult[[]database.User]{}, err
	}

	// get users
	paginatedListUsers, err := utils.Paginate(
		config.LoadConfig(),
		count,
		page,
		pageSize,
		func(offset int32, limit int32) ([]database.User, error) {
			return s.userRepository.ListUsers(ctx, limit, offset)
		},
	)
	if err != nil {
		logger.Error("failed to fetch users: (US)")
		return model.PaginationResult[[]database.User]{}, err
	}

	// return paginated list of users
	return *paginatedListUsers, nil
}
