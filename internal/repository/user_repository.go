package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

// NewUserRepository initializes a new UserRepository with dependency injection for logging
func NewUserRepository(db *sql.DB, logger *zap.Logger) *UserRepository {
	return &UserRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *UserRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	q := database.New(tx)
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			r.logger.Error("transaction failed, rolling back", zap.Error(err))
			if rbErr := tx.Rollback(); rbErr != nil {
				r.logger.Error("rollback failed", zap.Error(rbErr))
				err = fmt.Errorf("rollback transaction: %w", rbErr)
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("commit transaction: %w", commitErr)
			}
		}
	}()

	err = fn(q)
	return err
}

var ErrUserAlreadyExists = errors.New("user already exists")

// CreateUser creates a new user
func (r *UserRepository) CreateUser(
	ctx context.Context,
	user database.CreateUserParams,
) (*model.User, error) {
	var createdUser database.CreateUserRow
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		createdUser, err = q.CreateUser(ctx, user)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				r.logger.Error("user already exists", zap.Error(err))
				return ErrUserAlreadyExists
			}
			r.logger.Error("failed to create user", zap.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		r.logger.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	r.logger.Info("User created successfully", zap.String("userID", createdUser.ID.String()))

	var name string
	if createdUser.FirstName.Valid && createdUser.LastName.Valid {
		name = createdUser.FirstName.String + " " + createdUser.LastName.String
	}
	var date string
	if createdUser.DateOfBirth.Valid {
		date = createdUser.DateOfBirth.Time.String()
	}

	var provider, providerID *string
	if createdUser.Provider.Valid {
		provider = &createdUser.Provider.String
	}

	if createdUser.ProviderID.Valid {
		providerID = &createdUser.ProviderID.String
	}

	return &model.User{
		ID:          createdUser.ID,
		Email:       createdUser.Email,
		Name:        name,
		Phone:       createdUser.PhoneNumber.String,
		Image:       createdUser.ProfileImageUrl.String,
		DateOfBirth: date,
		IsActive:    createdUser.IsActive,
		Provider:    provider,
		ProviderID:  providerID,
	}, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := r.Queries.GetUserByID(ctx, id)
	if err != nil {
		r.logger.Error("failed to get user by ID", zap.Error(err))
		return nil, err
	}

	var name string
	if user.FirstName.Valid && user.LastName.Valid {
		name = user.FirstName.String + " " + user.LastName.String
	}
	var date string
	if user.DateOfBirth.Valid {
		date = user.DateOfBirth.Time.String()
	}

	var providerValue, providerIDValue *string
	if user.Provider.Valid {
		providerValue = &user.Provider.String
	}

	if user.ProviderID.Valid {
		providerIDValue = &user.ProviderID.String
	}

	return &model.User{
		ID:          user.ID,
		Email:       user.Email,
		Name:        name,
		Phone:       user.PhoneNumber.String,
		Image:       user.ProfileImageUrl.String,
		DateOfBirth: date,
		IsActive:    user.IsActive,
		Provider:    providerValue,
		ProviderID:  providerIDValue,
	}, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := r.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		r.logger.Error("failed to get user by email", zap.Error(err))
		return nil, err
	}

	var name string
	if user.FirstName.Valid && user.LastName.Valid {
		name = user.FirstName.String + " " + user.LastName.String
	}
	var date string
	if user.DateOfBirth.Valid {
		date = user.DateOfBirth.Time.String()
	}

	var providerValue, providerIDValue *string
	if user.Provider.Valid {
		providerValue = &user.Provider.String
	}

	if user.ProviderID.Valid {
		providerIDValue = &user.ProviderID.String
	}

	return &model.User{
		ID:          user.ID,
		Email:       user.Email,
		Name:        name,
		Password:    user.HashedPassword.String,
		Phone:       user.PhoneNumber.String,
		Image:       user.ProfileImageUrl.String,
		DateOfBirth: date,
		IsActive:    user.IsActive,
		Provider:    providerValue,
		ProviderID:  providerIDValue,
	}, nil
}

// UpdateUserProfile updates a user's profile
func (r *UserRepository) UpdateUserProfile(
	ctx context.Context,
	user database.UpdateUserProfileParams,
) (*model.User, error) {
	updatedUser, err := r.Queries.UpdateUserProfile(ctx, user)
	if err != nil {
		r.logger.Error("failed to update user profile", zap.Error(err))
		return nil, err
	}

	r.logger.Info("User profile updated successfully", zap.String("userID", updatedUser.ID.String()))
	var name string
	if updatedUser.FirstName.Valid && updatedUser.LastName.Valid {
		name = updatedUser.FirstName.String + " " + updatedUser.LastName.String
	}
	var date string
	if updatedUser.DateOfBirth.Valid {
		date = updatedUser.DateOfBirth.Time.String()
	}

	var providerValue, providerIDValue *string
	if updatedUser.Provider.Valid {
		providerValue = &updatedUser.Provider.String
	}

	if updatedUser.ProviderID.Valid {
		providerIDValue = &updatedUser.ProviderID.String
	}

	return &model.User{
		ID:          updatedUser.ID,
		Email:       updatedUser.Email,
		Name:        name,
		Phone:       updatedUser.PhoneNumber.String,
		Image:       updatedUser.ProfileImageUrl.String,
		DateOfBirth: date,
		IsActive:    updatedUser.IsActive,
		Provider:    providerValue,
		ProviderID:  providerIDValue,
	}, nil
}

