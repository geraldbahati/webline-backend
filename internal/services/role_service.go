package services

import (
	"context"
	"go.uber.org/zap"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
)

type RoleService struct {
	roleRepo *repository.RoleRepository
	logger   *zap.Logger
}

// NewRoleService initializes a new RoleService with dependency injection for logging
func NewRoleService(roleRepo *repository.RoleRepository, logger *zap.Logger) *RoleService {
	return &RoleService{
		roleRepo: roleRepo,
		logger:   logger,
	}
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(ctx context.Context, name, description string) (*model.Role, error) {
	return s.roleRepo.CreateRole(ctx, name, description)
}
