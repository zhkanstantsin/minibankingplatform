package infrastructure

import (
	"context"
	"fmt"
	"minibankingplatform/internal/domain"
	"minibankingplatform/pkg/trm"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionsFilter struct {
	UserID          domain.UserID
	TransactionType *domain.TransactionType
	Limit           int
	Offset          int
}

type TransactionsRepository struct {
	injector *trm.Injector[DBTX]
}

func NewTransactionsRepository(injector *trm.Injector[DBTX]) *TransactionsRepository {
	return &TransactionsRepository{injector: injector}
}

func (r *TransactionsRepository) GetList(ctx context.Context, filter TransactionsFilter) ([]*domain.TransactionWithDetails, error) {
	const query = `
		SELECT
			t.id, t.type, t.account_id, t.timestamp,
			td.id, td.recipient_account_id, td.amount, td.currency,
			ed.id, ed.source_account_id, ed.target_account_id,
			ed.source_amount, ed.source_currency,
			ed.target_amount, ed.target_currency, ed.exchange_rate
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		LEFT JOIN transfer_details td ON t.id = td.transaction_id AND t.type = 'transfer'
		LEFT JOIN accounts a_recipient ON td.recipient_account_id = a_recipient.id
		LEFT JOIN exchange_details ed ON t.id = ed.transaction_id AND t.type = 'exchange'
		LEFT JOIN accounts a_target ON ed.target_account_id = a_target.id
		WHERE ($1::transaction_type IS NULL OR t.type = $1)
		  AND (a.user_id = $4 OR a_recipient.user_id = $4 OR a_target.user_id = $4)
		ORDER BY t.timestamp DESC
		LIMIT $2 OFFSET $3
	`

	var typeArg any
	if filter.TransactionType != nil {
		typeArg = string(*filter.TransactionType)
	}

	rows, err := r.injector.DB(ctx).Query(ctx, query, typeArg, filter.Limit, filter.Offset, uuid.UUID(filter.UserID))
	if err != nil {
		return nil, fmt.Errorf("querying transactions: %w", err)
	}
	defer rows.Close()

	var result []*domain.TransactionWithDetails
	for rows.Next() {
		var (
			txID        uuid.UUID
			txType      string
			txAccountID uuid.UUID
			txTimestamp time.Time

			tdID          *uuid.UUID
			tdRecipientID *uuid.UUID
			tdAmount      *decimal.Decimal
			tdCurrency    *string

			edID             *uuid.UUID
			edSourceAccID    *uuid.UUID
			edTargetAccID    *uuid.UUID
			edSourceAmount   *decimal.Decimal
			edSourceCurrency *string
			edTargetAmount   *decimal.Decimal
			edTargetCurrency *string
			edExchangeRate   *decimal.Decimal
		)

		err := rows.Scan(
			&txID, &txType, &txAccountID, &txTimestamp,
			&tdID, &tdRecipientID, &tdAmount, &tdCurrency,
			&edID, &edSourceAccID, &edTargetAccID,
			&edSourceAmount, &edSourceCurrency,
			&edTargetAmount, &edTargetCurrency, &edExchangeRate,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning transaction row: %w", err)
		}

		transaction := domain.NewTransaction(
			domain.TransactionID(txID),
			domain.TransactionType(txType),
			domain.AccountID(txAccountID),
			txTimestamp,
		)

		var transferDetails *domain.TransferDetailsView
		var exchangeDetails *domain.ExchangeDetailsView

		if tdID != nil && tdRecipientID != nil && tdAmount != nil && tdCurrency != nil {
			amount, err := domain.NewMoney(*tdAmount, domain.Currency(*tdCurrency))
			if err != nil {
				return nil, fmt.Errorf("creating transfer money: %w", err)
			}
			transferDetails = domain.NewTransferDetailsView(
				*tdID,
				domain.AccountID(*tdRecipientID),
				amount,
			)
		}

		if edID != nil && edSourceAccID != nil && edTargetAccID != nil &&
			edSourceAmount != nil && edSourceCurrency != nil &&
			edTargetAmount != nil && edTargetCurrency != nil && edExchangeRate != nil {
			sourceAmount, err := domain.NewMoney(*edSourceAmount, domain.Currency(*edSourceCurrency))
			if err != nil {
				return nil, fmt.Errorf("creating exchange source money: %w", err)
			}
			targetAmount, err := domain.NewMoney(*edTargetAmount, domain.Currency(*edTargetCurrency))
			if err != nil {
				return nil, fmt.Errorf("creating exchange target money: %w", err)
			}
			exchangeDetails = domain.NewExchangeDetailsView(
				*edID,
				domain.AccountID(*edSourceAccID),
				domain.AccountID(*edTargetAccID),
				sourceAmount,
				targetAmount,
				*edExchangeRate,
			)
		}

		result = append(result, domain.NewTransactionWithDetails(
			transaction,
			transferDetails,
			exchangeDetails,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating transaction rows: %w", err)
	}

	return result, nil
}

func (r *TransactionsRepository) Count(ctx context.Context, filter TransactionsFilter) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		LEFT JOIN transfer_details td ON t.id = td.transaction_id AND t.type = 'transfer'
		LEFT JOIN accounts a_recipient ON td.recipient_account_id = a_recipient.id
		LEFT JOIN exchange_details ed ON t.id = ed.transaction_id AND t.type = 'exchange'
		LEFT JOIN accounts a_target ON ed.target_account_id = a_target.id
		WHERE ($1::transaction_type IS NULL OR t.type = $1)
		  AND (a.user_id = $2 OR a_recipient.user_id = $2 OR a_target.user_id = $2)
	`

	var typeArg any
	if filter.TransactionType != nil {
		typeArg = string(*filter.TransactionType)
	}

	var count int
	err := r.injector.DB(ctx).QueryRow(ctx, query, typeArg, uuid.UUID(filter.UserID)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting transactions: %w", err)
	}

	return count, nil
}
