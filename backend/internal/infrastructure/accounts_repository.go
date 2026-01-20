package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"minibankingplatform/internal/domain"
	"minibankingplatform/pkg/trm"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type AccountsRepository struct {
	injector *trm.Injector[DBTX]
}

func NewAccountsRepository(injector *trm.Injector[DBTX]) *AccountsRepository {
	return &AccountsRepository{
		injector: injector,
	}
}

func (ar *AccountsRepository) GetForUpdate(ctx context.Context, accountID domain.AccountID) (*domain.Account, error) {
	const query = `
		SELECT
		    id,
		    user_id,
		    balance,
		    currency
		FROM accounts
		WHERE id = $1
		FOR UPDATE
	`

	var (
		id       uuid.UUID
		userID   uuid.UUID
		amount   decimal.Decimal
		currency string
	)

	err := ar.injector.DB(ctx).QueryRow(ctx, query, uuid.UUID(accountID)).Scan(&id, &userID, &amount, &currency)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewAccountNotFoundError(accountID)
		}
		return nil, fmt.Errorf("querying account: %w", err)
	}

	balance, err := domain.NewMoney(amount, domain.Currency(currency))
	if err != nil {
		return nil, fmt.Errorf("creating money: %w", err)
	}

	return domain.NewAccount(domain.AccountID(id), domain.UserID(userID), balance), nil
}

func (ar *AccountsRepository) Save(ctx context.Context, account *domain.Account) error {
	const query = `
		INSERT INTO accounts (id, user_id, balance, currency)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET 
		    balance = EXCLUDED.balance,
		    currency = EXCLUDED.currency
	`

	_, err := ar.injector.DB(ctx).Exec(
		ctx,
		query,
		uuid.UUID(account.ID()),
		uuid.UUID(account.UserID()),
		account.Balance().Amount(),
		account.Balance().Currency(),
	)
	if err != nil {
		return fmt.Errorf("upserting account: %w", err)
	}

	return nil
}

func (ar *AccountsRepository) Count(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM accounts`

	var count int
	err := ar.injector.DB(ctx).QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting accounts: %w", err)
	}

	return count, nil
}

func (ar *AccountsRepository) GetByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Account, error) {
	const query = `
		SELECT 
		    id,
		    user_id,
		    balance,
		    currency
		FROM accounts
		WHERE user_id = $1
	`

	rows, err := ar.injector.DB(ctx).Query(ctx, query, uuid.UUID(userID))
	if err != nil {
		return nil, fmt.Errorf("querying accounts by user_id: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.Account
	for rows.Next() {
		var (
			id       uuid.UUID
			uid      uuid.UUID
			amount   decimal.Decimal
			currency string
		)

		if err := rows.Scan(&id, &uid, &amount, &currency); err != nil {
			return nil, fmt.Errorf("scanning account row: %w", err)
		}

		balance, err := domain.NewMoney(amount, domain.Currency(currency))
		if err != nil {
			return nil, fmt.Errorf("creating money: %w", err)
		}

		accounts = append(accounts, domain.NewAccount(domain.AccountID(id), domain.UserID(uid), balance))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating account rows: %w", err)
	}

	return accounts, nil
}

func (ar *AccountsRepository) Get(ctx context.Context, accountID domain.AccountID) (*domain.Account, error) {
	const query = `
		SELECT
		    id,
		    user_id,
		    balance,
		    currency
		FROM accounts
		WHERE id = $1
	`

	var (
		id       uuid.UUID
		userID   uuid.UUID
		amount   decimal.Decimal
		currency string
	)

	err := ar.injector.DB(ctx).QueryRow(ctx, query, uuid.UUID(accountID)).Scan(&id, &userID, &amount, &currency)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewAccountNotFoundError(accountID)
		}
		return nil, fmt.Errorf("querying account: %w", err)
	}

	balance, err := domain.NewMoney(amount, domain.Currency(currency))
	if err != nil {
		return nil, fmt.Errorf("creating money: %w", err)
	}

	return domain.NewAccount(domain.AccountID(id), domain.UserID(userID), balance), nil
}
