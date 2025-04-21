package clients

import (
	"context"
	"fmt"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"test-task/api/gen/grpc/exchange"
	"test-task/wallet/internal/domain/models"
)

type ExchangerClient struct {
	conn   *grpc.ClientConn
	client exchange.ExchangeClient
}

func NewExchangerClient(url string) (*ExchangerClient, error) {
	conn, err := grpc.NewClient(
		url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}

	client := exchange.NewExchangeClient(conn)
	return &ExchangerClient{conn: conn, client: client}, nil
}

func NewExchangerClientWithConsul(consulAddress string) (*ExchangerClient, error) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s/%s?wait=14s", consulAddress, "exchanger-service"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		return nil, err
	}

	client := exchange.NewExchangeClient(conn)
	return &ExchangerClient{conn: conn, client: client}, nil
}

func (e *ExchangerClient) GetExchangeRates(ctx context.Context) (map[models.Currency]float64, error) {
	resp, err := e.client.GetExchangeRates(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	rates := make(map[models.Currency]float64)
	for currency, rate := range resp.GetRates() {
		rates[models.Currency(currency)] = rate
	}
	return rates, nil
}

func (e *ExchangerClient) GetExchangeRateForOne(ctx context.Context, from models.Currency, to models.Currency) (float64, error) {
	resp, err := e.client.GetExchangeRateForOne(ctx, &exchange.ExchangeRateRequest{
		FromCurrency: string(from),
		ToCurrency:   string(to),
	})
	if err != nil {
		return 0, err
	}
	return resp.GetRate(), nil
}

func (e *ExchangerClient) Close() {
	e.conn.Close()
}
