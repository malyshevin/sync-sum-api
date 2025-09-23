package repository

import (
	"context"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// CounterStore is a concrete type encapsulating counter persistence.
type CounterStore struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewCounterStore(pool *pgxpool.Pool, logger *slog.Logger) *CounterStore {
	v := validator.New()
	if err := v.Var(pool, "required"); err != nil {
		panic("CounterStore: pool is required")
	}
	return &CounterStore{pool: pool, logger: logger}
}

// Increment atomically increments the counter by 1 and returns the new value.
// It uses a transaction with SELECT ... FOR UPDATE to avoid lost updates.
func (r *CounterStore) Increment(ctx context.Context) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, errors.Wrap(err, "begin tx")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var current int64
	if err := tx.QueryRow(ctx, `SELECT value FROM counters WHERE id = 1 FOR UPDATE`).Scan(&current); err != nil {
		return 0, errors.Wrap(err, "select for update")
	}

	newVal := current + 1
	if _, err := tx.Exec(ctx, `UPDATE counters SET value = $1 WHERE id = 1`, newVal); err != nil {
		return 0, errors.Wrap(err, "update counter")
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, errors.Wrap(err, "commit tx")
	}
	return newVal, nil
}

func (r *CounterStore) Get(ctx context.Context) (int64, error) {
	var v int64
	if err := r.pool.QueryRow(ctx, `SELECT value FROM counters WHERE id = 1`).Scan(&v); err != nil {
		return 0, errors.Wrap(err, "select value")
	}
	return v, nil
}