// UpdateUserPassword updates a user's password
func (r *UserRepository) UpdateUserPassword(
	ctx context.Context,
	user database.UpdateUserPasswordParams,
) (database.User, error) {
	updatedUser, err := r.Queries.UpdateUserPassword(ctx, user)
	if err != nil {
		r.logger.Error("failed to update user password", zap.Error(err))
		return database.User{}, err
	}

	r.logger.Info("User password updated successfully", zap.String("userID", updatedUser.ID.String()))
	return database.User{
		ID:              updatedUser.ID,
		Email:           updatedUser.Email,
		FirstName:       updatedUser.FirstName,
		LastName:        updatedUser.LastName,
		PhoneNumber:     updatedUser.PhoneNumber,
		ProfileImageUrl: updatedUser.ProfileImageUrl,
		DateOfBirth:     updatedUser.DateOfBirth,
		IsActive:        updatedUser.IsActive,
		CreatedAt:       updatedUser.CreatedAt,
		UpdatedAt:       updatedUser.UpdatedAt,
		LastLogin:       updatedUser.LastLogin,
	}, nil
}

// UpdateUserLastLogin updates a user's last login
func (r *UserRepository) UpdateUserLastLogin(ctx context.Context, id uuid.UUID) (database.User, error) {
	updatedUser, err := r.Queries.UpdateUserLastLogin(ctx, id)
	if err != nil {
		r.logger.Error("failed to update user last login", zap.Error(err))
		return database.User{}, err
	}

	r.logger.Info("User last login updated successfully", zap.String("userID", updatedUser.ID.String()))
	return database.User{
		ID:              updatedUser.ID,
		Email:           updatedUser.Email,
		FirstName:       updatedUser.FirstName,
		LastName:        updatedUser.LastName,
		PhoneNumber:     updatedUser.PhoneNumber,
		ProfileImageUrl: updatedUser.ProfileImageUrl,
		DateOfBirth:     updatedUser.DateOfBirth,
		IsActive:        updatedUser.IsActive,
		CreatedAt:       updatedUser.CreatedAt,
		UpdatedAt:       updatedUser.UpdatedAt,
		LastLogin:       updatedUser.LastLogin,
	}, nil
}

// DeactivateUser deactivates a user
func (r *UserRepository) DeactivateUser(ctx context.Context, id uuid.UUID) (database.User, error) {
	updatedUser, err := r.Queries.DeactivateUser(ctx, id)
	if err != nil {
		r.logger.Error("failed to deactivate user", zap.Error(err))
		return database.User{}, err
	}

	r.logger.Info("User deactivated successfully", zap.String("userID", updatedUser.ID.String()))
	return database.User{
		ID:              updatedUser.ID,
		Email:           updatedUser.Email,
		FirstName:       updatedUser.FirstName,
		LastName:        updatedUser.LastName,
		PhoneNumber:     updatedUser.PhoneNumber,
		ProfileImageUrl: updatedUser.ProfileImageUrl,
		DateOfBirth:     updatedUser.DateOfBirth,
		IsActive:        updatedUser.IsActive,
		CreatedAt:       updatedUser.CreatedAt,
		UpdatedAt:       updatedUser.UpdatedAt,
		LastLogin:       updatedUser.LastLogin,
	}, nil
}

// DeleteUser deletes a user
func (r *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	err := r.Queries.DeleteUser(ctx, id)
	if err != nil {
		r.logger.Error("failed to delete user", zap.Error(err))
		return err
	}

	r.logger.Info("User deleted successfully", zap.String("userID", id.String()))
	return nil
}

// ListUsers retrieves a list of users
func (r *UserRepository) ListUsers(ctx context.Context, limit int32, offset int32) ([]database.User, error) {
	users, err := r.Queries.ListUsers(ctx, database.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		r.logger.Error("failed to list users", zap.Error(err))
		return nil, err
	}

	userList := make([]database.User, len(users))
	for i, user := range users {
		userList[i] = database.User{
			ID:              user.ID,
			Email:           user.Email,
			FirstName:       user.FirstName,
			LastName:        user.LastName,
			PhoneNumber:     user.PhoneNumber,
			ProfileImageUrl: user.ProfileImageUrl,
			DateOfBirth:     user.DateOfBirth,
			IsActive:        user.IsActive,
			CreatedAt:       user.CreatedAt,
			UpdatedAt:       user.UpdatedAt,
			LastLogin:       user.LastLogin,
		}
	}

	return userList, nil
}

// CountAllUsers gets the total number of users
func (r *UserRepository) CountAllUsers(ctx context.Context) (int64, error) {
	count, err := r.Queries.CountAllUsers(ctx)
	if err != nil {
		r.logger.Error("failed to get the total number of users", zap.Error(err))
		return 0, err
	}

	return count, nil
}

// GetUserByProvider retrieves a user by provider
func (r *UserRepository) GetUserByProvider(ctx context.Context, provider string, providerID string) (*model.User, error) {
	user, err := r.Queries.GetUserByProvider(ctx, database.GetUserByProviderParams{
		Provider:   sql.NullString{String: provider, Valid: true},
		ProviderID: sql.NullString{String: providerID, Valid: true},
	})
	if err != nil {
		r.logger.Error("failed to get user by provider", zap.Error(err))
		return nil, err
	}

	var name string
	if user.FirstName.Valid && user.LastName.Valid {
		name = user.FirstName.String + " " + user.LastName.String
	}
	var date string
	if user.DateOfBirth.Valid {
		date = user.DateOfBirth.Time.String()
	}

	var providerValue, providerIDValue *string
	if user.Provider.Valid {
		providerValue = &user.Provider.String
	}

	if user.ProviderID.Valid {
		providerIDValue = &user.ProviderID.String
	}

	return &model.User{
		ID:          user.ID,
		Email:       user.Email,
		Name:        name,
		Phone:       user.PhoneNumber.String,
		Image:       user.ProfileImageUrl.String,
		DateOfBirth: date,
		IsActive:    user.IsActive,
		Provider:    providerValue,
		ProviderID:  providerIDValue,
	}, nil
}
