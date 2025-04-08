package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"test-task/api/gen/grpc/exchange"
	"test-task/exchanger/internal/models"
	"testing"
)

func TestGetRates(t *testing.T) {

	ratesResponse, err := exchangeClient.GetExchangeRates(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)

	assert.Equal(t, convertRates(rates), ratesResponse.Rates)
}

func TestGetRate_FromBaseCurrency(t *testing.T) {

	req := exchange.ExchangeRateRequest{
		FromCurrency: string(models.USD),
		ToCurrency:   string(models.RUB),
	}
	resp, err := exchangeClient.GetExchangeRateForOne(context.Background(), &req)
	require.NoError(t, err)

	assert.Equal(t, rates[models.Currency(req.ToCurrency)], resp.Rate)
}

func TestGetRate_ToBaseCurrency(t *testing.T) {

	req := exchange.ExchangeRateRequest{
		FromCurrency: string(models.RUB),
		ToCurrency:   string(models.USD),
	}
	resp, err := exchangeClient.GetExchangeRateForOne(context.Background(), &req)
	require.NoError(t, err)

	assert.Equal(t, 1/rates[models.Currency(req.FromCurrency)], resp.Rate)
}

func TestGetRate_NotBaseCurrencies(t *testing.T) {

	req := exchange.ExchangeRateRequest{
		FromCurrency: string(models.RUB),
		ToCurrency:   string(models.EUR),
	}
	resp, err := exchangeClient.GetExchangeRateForOne(context.Background(), &req)
	require.NoError(t, err)

	assert.Equal(t, rates[models.Currency(req.FromCurrency)]/rates[models.Currency(req.ToCurrency)], resp.Rate)
}

func TestGetRate_InvalidCurrency(t *testing.T) {

	req := exchange.ExchangeRateRequest{
		FromCurrency: "tugrik",
		ToCurrency:   string(models.EUR),
	}
	_, err := exchangeClient.GetExchangeRateForOne(context.Background(), &req)
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())

	req = exchange.ExchangeRateRequest{
		FromCurrency: string(models.RUB),
		ToCurrency:   "tugrik",
	}
	_, err = exchangeClient.GetExchangeRateForOne(context.Background(), &req)
	assert.Error(t, err)

	st, ok = status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}
