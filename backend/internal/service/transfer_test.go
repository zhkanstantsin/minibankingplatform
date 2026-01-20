package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"minibankingplatform/internal/domain"
	"minibankingplatform/internal/service"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransfer_HappyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange - register two users (each gets 1000 USD, 500 EUR)
	fromUser := registerTestUser(ctx, t, svc, testPool)
	toUser := registerTestUser(ctx, t, svc, testPool)

	transferAmount, _ := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	cmd := &service.TransferCommand{
		From:  domain.AccountID(fromUser.USDAccountID),
		To:    domain.AccountID(toUser.USDAccountID),
		Money: transferAmount,
		Time:  time.Now(),
	}

	// Act
	err := svc.Transfer(ctx, cmd)

	// Assert
	require.NoError(t, err)

	assertBalanceEquals(t, ctx, testPool, fromUser.USDAccountID, decimal.NewFromInt(900))
	assertBalanceEquals(t, ctx, testPool, toUser.USDAccountID, decimal.NewFromInt(1100))

	// 1 ledger record from registration + 1 from transfer = 2
	assert.Equal(t, 2, countLedgerRecords(ctx, t, testPool, fromUser.USDAccountID))
	assert.Equal(t, 2, countLedgerRecords(ctx, t, testPool, toUser.USDAccountID))

	assertLedgerBalanced(ctx, t, svc)
}

func TestTransfer_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupAccounts    func(ctx context.Context, t *testing.T, svc *service.Service) (from, to uuid.UUID)
		transferAmount   decimal.Decimal
		transferCurrency domain.Currency
		checkError       func(t *testing.T, err error)
	}{
		{
			name: "account not found - sender",
			setupAccounts: func(ctx context.Context, t *testing.T, svc *service.Service) (from, to uuid.UUID) {
				user := registerTestUser(ctx, t, svc, testPool)
				return uuid.New(), user.USDAccountID // non-existent sender
			},
			transferAmount:   decimal.NewFromInt(100),
			transferCurrency: domain.CurrencyUSD,
			checkError: func(t *testing.T, err error) {
				var accountNotFoundErr *domain.AccountNotFoundError
				assert.ErrorAs(t, err, &accountNotFoundErr)
			},
		},
		{
			name: "account not found - recipient",
			setupAccounts: func(ctx context.Context, t *testing.T, svc *service.Service) (from, to uuid.UUID) {
				user := registerTestUser(ctx, t, svc, testPool)
				return user.USDAccountID, uuid.New() // non-existent recipient
			},
			transferAmount:   decimal.NewFromInt(100),
			transferCurrency: domain.CurrencyUSD,
			checkError: func(t *testing.T, err error) {
				var accountNotFoundErr *domain.AccountNotFoundError
				assert.ErrorAs(t, err, &accountNotFoundErr)
			},
		},
		{
			name: "currency mismatch",
			setupAccounts: func(ctx context.Context, t *testing.T, svc *service.Service) (from, to uuid.UUID) {
				user := registerTestUser(ctx, t, svc, testPool)
				// Try to transfer USD to EUR account
				return user.USDAccountID, user.EURAccountID
			},
			transferAmount:   decimal.NewFromInt(100),
			transferCurrency: domain.CurrencyUSD,
			checkError: func(t *testing.T, err error) {
				var currencyMismatchErr *domain.CurrencyMismatchError
				assert.ErrorAs(t, err, &currencyMismatchErr)
			},
		},
		{
			name: "negative amount",
			setupAccounts: func(ctx context.Context, t *testing.T, svc *service.Service) (from, to uuid.UUID) {
				fromUser := registerTestUser(ctx, t, svc, testPool)
				toUser := registerTestUser(ctx, t, svc, testPool)
				return fromUser.USDAccountID, toUser.USDAccountID
			},
			transferAmount:   decimal.NewFromInt(-50),
			transferCurrency: domain.CurrencyUSD,
			checkError: func(t *testing.T, err error) {
				var negativeTransferErr *domain.NegativeTransferError
				assert.ErrorAs(t, err, &negativeTransferErr)
			},
		},
		{
			name: "insufficient funds",
			setupAccounts: func(ctx context.Context, t *testing.T, svc *service.Service) (from, to uuid.UUID) {
				fromUser := registerTestUser(ctx, t, svc, testPool)
				toUser := registerTestUser(ctx, t, svc, testPool)
				return fromUser.USDAccountID, toUser.USDAccountID
			},
			transferAmount:   decimal.NewFromInt(2000), // User has only 1000 USD
			transferCurrency: domain.CurrencyUSD,
			checkError: func(t *testing.T, err error) {
				var insufficientFundsErr *domain.InsufficientFundsError
				assert.ErrorAs(t, err, &insufficientFundsErr)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			svc := setupService(t, testPool)

			// Capture initial balances before setup
			fromAccountID, toAccountID := tt.setupAccounts(ctx, t, svc)
			initialFromBalance := getAccountBalanceOrZero(ctx, t, testPool, fromAccountID)
			initialToBalance := getAccountBalanceOrZero(ctx, t, testPool, toAccountID)

			transferAmount, _ := domain.NewMoney(tt.transferAmount, tt.transferCurrency)
			cmd := &service.TransferCommand{
				From:  domain.AccountID(fromAccountID),
				To:    domain.AccountID(toAccountID),
				Money: transferAmount,
				Time:  time.Now(),
			}

			// Act
			err := svc.Transfer(ctx, cmd)

			// Assert
			require.Error(t, err)
			tt.checkError(t, err)

			// Verify no changes were made (transaction rolled back)
			assertBalanceEquals(t, ctx, testPool, fromAccountID, initialFromBalance)
			assertBalanceEquals(t, ctx, testPool, toAccountID, initialToBalance)

			// Ledger should still be balanced
			assertLedgerBalanced(ctx, t, svc)
		})
	}
}

