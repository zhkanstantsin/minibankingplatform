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

func TestExchange_HappyPath_USDtoEUR(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange - user gets 1000 USD and 500 EUR on registration
	user := registerTestUser(ctx, t, svc, testPool)

	// Exchange 100 USD to EUR (should get 92 EUR)
	exchangeAmount, _ := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	cmd := &service.ExchangeCommand{
		SourceAccount: domain.AccountID(user.USDAccountID),
		TargetAccount: domain.AccountID(user.EURAccountID),
		SourceAmount:  exchangeAmount,
		Time:          time.Now(),
	}

	// Act
	err := svc.Exchange(ctx, cmd)

	// Assert
	require.NoError(t, err)

	// USD account should have 900 USD (1000 - 100)
	assertBalanceEquals(t, ctx, testPool, user.USDAccountID, decimal.NewFromInt(900))

	// EUR account should have 500 + 92 = 592 EUR
	expectedEUR := decimal.NewFromInt(500).Add(decimal.NewFromInt(100).Mul(decimal.NewFromFloat(0.92)))
	assertBalanceEquals(t, ctx, testPool, user.EURAccountID, expectedEUR)

	// Verify ledger is balanced
	assertLedgerBalanced(ctx, t, svc)
}

func TestExchange_HappyPath_EURtoUSD(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange - user gets 1000 USD and 500 EUR on registration
	user := registerTestUser(ctx, t, svc, testPool)

	// Exchange 92 EUR to USD
	// EUR->USD rate is 1/0.92 ≈ 1.086957
	// 92 * 1.086957 = 100.00 (rounded to 2 decimal places)
	exchangeAmount, _ := domain.NewMoney(decimal.NewFromInt(92), domain.CurrencyEUR)
	cmd := &service.ExchangeCommand{
		SourceAccount: domain.AccountID(user.EURAccountID),
		TargetAccount: domain.AccountID(user.USDAccountID),
		SourceAmount:  exchangeAmount,
		Time:          time.Now(),
	}

	// Act
	err := svc.Exchange(ctx, cmd)

	// Assert
	require.NoError(t, err)

	// EUR account should have 408 EUR (500 - 92)
	assertBalanceEquals(t, ctx, testPool, user.EURAccountID, decimal.NewFromInt(408))

	// USD account should have 1000 + 92 * (1/0.92) rounded to 2 decimal places
	inverseRate := decimal.NewFromInt(1).Div(decimal.NewFromFloat(0.92)).Round(6)
	expectedUSD := decimal.NewFromInt(1000).Add(decimal.NewFromInt(92).Mul(inverseRate).Round(2))
	assertBalanceEquals(t, ctx, testPool, user.USDAccountID, expectedUSD)

	// Verify ledger is balanced
	assertLedgerBalanced(ctx, t, svc)
}

func TestExchange_SameCurrencyError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange - register two users to get two USD accounts
	user1 := registerTestUser(ctx, t, svc, testPool)
	user2 := registerTestUser(ctx, t, svc, testPool)

	exchangeAmount, _ := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	cmd := &service.ExchangeCommand{
		SourceAccount: domain.AccountID(user1.USDAccountID),
		TargetAccount: domain.AccountID(user2.USDAccountID),
		SourceAmount:  exchangeAmount,
		Time:          time.Now(),
	}

	// Act
	err := svc.Exchange(ctx, cmd)

	// Assert
	require.Error(t, err)
	// Error can come from either ExchangeRateProvider (SameCurrencyExchangeRateError)
	// or from domain service (SameCurrencyExchangeError)
	var sameCurrencyRateErr *domain.SameCurrencyExchangeRateError
	var sameCurrencyExchangeErr *domain.SameCurrencyExchangeError
	assert.True(t, assert.ErrorAs(t, err, &sameCurrencyRateErr) || assert.ErrorAs(t, err, &sameCurrencyExchangeErr))

	// Balances should remain unchanged
	assertBalanceEquals(t, ctx, testPool, user1.USDAccountID, decimal.NewFromInt(1000))
	assertBalanceEquals(t, ctx, testPool, user2.USDAccountID, decimal.NewFromInt(1000))

	assertLedgerBalanced(ctx, t, svc)
}

