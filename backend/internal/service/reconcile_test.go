package service_test

import (
	"context"
	"testing"
	"time"

	"minibankingplatform/internal/domain"
	"minibankingplatform/internal/service"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReconcile_ConsistentSystem(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange: register two users (each gets 1000 USD and 500 EUR)
	user1 := registerTestUser(ctx, t, svc, testPool)
	user2 := registerTestUser(ctx, t, svc, testPool)

	// Perform a transfer
	transferAmount, _ := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	err := svc.Transfer(ctx, &service.TransferCommand{
		From:  domain.AccountID(user1.USDAccountID),
		To:    domain.AccountID(user2.USDAccountID),
		Money: transferAmount,
		Time:  time.Now(),
	})
	require.NoError(t, err)

	// Act
	report, err := svc.Reconcile(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.True(t, report.IsConsistent, "system should be consistent after valid transfers")
	assert.NotZero(t, report.Timestamp)
	assert.Empty(t, report.AccountMismatches, "should have no account mismatches")
	assert.GreaterOrEqual(t, report.TotalAccountsChecked, 4, "should have checked at least the test accounts (2 users * 2 accounts)")

	// Verify ledger balances - should all be zero
	for _, balance := range report.LedgerBalances {
		assert.True(t, balance.IsBalanced, "currency %s should be balanced", balance.Currency)
		assert.True(t, balance.TotalSum.IsZero(), "currency %s total should be zero", balance.Currency)
	}

	t.Logf("%+v", report)
}

func TestReconcile_EmptySystem(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Act - reconcile on a system with only cashbook accounts
	report, err := svc.Reconcile(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.True(t, report.IsConsistent, "empty system should be consistent")
	assert.Empty(t, report.AccountMismatches)
}

func TestReconcile_AfterExchange(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange: register user (gets 1000 USD and 500 EUR)
	user := registerTestUser(ctx, t, svc, testPool)

	// Perform an exchange
	exchangeAmount, _ := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	err := svc.Exchange(ctx, &service.ExchangeCommand{
		SourceAccount: domain.AccountID(user.USDAccountID),
		TargetAccount: domain.AccountID(user.EURAccountID),
		SourceAmount:  exchangeAmount,
		Time:          time.Now(),
	})
	require.NoError(t, err)

	// Act
	report, err := svc.Reconcile(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.True(t, report.IsConsistent, "system should be consistent after valid exchange")
	assert.Empty(t, report.AccountMismatches)

	// Both USD and EUR ledgers should be balanced
	for _, balance := range report.LedgerBalances {
		assert.True(t, balance.IsBalanced, "currency %s should be balanced after exchange", balance.Currency)
	}
}

func TestReconcile_MultipleTransfers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange: register three users (each gets 1000 USD and 500 EUR)
	user1 := registerTestUser(ctx, t, svc, testPool)
	user2 := registerTestUser(ctx, t, svc, testPool)
	user3 := registerTestUser(ctx, t, svc, testPool)

	// Chain of transfers
	transfers := []struct {
		from   domain.AccountID
		to     domain.AccountID
		amount int64
	}{
		{domain.AccountID(user1.USDAccountID), domain.AccountID(user2.USDAccountID), 100},
		{domain.AccountID(user2.USDAccountID), domain.AccountID(user3.USDAccountID), 200},
		{domain.AccountID(user3.USDAccountID), domain.AccountID(user1.USDAccountID), 50},
		{domain.AccountID(user1.USDAccountID), domain.AccountID(user3.USDAccountID), 75},
	}

	for _, tr := range transfers {
		money, _ := domain.NewMoney(decimal.NewFromInt(tr.amount), domain.CurrencyUSD)
		err := svc.Transfer(ctx, &service.TransferCommand{
			From:  tr.from,
			To:    tr.to,
			Money: money,
			Time:  time.Now(),
		})
		require.NoError(t, err)
	}

	// Act
	report, err := svc.Reconcile(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.True(t, report.IsConsistent, "system should remain consistent after multiple transfers")
	assert.Empty(t, report.AccountMismatches)
	assert.GreaterOrEqual(t, report.TotalAccountsChecked, 6, "should have checked at least 3 users * 2 accounts")
}

func TestReconcile_ReportContainsTimestamp(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	beforeReconcile := time.Now()

	// Act
	report, err := svc.Reconcile(ctx)

	afterReconcile := time.Now()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.True(t, report.Timestamp.After(beforeReconcile) || report.Timestamp.Equal(beforeReconcile),
		"timestamp should be after or equal to before time")
	assert.True(t, report.Timestamp.Before(afterReconcile) || report.Timestamp.Equal(afterReconcile),
		"timestamp should be before or equal to after time")
}
