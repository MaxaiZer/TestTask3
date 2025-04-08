package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"test-task/exchanger/internal/models"
)

type Storage struct {
	pool         *pgxpool.Pool
	baseCurrency models.Currency
}

func NewStorage(pool *pgxpool.Pool) (*Storage, error) {

	baseCurrency, err := getBaseCurrency(context.Background(), pool)
	if err != nil {
		return nil, fmt.Errorf("failed to set base currency: %w", err)
	}

	return &Storage{pool: pool, baseCurrency: baseCurrency}, nil
}

func (p *Storage) SetBaseCurrency(currency models.Currency) {
	p.baseCurrency = currency
}

func (p *Storage) GetRate(ctx context.Context, from models.Currency, to models.Currency) (float64, error) {

	if from == to {
		return 1, nil
	}

	if from == p.baseCurrency {
		return p.getRate(ctx, to)
	}

	if to == p.baseCurrency {
		rate, err := p.getRate(ctx, from)
		return 1 / rate, err
	}

	from_rate, err := p.getRate(ctx, from)
	if err != nil {
		return 0, fmt.Errorf("failed to get first rate: %w", err)
	}

	to_rate, err := p.getRate(ctx, to)
	if err != nil {
		return 0, fmt.Errorf("failed to get second rate: %w", err)
	}

	return from_rate / to_rate, nil
}

func (p *Storage) GetRates(ctx context.Context) (map[models.Currency]float64, error) {

	rates := make(map[models.Currency]float64)
	rows, err := p.pool.Query(ctx, "SELECT currency, rate FROM exchange_rates")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rates from DB: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var currency models.Currency
		rate := 0.0

		err = rows.Scan(&currency, &rate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row during rates fetching: %w", err)
		}

		rates[currency] = rate
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to fetch rates from DB: %w", rows.Err())
	}

	return rates, nil
}

func (p *Storage) GetBaseCurrency(ctx context.Context) (models.Currency, error) {

	var currencyStr string
	res := p.pool.QueryRow(ctx, "SELECT currency FROM exchange_rates WHERE rate = 1")

	err := res.Scan(&currencyStr)
	if err != nil {
		return "", err
	}

	currency := models.Currency(currencyStr)
	if !currency.IsValid() {
		return "", fmt.Errorf("invalid currency '%s'", currencyStr)
	}

	return currency, nil
}

func (p *Storage) getRate(ctx context.Context, currency models.Currency) (float64, error) {

	rate := 0.0
	row := p.pool.QueryRow(ctx, "SELECT rate FROM exchange_rates WHERE currency = $1", string(currency))

	err := row.Scan(&rate)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch rate from DB: %w", err)
	}
	return rate, nil
}

func getBaseCurrency(ctx context.Context, pool *pgxpool.Pool) (models.Currency, error) {

	var currencyStr string
	res := pool.QueryRow(ctx, "SELECT currency FROM exchange_rates WHERE rate = 1")

	err := res.Scan(&currencyStr)
	if err != nil {
		return "", err
	}

	currency := models.Currency(currencyStr)
	if !currency.IsValid() {
		return "", fmt.Errorf("invalid currency '%s'", currencyStr)
	}

	return currency, nil
}
