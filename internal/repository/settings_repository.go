package repository

import "context"

type SettingsRepository interface {
	GetVATPercentage(ctx context.Context) (float64, error)
	UpdateVATPercentage(ctx context.Context, percentage float64) error
}
