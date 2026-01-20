package service_test

import (
	"context"
	"testing"
	"time"

	"minibankingplatform/internal/infrastructure"
	"minibankingplatform/internal/service"
	jwtpkg "minibankingplatform/pkg/jwt"
	"minibankingplatform/pkg/trm"
	"minibankingplatform/pkg/trm/pgxfactory"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupService creates a new Service instance with real repositories.
func setupService(t *testing.T, pool *pgxpool.Pool) *service.Service {
	t.Helper()

	ctx := context.Background()

	factory, err := pgxfactory.New(ctx, pool)
	require.NoError(t, err)

	transactionManager := trm.NewTransactionManager(factory)
	injector := trm.NewInjector[infrastructure.DBTX](pool)

	usersRepo := infrastructure.NewUsersRepository(injector)
	accountsRepo := infrastructure.NewAccountsRepository(injector)
	transfersRepo := infrastructure.NewTransfersRepository(injector)
	exchangesRepo := infrastructure.NewExchangesRepository(injector)
	transactionsRepo := infrastructure.NewTransactionsRepository(injector)
	ledgerRepo := infrastructure.NewLedgerRepository(injector)

	// Create fixed exchange rate provider: 1 USD = 0.92 EUR
	exchangeRateProvider := infrastructure.NewFixedExchangeRateProvider(decimal.NewFromFloat(0.92))

	// Create token manager for JWT
	tokenManager := jwtpkg.NewTokenManager("test-secret-key", time.Hour)

	return service.NewService(transactionManager, usersRepo, accountsRepo, transfersRepo, exchangesRepo, transactionsRepo, ledgerRepo, exchangeRateProvider, tokenManager)
}

// TestUserAccounts holds user info and account IDs created during registration.
type TestUserAccounts struct {
	UserID       uuid.UUID
	Email        string
	USDAccountID uuid.UUID
	EURAccountID uuid.UUID
}

// registerTestUser registers a new user via service and returns user info with account IDs.
// The user gets 1000 USD and 500 EUR initial balance from cashbook.
func registerTestUser(ctx context.Context, t *testing.T, svc *service.Service, pool *pgxpool.Pool) *TestUserAccounts {
	t.Helper()

	email := uuid.New().String() + "@test.com"
	result, err := svc.Register(ctx, &service.RegisterCommand{
		Email:    email,
		Password: "testpassword123",
	})
	require.NoError(t, err)

	// Get account IDs for the registered user
	var usdAccountID, eurAccountID uuid.UUID
	err = pool.QueryRow(ctx, `SELECT id FROM accounts WHERE user_id = $1 AND currency = 'USD'`, result.UserID).Scan(&usdAccountID)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, `SELECT id FROM accounts WHERE user_id = $1 AND currency = 'EUR'`, result.UserID).Scan(&eurAccountID)
	require.NoError(t, err)

	return &TestUserAccounts{
		UserID:       result.UserID,
		Email:        email,
		USDAccountID: usdAccountID,
		EURAccountID: eurAccountID,
	}
}

// getAccountBalance retrieves the current balance of an account.
func getAccountBalance(ctx context.Context, t *testing.T, pool *pgxpool.Pool, accountID uuid.UUID) decimal.Decimal {
	t.Helper()

	var balance decimal.Decimal
	err := pool.QueryRow(ctx, `SELECT balance FROM accounts WHERE id = $1`, accountID).Scan(&balance)
	require.NoError(t, err)

	return balance
}

// countLedgerRecords counts ledger records for a given account.
func countLedgerRecords(ctx context.Context, t *testing.T, pool *pgxpool.Pool, accountID uuid.UUID) int {
	t.Helper()

	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM ledger WHERE account = $1`, accountID).Scan(&count)
	require.NoError(t, err)

	return count
}

// assertLedgerBalanced verifies:
// 1. The ledger is balanced for each currency separately (sum = 0)
// 2. All account balances match their ledger sums.
func assertLedgerBalanced(ctx context.Context, t *testing.T, svc *service.Service) {
	t.Helper()

	err := svc.CheckLedgerBalanceByCurrency(ctx)
	assert.NoError(t, err, "ledger should be balanced for each currency")

	err = svc.CheckAllAccountBalances(ctx)
	assert.NoError(t, err, "all account balances should match ledger sums")
}

// getAccountBalanceOrZero returns account balance or zero if account doesn't exist.
func getAccountBalanceOrZero(ctx context.Context, t *testing.T, pool *pgxpool.Pool, accountID uuid.UUID) decimal.Decimal {
	t.Helper()

	var balance decimal.Decimal
	err := pool.QueryRow(ctx, `SELECT balance FROM accounts WHERE id = $1`, accountID).Scan(&balance)
	if err != nil {
		return decimal.Zero
	}
	return balance
}

// assertBalanceEquals checks if balance equals expected value.
func assertBalanceEquals(t *testing.T, ctx context.Context, pool *pgxpool.Pool, accountID uuid.UUID, expected decimal.Decimal, msgAndArgs ...any) {
	t.Helper()
	balance := getAccountBalanceOrZero(ctx, t, pool, accountID)
	assert.True(t, balance.Equal(expected), append([]any{"expected %s, got %s"}, expected, balance, msgAndArgs)...)
}
