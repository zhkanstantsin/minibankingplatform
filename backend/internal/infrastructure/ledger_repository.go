package infrastructure

import (
	"context"
	"fmt"
	"minibankingplatform/internal/domain"
	"minibankingplatform/pkg/trm"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LedgerRepository struct {
	injector *trm.Injector[DBTX]
}

func NewLedgerRepository(injector *trm.Injector[DBTX]) *LedgerRepository {
	return &LedgerRepository{injector: injector}
}

func (lr *LedgerRepository) GetTotalBalanceByCurrency(ctx context.Context) (map[domain.Currency]domain.Money, error) {
	const query = `SELECT currency, COALESCE(SUM(amount), 0) FROM ledger GROUP BY currency`

	rows, err := lr.injector.DB(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying ledger totals by currency: %w", err)
	}
	defer rows.Close()

	totals := make(map[domain.Currency]domain.Money)
	for rows.Next() {
		var currency domain.Currency
		var amount decimal.Decimal
		if err := rows.Scan(&currency, &amount); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		money, err := domain.NewMoney(amount, currency)
		if err != nil {
			return nil, fmt.Errorf("creating money for currency %s: %w", currency, err)
		}
		totals[currency] = money
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return totals, nil
}

func (lr *LedgerRepository) GetAccountBalance(ctx context.Context, accountID domain.AccountID, currency domain.Currency) (domain.Money, error) {
	const query = `SELECT COALESCE(SUM(amount), 0) FROM ledger WHERE account = $1`

	var amount decimal.Decimal
	err := lr.injector.DB(ctx).QueryRow(ctx, query, uuid.UUID(accountID)).Scan(&amount)
	if err != nil {
		return domain.Money{}, fmt.Errorf("querying account ledger balance: %w", err)
	}

	money, err := domain.NewMoney(amount, currency)
	if err != nil {
		return domain.Money{}, fmt.Errorf("creating money: %w", err)
	}

	return money, nil
}

type AccountBalanceMismatch struct {
	AccountID      domain.AccountID
	AccountBalance decimal.Decimal
	LedgerBalance  decimal.Decimal
	Currency       domain.Currency
}

func (lr *LedgerRepository) GetAccountBalanceMismatches(ctx context.Context) ([]AccountBalanceMismatch, error) {
	const query = `
		SELECT 
			a.id,
			a.balance,
			COALESCE(l.ledger_sum, 0) as ledger_sum,
			a.currency
		FROM accounts a
		LEFT JOIN (
			SELECT account, SUM(amount) as ledger_sum
			FROM ledger
			GROUP BY account
		) l ON a.id = l.account
		WHERE a.balance != COALESCE(l.ledger_sum, 0)
	`

	rows, err := lr.injector.DB(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying account balance mismatches: %w", err)
	}
	defer rows.Close()

	var mismatches []AccountBalanceMismatch
	for rows.Next() {
		var m AccountBalanceMismatch
		if err := rows.Scan(&m.AccountID, &m.AccountBalance, &m.LedgerBalance, &m.Currency); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		mismatches = append(mismatches, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return mismatches, nil
}
