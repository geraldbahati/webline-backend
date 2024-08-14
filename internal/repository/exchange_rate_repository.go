package repository

import (
	"context"
	"time"
)

type ExchangeRateRepository interface {
	GetLatestExchangeRate(ctx context.Context, baseCurrency string) (float64, error)
	InsertExchangeRate(ctx context.Context, baseCurrency string, rateToKes float64, validFrom time.Time, validTo time.Time) error
	UpdateExchangeRate(ctx context.Context, baseCurrency string, rateToKes float64, validFrom time.Time, validTo time.Time) error
}
