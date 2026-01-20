package trm_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"minibankingplatform/pkg/trm"
)

func TestTransactionManager_Do(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	factoryCalls := 0
	mockFactory := func(_ context.Context, _ any) (trm.Transaction[any], error) {
		factoryCalls++
		return MockTX{}, nil
	}

	t.Run("should every time retrieve new transaction from factory on Do", func(t *testing.T) {
		t.Parallel()

		sut := trm.NewTransactionManager(mockFactory)
		atomicAction := func(_ context.Context) error {
			return nil
		}

		require.NoError(t, sut.Do(ctx, atomicAction))
		require.NoError(t, sut.Do(ctx, atomicAction))

		assert.Equal(t, 2, factoryCalls)
	})
}

var _ trm.Transaction[any] = MockTX{}

type MockTX struct{}

func (MockTX) Raw() any {
	return nil
}

func (MockTX) Commit() error {
	return nil
}

func (MockTX) Rollback() error {
	return nil
}