func TestTransfer_ZeroAmount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	fromUser := registerTestUser(ctx, t, svc, testPool)
	toUser := registerTestUser(ctx, t, svc, testPool)

	transferAmount, _ := domain.NewMoney(decimal.Zero, domain.CurrencyUSD)
	cmd := &service.TransferCommand{
		From:  domain.AccountID(fromUser.USDAccountID),
		To:    domain.AccountID(toUser.USDAccountID),
		Money: transferAmount,
		Time:  time.Now(),
	}

	// Act - zero amount transfer should fail due to DB constraint (amount > 0)
	err := svc.Transfer(ctx, cmd)

	// Assert
	require.Error(t, err)

	assertBalanceEquals(t, ctx, testPool, fromUser.USDAccountID, decimal.NewFromInt(1000))
	assertBalanceEquals(t, ctx, testPool, toUser.USDAccountID, decimal.NewFromInt(1000))

	assertLedgerBalanced(ctx, t, svc)
}

func TestTransfer_SelfTransfer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	user := registerTestUser(ctx, t, svc, testPool)

	transferAmount, _ := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	cmd := &service.TransferCommand{
		From:  domain.AccountID(user.USDAccountID),
		To:    domain.AccountID(user.USDAccountID),
		Money: transferAmount,
		Time:  time.Now(),
	}

	// Act - self transfer fails due to invariant violation
	// The implementation creates two separate Account objects in memory.
	// The "from" gets debited (900), the "to" gets credited (1100).
	// The invariant check detects this: ledger sum = 1000, but account balance would be wrong.
	err := svc.Transfer(ctx, cmd)

	// Assert
	require.Error(t, err)
	var balanceMismatchErr *domain.AccountBalanceMismatchError
	assert.ErrorAs(t, err, &balanceMismatchErr)

	assertBalanceEquals(t, ctx, testPool, user.USDAccountID, decimal.NewFromInt(1000))
	assertLedgerBalanced(ctx, t, svc)
}

