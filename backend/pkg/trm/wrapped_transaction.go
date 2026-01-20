package trm

func WrapTransaction[T any](
	raw T,
	commit func() error,
	rollback func() error,
) Transaction[T] {
	return &wrappedTransaction[T]{
		getRawTx: func() T {
			return raw
		},
		commit:   commit,
		rollback: rollback,
	}
}

type wrappedTransaction[T any] struct {
	getRawTx func() T
	commit   func() error
	rollback func() error
}

func (wtx *wrappedTransaction[T]) Raw() T {
	return wtx.getRawTx()
}

func (wtx *wrappedTransaction[T]) Commit() error {
	return wtx.commit()
}

func (wtx *wrappedTransaction[T]) Rollback() error {
	return wtx.rollback()
}
