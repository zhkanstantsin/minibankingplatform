package trm

import (
	"context"
	"fmt"
)

type TransactionFactory[Tx any, Opts any] func(ctx context.Context, opts Opts) (Transaction[Tx], error)

type TransactionManager[Tx any, Opts any] struct {
	factory  TransactionFactory[Tx, Opts]
	injector Injector[Tx]
}

func NewTransactionManager[Tx any, Opts any](
	factory TransactionFactory[Tx, Opts],
) *TransactionManager[Tx, Opts] {
	return &TransactionManager[Tx, Opts]{
		factory: factory,
	}
}

// Do invoke function in transaction with zero-value options, that means to use default settings.
func (trm *TransactionManager[Tx, Opts]) Do(ctx context.Context, fn func(context.Context) error) error {
	var zero Opts

	return trm.DoTx(ctx, zero, fn)
}

// DoTx invoke function in transaction.
// It accepts options, so you can use it to pass you transaction options.
func (trm *TransactionManager[Tx, Opts]) DoTx(ctx context.Context, opts Opts, fn func(context.Context) error) error {
	tx, err := trm.factory(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	ctx = withTx(ctx, tx.Raw())

	if err = fn(ctx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("failed to rollback transaction: %w", rerr)
		}

		return err
	}

	if err = tx.Commit(); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("failed to rollback transaction: %w", rerr)
		}

		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
