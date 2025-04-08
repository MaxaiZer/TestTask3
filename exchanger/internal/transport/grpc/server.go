package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"test-task/api/gen/grpc/exchange"
	"test-task/exchanger/internal/models"
)

type ExchangeService interface {
	GetRate(ctx context.Context, from models.Currency, to models.Currency) (float64, error)
	GetRates(ctx context.Context) (map[models.Currency]float64, error)
}

type ExchangeServer struct {
	service ExchangeService
	exchange.UnimplementedExchangeServer
}

func NewExchangeServer(service ExchangeService) *ExchangeServer {
	return &ExchangeServer{service: service}
}

func (e ExchangeServer) GetExchangeRates(ctx context.Context, in *emptypb.Empty) (*exchange.ExchangeRatesResponse, error) {

	rates, err := e.service.GetRates(ctx)
	if err != nil {
		return nil, err
	}

	convertedRates := make(map[string]float64)
	for key, value := range rates {
		convertedRates[string(key)] = value
	}

	return &exchange.ExchangeRatesResponse{Rates: convertedRates}, nil
}

func (e ExchangeServer) GetExchangeRateForOne(ctx context.Context, in *exchange.ExchangeRateRequest) (*exchange.ExchangeRateResponse, error) {

	from := models.Currency(in.FromCurrency)
	if !from.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid currency to convert to")
	}

	to := models.Currency(in.ToCurrency)
	if !to.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid currency to convert from")
	}

	res, err := e.service.GetRate(ctx, from, to)
	if err != nil {
		return nil, err
	}

	return &exchange.ExchangeRateResponse{Rate: res}, nil
}
