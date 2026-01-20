package trm

import "context"

// Injector is an entity, that you need to use to get API to database.
// It returns transaction from context or default value: DB connection without transaction.
type Injector[T any] struct {
	// db is a default connection/pool for sending queries.
	// it will be returned in case when there is no transaction in context
	db T
}

func NewInjector[T any](db T) *Injector[T] {
	return &Injector[T]{
		db: db,
	}
}

func (i *Injector[T]) DB(ctx context.Context) T {
	return ctxOr(ctx, i.db)
}

// HasContextTransaction returns true if there is a transaction in the context
func (i *Injector[T]) HasContextTransaction(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKey{}).(T)
	return ok
}

type ctxKey struct{}

func withTx[T any](ctx context.Context, tx T) context.Context {
	return context.WithValue(ctx, ctxKey{}, tx)
}

func ctxOr[T any](ctx context.Context, tx T) T {
	if tx, ok := ctx.Value(ctxKey{}).(T); ok {
		return tx
	}

	return tx
}
