package sqlc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"time"
	"weblineBackend/internal/database"
)

type ExchangeRateRepoImpl struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewExchangeRateRepositoryImpl(db *sql.DB, logger *zap.Logger) *ExchangeRateRepoImpl {
	return &ExchangeRateRepoImpl{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *ExchangeRateRepoImpl) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
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

// GetLatestExchangeRate retrieves the latest exchange rate for the given base currency
func (r *ExchangeRateRepoImpl) GetLatestExchangeRate(ctx context.Context, baseCurrency string) (float64, error) {
	rate, err := r.Queries.GetLatestExchangeRate(ctx, baseCurrency)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return 135, nil
		}

		r.logger.Error("failed to get latest exchange rate", zap.Error(err))
		return 0, fmt.Errorf("get latest exchange rate: %w", err)
	}

	// convert rate to float64
	rateFloat, err := strconv.ParseFloat(rate, 64)
	if err != nil {
		r.logger.Error("failed to convert exchange rate to float64", zap.Error(err))
		return 0, fmt.Errorf("convert exchange rate to float64: %w", err)
	}

	return rateFloat, nil
}

// InsertExchangeRate inserts a new exchange rate record
func (r *ExchangeRateRepoImpl) InsertExchangeRate(ctx context.Context, baseCurrency string, rateToKes float64, validFrom, validTo time.Time) error {
	return r.execTx(ctx, func(q *database.Queries) error {
		err := q.InsertExchangeRate(ctx, database.InsertExchangeRateParams{
			CurrencyCode: baseCurrency,
			RateToKes:    strconv.FormatFloat(rateToKes, 'f', -1, 64),
			ValidFrom:    validFrom,
			ValidTo:      sql.NullTime{Time: validTo, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("insert exchange rate: %w", err)
		}

		return nil
	})
}

// UpdateExchangeRate updates an existing exchange rate record
func (r *ExchangeRateRepoImpl) UpdateExchangeRate(ctx context.Context, baseCurrency string, rateToKes float64, validFrom, validTo time.Time) error {
	return r.execTx(ctx, func(q *database.Queries) error {
		err := q.UpdateExchangeRate(ctx, database.UpdateExchangeRateParams{
			CurrencyCode: baseCurrency,
			RateToKes:    strconv.FormatFloat(rateToKes, 'f', -1, 64),
			ValidFrom:    validFrom,
			ValidTo:      sql.NullTime{Time: validTo, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("update exchange rate: %w", err)
		}

		return nil
	})
}
