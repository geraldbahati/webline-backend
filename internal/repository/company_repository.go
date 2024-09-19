package repository

import (
	"context"

	"github.com/google/uuid"
)

type CompanyRepository interface {
	CreateCompany(ctx context.Context, name string, kraPIN string, address string, phone string, email string) (*uuid.UUID, error)
	GetCompanyID(ctx context.Context, name string, kraPIN string) (*uuid.UUID, error)
}