func TestTransfer_InsufficientFunds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Register users - each gets 1000 USD
	fromUser := registerTestUser(ctx, t, svc, testPool)
	toUser := registerTestUser(ctx, t, svc, testPool)

	// Try to transfer more than the balance (1500 USD when balance is 1000 USD)
	transferAmount, _ := domain.NewMoney(decimal.NewFromInt(1500), domain.CurrencyUSD)
	cmd := &service.TransferCommand{
		From:  domain.AccountID(fromUser.USDAccountID),
		To:    domain.AccountID(toUser.USDAccountID),
		Money: transferAmount,
		Time:  time.Now(),
	}

	// Act - should fail due to insufficient funds
	err := svc.Transfer(ctx, cmd)

	// Assert
	require.Error(t, err)
	var insufficientFundsErr *domain.InsufficientFundsError
	assert.ErrorAs(t, err, &insufficientFundsErr)

	// Balances should remain unchanged (transaction rolled back)
	assertBalanceEquals(t, ctx, testPool, fromUser.USDAccountID, decimal.NewFromInt(1000))
	assertBalanceEquals(t, ctx, testPool, toUser.USDAccountID, decimal.NewFromInt(1000))

	assertLedgerBalanced(ctx, t, svc)
}

func TestTransfer_DecimalPrecision(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Register users - each gets 1000 USD, 500 EUR
	fromUser := registerTestUser(ctx, t, svc, testPool)
	toUser := registerTestUser(ctx, t, svc, testPool)

	initialFrom := decimal.NewFromInt(1000)
	initialTo := decimal.NewFromInt(1000)

	transferDecimal := decimal.NewFromFloat(0.0001)
	transferAmount, _ := domain.NewMoney(transferDecimal, domain.CurrencyUSD)
	cmd := &service.TransferCommand{
		From:  domain.AccountID(fromUser.USDAccountID),
		To:    domain.AccountID(toUser.USDAccountID),
		Money: transferAmount,
		Time:  time.Now(),
	}

	// Act
	err := svc.Transfer(ctx, cmd)

	// Assert
	require.NoError(t, err)

	assertBalanceEquals(t, ctx, testPool, fromUser.USDAccountID, initialFrom.Sub(transferDecimal))
	assertBalanceEquals(t, ctx, testPool, toUser.USDAccountID, initialTo.Add(transferDecimal))

	assertLedgerBalanced(ctx, t, svc)
}

func TestTransfer_Atomicity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Register user - gets USD and EUR accounts
	user := registerTestUser(ctx, t, svc, testPool)

	// Try to transfer USD to EUR account - this should fail
	initialFromBalance := getAccountBalance(ctx, t, testPool, user.USDAccountID)
	initialToBalance := getAccountBalance(ctx, t, testPool, user.EURAccountID)

	transferAmount, _ := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	cmd := &service.TransferCommand{
		From:  domain.AccountID(user.USDAccountID),
		To:    domain.AccountID(user.EURAccountID),
		Money: transferAmount,
		Time:  time.Now(),
	}

	// Act
	err := svc.Transfer(ctx, cmd)

	// Assert - verify atomicity: all data should remain unchanged
	require.Error(t, err)

	assertBalanceEquals(t, ctx, testPool, user.USDAccountID, initialFromBalance)
	assertBalanceEquals(t, ctx, testPool, user.EURAccountID, initialToBalance)

	assertLedgerBalanced(ctx, t, svc)
}

func TestTransfer_ConcurrentTransfers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Register two users - fromUser will send money to toUser
	fromUser := registerTestUser(ctx, t, svc, testPool)
	toUser := registerTestUser(ctx, t, svc, testPool)

	transferAmount, _ := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)

	// Act - run 10 concurrent transfers of 100 USD each
	const numTransfers = 10
	var wg sync.WaitGroup
	errors := make(chan error, numTransfers)

	for i := 0; i < numTransfers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := &service.TransferCommand{
				From:  domain.AccountID(fromUser.USDAccountID),
				To:    domain.AccountID(toUser.USDAccountID),
				Money: transferAmount,
				Time:  time.Now(),
			}
			if err := svc.Transfer(ctx, cmd); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Collect errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	// Assert - all transfers should succeed (FOR UPDATE prevents race conditions)
	// fromUser starts with 1000, transfers 10*100=1000, ends with 0
	// toUser starts with 1000, receives 10*100=1000, ends with 2000
	require.Empty(t, errs)

	assertBalanceEquals(t, ctx, testPool, fromUser.USDAccountID, decimal.NewFromInt(0))
	assertBalanceEquals(t, ctx, testPool, toUser.USDAccountID, decimal.NewFromInt(2000))

	// 1 ledger record from registration + 10 from transfers = 11
	assert.Equal(t, 1+numTransfers, countLedgerRecords(ctx, t, testPool, fromUser.USDAccountID))
	// 1 ledger record from registration + 10 from transfers = 11
	assert.Equal(t, 1+numTransfers, countLedgerRecords(ctx, t, testPool, toUser.USDAccountID))

	assertLedgerBalanced(ctx, t, svc)
}