func TestExchange_NegativeAmount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange - user gets 1000 USD and 500 EUR
	user := registerTestUser(ctx, t, svc, testPool)

	exchangeAmount, _ := domain.NewMoney(decimal.NewFromInt(-100), domain.CurrencyUSD)
	cmd := &service.ExchangeCommand{
		SourceAccount: domain.AccountID(user.USDAccountID),
		TargetAccount: domain.AccountID(user.EURAccountID),
		SourceAmount:  exchangeAmount,
		Time:          time.Now(),
	}

	// Act
	err := svc.Exchange(ctx, cmd)

	// Assert
	require.Error(t, err)
	var negativeExchangeErr *domain.NegativeExchangeError
	assert.ErrorAs(t, err, &negativeExchangeErr)

	// Balances should remain unchanged
	assertBalanceEquals(t, ctx, testPool, user.USDAccountID, decimal.NewFromInt(1000))
	assertBalanceEquals(t, ctx, testPool, user.EURAccountID, decimal.NewFromInt(500))

	assertLedgerBalanced(ctx, t, svc)
}

func TestExchange_ZeroAmount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange - user gets 1000 USD and 500 EUR
	user := registerTestUser(ctx, t, svc, testPool)

	exchangeAmount, _ := domain.NewMoney(decimal.Zero, domain.CurrencyUSD)
	cmd := &service.ExchangeCommand{
		SourceAccount: domain.AccountID(user.USDAccountID),
		TargetAccount: domain.AccountID(user.EURAccountID),
		SourceAmount:  exchangeAmount,
		Time:          time.Now(),
	}

	// Act
	err := svc.Exchange(ctx, cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exchange amount cannot be zero")

	// Balances should remain unchanged
	assertBalanceEquals(t, ctx, testPool, user.USDAccountID, decimal.NewFromInt(1000))
	assertBalanceEquals(t, ctx, testPool, user.EURAccountID, decimal.NewFromInt(500))

	assertLedgerBalanced(ctx, t, svc)
}

func TestExchange_InsufficientFunds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange - user gets 1000 USD and 500 EUR
	user := registerTestUser(ctx, t, svc, testPool)

	// Try to exchange 2000 USD (user only has 1000)
	exchangeAmount, _ := domain.NewMoney(decimal.NewFromInt(2000), domain.CurrencyUSD)
	cmd := &service.ExchangeCommand{
		SourceAccount: domain.AccountID(user.USDAccountID),
		TargetAccount: domain.AccountID(user.EURAccountID),
		SourceAmount:  exchangeAmount,
		Time:          time.Now(),
	}

	// Act
	err := svc.Exchange(ctx, cmd)

	// Assert
	require.Error(t, err)
	var insufficientFundsErr *domain.InsufficientFundsError
	assert.ErrorAs(t, err, &insufficientFundsErr)

	// Balances should remain unchanged
	assertBalanceEquals(t, ctx, testPool, user.USDAccountID, decimal.NewFromInt(1000))
	assertBalanceEquals(t, ctx, testPool, user.EURAccountID, decimal.NewFromInt(500))

	assertLedgerBalanced(ctx, t, svc)
}

func TestExchange_AccountNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange - user gets 1000 USD
	user := registerTestUser(ctx, t, svc, testPool)

	exchangeAmount, _ := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	cmd := &service.ExchangeCommand{
		SourceAccount: domain.AccountID(user.USDAccountID),
		TargetAccount: domain.AccountID(uuid.New()), // Non-existent account
		SourceAmount:  exchangeAmount,
		Time:          time.Now(),
	}

	// Act
	err := svc.Exchange(ctx, cmd)

	// Assert
	require.Error(t, err)
	var accountNotFoundErr *domain.AccountNotFoundError
	assert.ErrorAs(t, err, &accountNotFoundErr)

	// Balance should remain unchanged
	assertBalanceEquals(t, ctx, testPool, user.USDAccountID, decimal.NewFromInt(1000))

	assertLedgerBalanced(ctx, t, svc)
}

