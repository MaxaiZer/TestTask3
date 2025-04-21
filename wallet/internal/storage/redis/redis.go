package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"strconv"
	errs "test-task/wallet/internal/domain/errors"
	"test-task/wallet/internal/domain/models"
	"time"
)

type Config struct {
	Address  string
	Password string
}

type Redis struct {
	client *redis.Client
}

func New(config Config) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Password: config.Password,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	if err = redisotel.InstrumentTracing(client); err != nil {
		return nil, err
	}

	return &Redis{client: client}, nil
}

func (c *Redis) StoreRate(ctx context.Context, from models.Currency, to models.Currency, value float64, expiration time.Duration) error {
	res := c.client.Set(ctx, string(from)+"/"+string(to), value, expiration)
	return res.Err()
}

func (c *Redis) StoreRates(ctx context.Context, rates map[models.Currency]float64, base models.Currency, expiration time.Duration) error {

	key := "rates:" + string(base)
	data := make(map[string]interface{})
	for currency, value := range rates {
		data[string(currency)] = value
	}

	err := c.client.HSet(ctx, key, data).Err()
	if err != nil {
		return err
	}

	return c.client.Expire(ctx, key, expiration).Err()
}

func (c *Redis) GetRate(ctx context.Context, from models.Currency, to models.Currency) (float64, error) {
	res, err := c.client.Get(ctx, string(from)+"/"+string(to)).Float64()

	if errors.Is(err, redis.Nil) {
		return 0, errs.KeyNotExists
	}
	return res, nil
}

func (c *Redis) GetRates(ctx context.Context, base models.Currency) (map[models.Currency]float64, error) {

	res := c.client.HGetAll(ctx, "rates:"+string(base))
	if res.Err() != nil {
		return nil, res.Err()
	}

	if len(res.Val()) == 0 {
		return nil, errs.KeyNotExists
	}

	rates := make(map[models.Currency]float64)
	for currency, rateStr := range res.Val() {
		rate, _ := strconv.ParseFloat(rateStr, 64)
		rates[models.Currency(currency)] = rate
	}

	return rates, nil
}

func (c *Redis) Close(ctx context.Context) error {

	done := make(chan error, 1)
	go func() {
		done <- c.client.Close()
		close(done)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("timeout while closing redis client: %w", ctx.Err())
	}
}
