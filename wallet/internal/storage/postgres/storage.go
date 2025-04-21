package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	errs "test-task/wallet/internal/domain/errors"
	"test-task/wallet/internal/domain/models"
)

type executor interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (p *Storage) AddUser(ctx context.Context, name string, password []byte, email string) error {

	_, err := p.pool.Exec(ctx, "INSERT INTO users (name, password, email) values ($1, $2, $3)",
		name, password, email)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return errs.UserAlreadyExists
			}
		}
	}

	return err
}

func (p *Storage) GetUserByName(ctx context.Context, name string) (*models.User, error) {

	user := models.User{}
	err := p.pool.QueryRow(ctx, "SELECT id, name, password, email FROM users WHERE name = $1",
		name).Scan(&user.ID, &user.Name, &user.Password, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.UserNotExists
		}
		return nil, fmt.Errorf("failed to check user existance in DB: %w", err)
	}

	return &user, nil
}

func (p *Storage) ExchangeAccountAmount(ctx context.Context, userID string, from models.Currency,
	fromAmount float64, to models.Currency, toAmount float64) error {

	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = p.changeAccountAmount(ctx, tx, userID, from, -fromAmount)
	if err != nil {
		return err
	}

	err = p.changeAccountAmount(ctx, tx, userID, to, toAmount)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *Storage) ExchangeAccountAmountWithBalance(ctx context.Context, userID string, from models.Currency,
	fromAmount float64, to models.Currency, toAmount float64) (map[models.Currency]float64, error) {

	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = p.changeAccountAmount(ctx, tx, userID, from, -fromAmount)
	if err != nil {
		return nil, err
	}

	err = p.changeAccountAmount(ctx, tx, userID, to, toAmount)
	if err != nil {
		return nil, err
	}

	balance, err := p.getBalance(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	return balance, tx.Commit(ctx)
}

func (p *Storage) ChangeAccountAmount(ctx context.Context, userID string, currency models.Currency, delta float64) error {
	return p.changeAccountAmount(ctx, p.pool, userID, currency, delta)
}

func (p *Storage) ChangeAccountAmountWithBalance(ctx context.Context, userID string,
	currency models.Currency, delta float64) (map[models.Currency]float64, error) {

	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = p.changeAccountAmount(ctx, tx, userID, currency, delta)
	if err != nil {
		return nil, err
	}

	balance, err := p.getBalance(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	return balance, tx.Commit(ctx)
}

func (p *Storage) GetBalance(ctx context.Context, userID string) (map[models.Currency]float64, error) {
	return p.getBalance(ctx, p.pool, userID)
}

func (p *Storage) getBalance(ctx context.Context, executor executor, userID string) (map[models.Currency]float64, error) {

	res := make(map[models.Currency]float64)

	query := "SELECT currency, amount from accounts WHERE user_id = $1"
	rows, err := executor.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var currency string
		var amount float64

		err = rows.Scan(&currency, &amount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		res[models.Currency(currency)] = amount
	}

	if err = rows.Err(); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to read rows during get balance: %w", err)
	}
	return res, nil
}

func (p *Storage) changeAccountAmount(ctx context.Context, executor executor, userID string, currency models.Currency, delta float64) error {

	query := "UPDATE accounts SET amount = amount + $3 WHERE user_id = $1 AND currency = $2"
	if delta > 0 {
		query = "INSERT INTO accounts (user_id, currency, amount) VALUES ($1, $2, $3) ON CONFLICT (user_id, currency) DO UPDATE SET amount = accounts.amount + $3;"
	}

	res, err := executor.Exec(ctx, query, userID, string(currency), delta)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			return errs.InsufficientFunds
		}
		return fmt.Errorf("failed to change amount in DB: %w", err)
	}

	if res.RowsAffected() == 0 {
		return errs.InsufficientFunds
	}
	return nil
}
