package integration

import (
	"context"
	errs "test-task/wallet/internal/domain/errors"
	"test-task/wallet/internal/domain/models"
	"time"
)

type redisMock struct {
}

func (r redisMock) StoreRate(ctx context.Context, from models.Currency, to models.Currency, value float64, expiration time.Duration) error {
	return nil
}

func (r redisMock) StoreRates(ctx context.Context, rates map[models.Currency]float64, base models.Currency, expiration time.Duration) error {
	return nil
}

func (r redisMock) GetRate(ctx context.Context, from models.Currency, to models.Currency) (float64, error) {
	return 0, errs.KeyNotExists
}

func (r redisMock) GetRates(ctx context.Context, base models.Currency) (map[models.Currency]float64, error) {
	return nil, errs.KeyNotExists
}

type exchangerClientMock struct {
	base  models.Currency
	rates map[models.Currency]float64
}

func newExchangerClientMock(base models.Currency, rates map[models.Currency]float64) *exchangerClientMock {
	return &exchangerClientMock{base: base, rates: rates}
}

func (e exchangerClientMock) GetExchangeRates(ctx context.Context) (map[models.Currency]float64, error) {
	return e.rates, nil
}

func (e exchangerClientMock) GetExchangeRateForOne(ctx context.Context, from models.Currency, to models.Currency) (float64, error) {
	if from == to {
		return 1, nil
	}

	if from == e.base {
		return e.getRate(to), nil
	}
	if to == e.base {
		return 1 / e.getRate(from), nil
	}
	from_rate := e.getRate(from)
	to_rate := e.getRate(to)
	return from_rate / to_rate, nil
}

func (e exchangerClientMock) getRate(currency models.Currency) float64 {
	for k, v := range e.rates {
		if k == currency {
			return v
		}
	}
	return 0
}
