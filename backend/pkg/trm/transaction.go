package trm

type Transaction[T any] interface {
	// Raw returns the real transaction sql.Tx, sqlx.Tx or another.
	Raw() T
	// Commit the trm.Transaction.
	// Commit should be used only inside Manager.
	Commit() error
	// Rollback the trm.Transaction.
	// Rollback should be used only inside Manager.
	Rollback() error
}
