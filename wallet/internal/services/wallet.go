package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	errs "test-task/wallet/internal/domain/errors"
	"test-task/wallet/internal/domain/models"
	"test-task/wallet/internal/tracing"
	"time"
)

type Redis interface {
	StoreRate(ctx context.Context, from models.Currency, to models.Currency, value float64, expiration time.Duration) error
	StoreRates(ctx context.Context, rates map[models.Currency]float64, base models.Currency, expiration time.Duration) error
	GetRate(ctx context.Context, from models.Currency, to models.Currency) (float64, error)
	GetRates(ctx context.Context, base models.Currency) (map[models.Currency]float64, error)
}

type ExchangerClient interface {
	GetExchangeRates(ctx context.Context) (map[models.Currency]float64, error)
	GetExchangeRateForOne(ctx context.Context, from models.Currency, to models.Currency) (float64, error)
}

type AccountsRepository interface {
	GetBalance(ctx context.Context, userID string) (map[models.Currency]float64, error)
	ChangeAccountAmountWithBalance(ctx context.Context, userID string, currency models.Currency,
		delta float64) (map[models.Currency]float64, error)
	ExchangeAccountAmountWithBalance(ctx context.Context, userID string, from models.Currency,
		fromAmount float64, to models.Currency, toAmount float64) (map[models.Currency]float64, error)
}

type BalanceInfo struct {
	Accounts map[models.Currency]float64
}

type ExchangeInfo struct {
	Accounts        map[models.Currency]float64
	ExchangedAmount float64
}

type WalletService struct {
	accounts        AccountsRepository
	exchangerClient ExchangerClient
	redis           Redis
	ratesExpiration time.Duration
}

func NewWalletService(accounts AccountsRepository, exchangerClient ExchangerClient, redis Redis) *WalletService {
	return &WalletService{
		accounts:        accounts,
		exchangerClient: exchangerClient,
		redis:           redis,
		ratesExpiration: 5 * time.Minute,
	}
}

func (w *WalletService) GetExchangeRates(ctx context.Context) (map[models.Currency]float64, error) {

	ctx, span := tracing.GetTracer().Start(ctx, "GetExchangeRates")
	defer span.End()

	rates, err := w.redis.GetRates(ctx, models.USD)
	if err != nil {
		if errors.Is(err, errs.KeyNotExists) {
			slog.Debug("key does not exist")
		} else {
			return nil, fmt.Errorf("failed to get rates from cache: %w", err)
		}
	} else {
		return rates, nil
	}

	rates, err = w.exchangerClient.GetExchangeRates(ctx)
	if err != nil {
		return nil, err
	}

	err = w.redis.StoreRates(ctx, rates, models.USD, w.ratesExpiration)
	if err != nil {
		slog.Error("failed to store rates in cache:", "error", err)
	}

	return rates, nil
}

func (w *WalletService) Exchange(ctx context.Context, userID string, from models.Currency, to models.Currency, amount float64) (*ExchangeInfo, error) {

	ctx, span := tracing.GetTracer().Start(ctx, "Exchange")
	defer span.End()

	if amount <= 0 {
		return nil, errs.InvalidAmount
	}

	if !from.IsValid() || !to.IsValid() || from == to {
		return nil, errs.InvalidCurrency
	}

	rate, err := w.getExchangeRate(ctx, from, to)
	if err != nil {
		return nil, err
	}

	balance, err := w.accounts.ExchangeAccountAmountWithBalance(ctx, userID, from, amount, to, amount*rate)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange: %w", err)
	}

	return &ExchangeInfo{Accounts: balance, ExchangedAmount: amount * rate}, nil
}

func (w *WalletService) Withdraw(ctx context.Context, userID string, currency models.Currency, amount float64) (*BalanceInfo, error) {

	ctx, span := tracing.GetTracer().Start(ctx, "Withdraw")
	defer span.End()

	if amount <= 0 {
		return nil, errs.InvalidAmount
	}

	if !currency.IsValid() {
		return nil, errs.InvalidCurrency
	}

	balance, err := w.accounts.ChangeAccountAmountWithBalance(ctx, userID, currency, -amount)
	if err != nil {
		return nil, err
	}
	return &BalanceInfo{Accounts: balance}, nil
}

func (w *WalletService) Deposit(ctx context.Context, userID string, currency models.Currency, amount float64) (*BalanceInfo, error) {

	ctx, span := tracing.GetTracer().Start(ctx, "Deposit")
	defer span.End()

	if amount <= 0 {
		return nil, errs.InvalidAmount
	}

	if !currency.IsValid() {
		return nil, errs.InvalidCurrency
	}

	balance, err := w.accounts.ChangeAccountAmountWithBalance(ctx, userID, currency, amount)
	if err != nil {
		return nil, err
	}
	return &BalanceInfo{Accounts: balance}, nil
}

func (w *WalletService) GetBalance(ctx context.Context, userID string) (*BalanceInfo, error) {

	ctx, span := tracing.GetTracer().Start(ctx, "GetBalance")
	defer span.End()

	balance, err := w.accounts.GetBalance(ctx, userID)
	return &BalanceInfo{Accounts: balance}, err
}

func (w *WalletService) getExchangeRate(ctx context.Context, from models.Currency, to models.Currency) (float64, error) {

	ctx, span := tracing.GetTracer().Start(ctx, "getExchangeRate")
	defer span.End()

	rate, err := w.redis.GetRate(ctx, from, to)
	if err != nil {
		if errors.Is(err, errs.KeyNotExists) {
			slog.Debug("key does not exist")
		} else {
			return 0, fmt.Errorf("failed to get rate from cache: %w", err)
		}
	} else {
		return rate, nil
	}

	rate, err = w.exchangerClient.GetExchangeRateForOne(ctx, from, to)
	if err != nil {
		return 0, err
	}

	err = w.redis.StoreRate(ctx, from, to, rate, w.ratesExpiration)
	if err != nil {
		slog.Error("failed to store rate in cache", "error", err)
	}

	return rate, nil
}
