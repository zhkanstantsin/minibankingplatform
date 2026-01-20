package pgxfactory

import (
	"context"
	"fmt"
	"minibankingplatform/pkg/trm"

	"github.com/jackc/pgx/v5"
)

type DB interface {
	Ping(ctx context.Context) error
	BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
}

func New(ctx context.Context, db DB) (trm.TransactionFactory[pgx.Tx, pgx.TxOptions], error) {
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	return func(ctx context.Context, opts pgx.TxOptions) (trm.Transaction[pgx.Tx], error) {
		// TODO: add nested transactions support using savepoints
		tx, err := db.BeginTx(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to begin new transaction: %w", err)
		}

		return trm.WrapTransaction[pgx.Tx](
			tx,
			injectContext(ctx, tx.Commit),
			injectContext(ctx, tx.Rollback),
		), nil
	}, nil
}

func injectContext(ctx context.Context, fn func(ctx context.Context) error) func() error {
	return func() error {
		return fn(ctx)
	}
}