func TestExchange_DecimalPrecision(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange - user gets 1000 USD and 500 EUR
	user := registerTestUser(ctx, t, svc, testPool)

	initialUSD := decimal.NewFromInt(1000)
	initialEUR := decimal.NewFromInt(500)

	// Exchange 100.25 USD to EUR
	exchangeDecimal := decimal.NewFromFloat(100.25)
	exchangeAmount, _ := domain.NewMoney(exchangeDecimal, domain.CurrencyUSD)
	cmd := &service.ExchangeCommand{
		SourceAccount: domain.AccountID(user.USDAccountID),
		TargetAccount: domain.AccountID(user.EURAccountID),
		SourceAmount:  exchangeAmount,
		Time:          time.Now(),
	}

	// Act
	err := svc.Exchange(ctx, cmd)

	// Assert
	require.NoError(t, err)

	// USD account should have initial - exchange amount
	expectedUSD := initialUSD.Sub(exchangeDecimal)
	assertBalanceEquals(t, ctx, testPool, user.USDAccountID, expectedUSD)

	// EUR account should have initial + exchange amount * rate, rounded to 2 decimal places
	expectedEUR := initialEUR.Add(exchangeDecimal.Mul(decimal.NewFromFloat(0.92)).Round(2))
	assertBalanceEquals(t, ctx, testPool, user.EURAccountID, expectedEUR)

	assertLedgerBalanced(ctx, t, svc)
}

func TestExchange_Atomicity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange - user gets 1000 USD
	user := registerTestUser(ctx, t, svc, testPool)
	initialBalance := getAccountBalance(ctx, t, testPool, user.USDAccountID)

	exchangeAmount, _ := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	cmd := &service.ExchangeCommand{
		SourceAccount: domain.AccountID(user.USDAccountID),
		TargetAccount: domain.AccountID(uuid.New()), // Non-existent target
		SourceAmount:  exchangeAmount,
		Time:          time.Now(),
	}

	// Act
	err := svc.Exchange(ctx, cmd)

	// Assert - verify atomicity: all data should remain unchanged
	require.Error(t, err)

	assertBalanceEquals(t, ctx, testPool, user.USDAccountID, initialBalance)

	assertLedgerBalanced(ctx, t, svc)
}

func TestExchange_ConcurrentExchanges(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange - user gets 1000 USD and 500 EUR
	user := registerTestUser(ctx, t, svc, testPool)

	exchangeAmount, _ := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)

	// Act - run 10 concurrent exchanges of 100 USD each
	const numExchanges = 10
	var wg sync.WaitGroup
	errors := make(chan error, numExchanges)

	for i := 0; i < numExchanges; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := &service.ExchangeCommand{
				SourceAccount: domain.AccountID(user.USDAccountID),
				TargetAccount: domain.AccountID(user.EURAccountID),
				SourceAmount:  exchangeAmount,
				Time:          time.Now(),
			}
			if err := svc.Exchange(ctx, cmd); err != nil {
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

	// Assert - all exchanges should succeed
	require.Empty(t, errs)

	// USD account should have 0 (1000 - 10*100)
	assertBalanceEquals(t, ctx, testPool, user.USDAccountID, decimal.NewFromInt(0))

	// EUR account should have 500 + 920 = 1420 (initial 500 + 10 * 100 * 0.92)
	expectedEUR := decimal.NewFromInt(500).Add(decimal.NewFromInt(1000).Mul(decimal.NewFromFloat(0.92)))
	assertBalanceEquals(t, ctx, testPool, user.EURAccountID, expectedEUR)

	assertLedgerBalanced(ctx, t, svc)
}

