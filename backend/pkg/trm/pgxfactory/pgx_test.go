package pgxfactory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"minibankingplatform/pkg/trm"
	"minibankingplatform/pkg/trm/pgxfactory"
)

func TestPGXTRM(t *testing.T) {
	t.Logf("connecting to %s\n", postgresURL)

	type Execer interface {
		Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, postgresURL)
	require.NoError(t, err)

	transactionFactory, err := pgxfactory.New(ctx, pool)
	require.NoError(t, err)

	transactionManager := trm.NewTransactionManager(transactionFactory)
	txInjector := trm.NewInjector[Execer](pool)

	t.Run("rollback on error", func(t *testing.T) {
		errFromTransaction := errors.New("some unexpected error")
		err := transactionManager.Do(ctx, func(ctx context.Context) error {
			schema := `
				CREATE TABLE test_users (
					id SERIAL PRIMARY KEY 
				)
			`

			_, err := txInjector.DB(ctx).Exec(ctx, schema)
			require.NoError(t, err)

			return errFromTransaction
		})

		assert.ErrorIs(t, err, errFromTransaction)

		_, err = pool.Exec(ctx, "SELECT * FROM test_users")
		assert.Error(t, err, "table should not exists")
	})

	t.Run("commit on success", func(t *testing.T) {
		err := transactionManager.Do(ctx, func(ctx context.Context) error {
			schema := `
				CREATE TABLE test_users (
					id SERIAL PRIMARY KEY 
				)
			`

			_, err := txInjector.DB(ctx).Exec(ctx, schema)
			require.NoError(t, err)

			return nil
		})

		assert.NoError(t, err)

		_, err = pool.Exec(ctx, "SELECT * FROM test_users")
		assert.NoError(t, err, "schema should be committed after successful transaction")
	})
}
