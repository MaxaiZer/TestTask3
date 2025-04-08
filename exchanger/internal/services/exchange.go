package services

import (
	"context"
	"test-task/exchanger/internal/models"
)

type Storage interface {
	GetRate(ctx context.Context, from models.Currency, to models.Currency) (float64, error)
	GetRates(ctx context.Context) (map[models.Currency]float64, error)
}

type ExchangeService struct {
	storage Storage
}

func NewExchangeService(storage Storage) *ExchangeService {
	return &ExchangeService{storage: storage}
}

func (e *ExchangeService) GetRate(ctx context.Context, from models.Currency, to models.Currency) (float64, error) {
	return e.storage.GetRate(ctx, from, to)
}

func (e *ExchangeService) GetRates(ctx context.Context) (map[models.Currency]float64, error) {
	return e.storage.GetRates(ctx)
}