func TestTransfer_MultipleSequentialTransfers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Register 3 users - each gets 1000 USD
	user1 := registerTestUser(ctx, t, svc, testPool)
	user2 := registerTestUser(ctx, t, svc, testPool)
	user3 := registerTestUser(ctx, t, svc, testPool)

	// Act - chain of transfers: user1 -> user2 -> user3
	transfer1, _ := domain.NewMoney(decimal.NewFromInt(500), domain.CurrencyUSD)
	err := svc.Transfer(ctx, &service.TransferCommand{
		From:  domain.AccountID(user1.USDAccountID),
		To:    domain.AccountID(user2.USDAccountID),
		Money: transfer1,
		Time:  time.Now(),
	})
	require.NoError(t, err)

	transfer2, _ := domain.NewMoney(decimal.NewFromInt(700), domain.CurrencyUSD)
	err = svc.Transfer(ctx, &service.TransferCommand{
		From:  domain.AccountID(user2.USDAccountID),
		To:    domain.AccountID(user3.USDAccountID),
		Money: transfer2,
		Time:  time.Now(),
	})
	require.NoError(t, err)

	// Assert
	// user1: 1000 - 500 = 500
	// user2: 1000 + 500 - 700 = 800
	// user3: 1000 + 700 = 1700
	assertBalanceEquals(t, ctx, testPool, user1.USDAccountID, decimal.NewFromInt(500))
	assertBalanceEquals(t, ctx, testPool, user2.USDAccountID, decimal.NewFromInt(800))
	assertBalanceEquals(t, ctx, testPool, user3.USDAccountID, decimal.NewFromInt(1700))

	assertLedgerBalanced(ctx, t, svc)
}

func TestNewTransferCommand(t *testing.T) {
	t.Parallel()

	from := uuid.New()
	to := uuid.New()
	now := time.Now()

	tests := []struct {
		name             string
		from             uuid.UUID
		to               uuid.UUID
		amount           string
		currency         string
		time             time.Time
		expectError      bool
		expectedErrorMsg string
		validate         func(t *testing.T, cmd *service.TransferCommand)
	}{
		{
			name:        "valid input",
			from:        from,
			to:          to,
			amount:      "100.50",
			currency:    "USD",
			time:        now,
			expectError: false,
			validate: func(t *testing.T, cmd *service.TransferCommand) {
				assert.Equal(t, domain.AccountID(from), cmd.From)
				assert.Equal(t, domain.AccountID(to), cmd.To)
				assert.True(t, cmd.Money.Amount().Equal(decimal.NewFromFloat(100.50)))
				assert.Equal(t, domain.CurrencyUSD, cmd.Money.Currency())
				assert.Equal(t, now, cmd.Time)
			},
		},
		{
			name:             "invalid amount - not a number",
			from:             from,
			to:               to,
			amount:           "not-a-number",
			currency:         "USD",
			time:             now,
			expectError:      true,
			expectedErrorMsg: "invalid amount",
		},
		{
			name:        "invalid currency",
			from:        from,
			to:          to,
			amount:      "100",
			currency:    "GBP",
			time:        now,
			expectError: true,
			validate: func(t *testing.T, cmd *service.TransferCommand) {
				// Verify it's specifically the ErrInvalidCurrency error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd, err := service.NewTransferCommand(tt.from, tt.to, tt.amount, tt.currency, tt.time)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, cmd)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
				if tt.currency == "GBP" {
					assert.ErrorIs(t, err, domain.ErrInvalidCurrency)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, cmd)
				if tt.validate != nil {
					tt.validate(t, cmd)
				}
			}
		})
	}
}
