package repository

import (
	"context"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

func OpenPostgres(ctx context.Context, dsn string, logger *slog.Logger) (*pgxpool.Pool, error) {
	v := validator.New()
	if err := v.Var(dsn, "required"); err != nil {
		return nil, errors.New("postgres dsn is required")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "parse config")
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "create pool")
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.Wrap(err, "ping")
	}
	return pool, nil
}