func TestExchange_MultipleSequentialExchanges(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := setupService(t, testPool)

	// Arrange - user gets 1000 USD and 500 EUR
	user := registerTestUser(ctx, t, svc, testPool)

	// Act - Exchange USD -> EUR -> USD
	// 1. Exchange 500 USD to EUR (500 * 0.92 = 460 EUR)
	exchange1, _ := domain.NewMoney(decimal.NewFromInt(500), domain.CurrencyUSD)
	err := svc.Exchange(ctx, &service.ExchangeCommand{
		SourceAccount: domain.AccountID(user.USDAccountID),
		TargetAccount: domain.AccountID(user.EURAccountID),
		SourceAmount:  exchange1,
		Time:          time.Now(),
	})
	require.NoError(t, err)

	// 2. Exchange 200 EUR back to USD
	// EUR->USD rate is 1/0.92 ≈ 1.086957
	// 200 * 1.086957 = 217.39 (rounded to 2 decimal places)
	exchange2, _ := domain.NewMoney(decimal.NewFromInt(200), domain.CurrencyEUR)
	err = svc.Exchange(ctx, &service.ExchangeCommand{
		SourceAccount: domain.AccountID(user.EURAccountID),
		TargetAccount: domain.AccountID(user.USDAccountID),
		SourceAmount:  exchange2,
		Time:          time.Now(),
	})
	require.NoError(t, err)

	// Assert
	// USD: 1000 - 500 + (200 * inverse_rate) = 500 + 217.39 = 717.39
	inverseRate := decimal.NewFromInt(1).Div(decimal.NewFromFloat(0.92)).Round(6)
	eurToUsdConverted := decimal.NewFromInt(200).Mul(inverseRate).Round(2)
	expectedUSD := decimal.NewFromInt(500).Add(eurToUsdConverted)
	assertBalanceEquals(t, ctx, testPool, user.USDAccountID, expectedUSD)

	// EUR: 500 + (500*0.92) - 200 = 500 + 460 - 200 = 760
	expectedEUR := decimal.NewFromInt(500).Add(decimal.NewFromInt(500).Mul(decimal.NewFromFloat(0.92)).Round(2)).Sub(decimal.NewFromInt(200))
	assertBalanceEquals(t, ctx, testPool, user.EURAccountID, expectedEUR)

	assertLedgerBalanced(ctx, t, svc)
}

func TestNewExchangeCommand(t *testing.T) {
	t.Parallel()

	sourceAccount := uuid.New()
	targetAccount := uuid.New()
	now := time.Now()

	tests := []struct {
		name             string
		sourceAccount    uuid.UUID
		targetAccount    uuid.UUID
		amount           string
		currency         string
		time             time.Time
		expectError      bool
		expectedErrorMsg string
		validate         func(t *testing.T, cmd *service.ExchangeCommand)
	}{
		{
			name:          "valid input USD",
			sourceAccount: sourceAccount,
			targetAccount: targetAccount,
			amount:        "100.50",
			currency:      "USD",
			time:          now,
			expectError:   false,
			validate: func(t *testing.T, cmd *service.ExchangeCommand) {
				assert.Equal(t, domain.AccountID(sourceAccount), cmd.SourceAccount)
				assert.Equal(t, domain.AccountID(targetAccount), cmd.TargetAccount)
				assert.True(t, cmd.SourceAmount.Amount().Equal(decimal.NewFromFloat(100.50)))
				assert.Equal(t, domain.CurrencyUSD, cmd.SourceAmount.Currency())
				assert.Equal(t, now, cmd.Time)
			},
		},
		{
			name:          "valid input EUR",
			sourceAccount: sourceAccount,
			targetAccount: targetAccount,
			amount:        "92.00",
			currency:      "EUR",
			time:          now,
			expectError:   false,
			validate: func(t *testing.T, cmd *service.ExchangeCommand) {
				assert.Equal(t, domain.AccountID(sourceAccount), cmd.SourceAccount)
				assert.Equal(t, domain.AccountID(targetAccount), cmd.TargetAccount)
				assert.True(t, cmd.SourceAmount.Amount().Equal(decimal.NewFromFloat(92.00)))
				assert.Equal(t, domain.CurrencyEUR, cmd.SourceAmount.Currency())
				assert.Equal(t, now, cmd.Time)
			},
		},
		{
			name:             "invalid amount - not a number",
			sourceAccount:    sourceAccount,
			targetAccount:    targetAccount,
			amount:           "not-a-number",
			currency:         "USD",
			time:             now,
			expectError:      true,
			expectedErrorMsg: "invalid amount",
		},
		{
			name:          "invalid currency",
			sourceAccount: sourceAccount,
			targetAccount: targetAccount,
			amount:        "100",
			currency:      "GBP",
			time:          now,
			expectError:   true,
			validate: func(t *testing.T, cmd *service.ExchangeCommand) {
				// Verify it's specifically the ErrInvalidCurrency error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd, err := service.NewExchangeCommand(tt.sourceAccount, tt.targetAccount, tt.amount, tt.currency, tt.time)

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
