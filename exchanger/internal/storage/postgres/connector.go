package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/stdlib"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Connector struct {
	Pool *pgxpool.Pool
}

func NewConnector(ctx context.Context, dsn string) (*Connector, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(timeoutCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err = pool.Ping(timeoutCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &Connector{Pool: pool}, nil
}

func (p *Connector) DB() *sql.DB {
	return stdlib.OpenDBFromPool(p.Pool)
}

func (p *Connector) Close(ctx context.Context) error {

	done := make(chan struct{})
	go func() {
		p.Pool.Close()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout while closing Postgres pool: %w", ctx.Err())
	}
}
