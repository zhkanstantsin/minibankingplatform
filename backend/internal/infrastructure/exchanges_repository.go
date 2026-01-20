package infrastructure

import (
	"context"
	"fmt"
	"minibankingplatform/internal/domain"
	"minibankingplatform/pkg/trm"

	"github.com/google/uuid"
)

type ExchangesRepository struct {
	injector *trm.Injector[DBTX]
}

func NewExchangesRepository(injector *trm.Injector[DBTX]) *ExchangesRepository {
	return &ExchangesRepository{injector: injector}
}

func (er *ExchangesRepository) Insert(ctx context.Context, exchange *domain.ExchangeDetails) error {
	// TODO: it's better to have nested transaction here,
	//  but pgx factory doesn't support it for now
	if !er.injector.HasContextTransaction(ctx) {
		return fmt.Errorf("insert command must be called inside of running transaction")
	}

	err := er.insertTransaction(ctx, exchange)
	if err != nil {
		return fmt.Errorf("inserting transaction %w", err)
	}

	err = er.insertDetails(ctx, exchange)
	if err != nil {
		return fmt.Errorf("inserting details %w", err)
	}

	ledgerEntries, err := exchange.GetLedgerEntries()
	if err != nil {
		return fmt.Errorf("getting ledger entries: %w", err)
	}

	err = er.insertLedgerEntries(ctx, ledgerEntries)
	if err != nil {
		return fmt.Errorf("inserting ledger entries: %w", err)
	}

	return nil
}

func (er *ExchangesRepository) insertTransaction(ctx context.Context, exchange *domain.ExchangeDetails) error {
	const query = `
		INSERT INTO transactions (id, type, account_id, timestamp)
		VALUES ($1, $2, $3, $4)
	`

	_, err := er.injector.DB(ctx).Exec(ctx, query,
		uuid.UUID(exchange.TransactionID()),
		domain.TransactionTypeExchange,
		uuid.UUID(exchange.SourceAccount()),
		exchange.Time(),
	)
	if err != nil {
		return fmt.Errorf("executing query: %w", err)
	}

	return nil
}

func (er *ExchangesRepository) insertDetails(ctx context.Context, exchange *domain.ExchangeDetails) error {
	const query = `
		INSERT INTO exchange_details (
			id,
			transaction_id,
			source_account_id,
			target_account_id,
			source_amount,
			source_currency,
			target_amount,
			target_currency,
			exchange_rate
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := er.injector.DB(ctx).Exec(
		ctx,
		query,
		uuid.UUID(exchange.ID()),
		uuid.UUID(exchange.TransactionID()),
		uuid.UUID(exchange.SourceAccount()),
		uuid.UUID(exchange.TargetAccount()),
		exchange.SourceAmount().Amount(),
		exchange.SourceAmount().Currency(),
		exchange.TargetAmount().Amount(),
		exchange.TargetAmount().Currency(),
		exchange.ExchangeRate(),
	)
	if err != nil {
		return fmt.Errorf("executing query: %w", err)
	}

	return nil
}

func (er *ExchangesRepository) insertLedgerEntries(ctx context.Context, entries domain.ExchangeLedgerEntries) error {
	for i, record := range entries.Records() {
		err := er.insertLedgerRecord(ctx, record)
		if err != nil {
			return fmt.Errorf("inserting ledger record %d: %w", i+1, err)
		}
	}

	return nil
}

func (er *ExchangesRepository) insertLedgerRecord(ctx context.Context, ledgerRecord *domain.LedgerRecord) error {
	const query = `
		INSERT INTO ledger (id, transaction, account, amount, currency, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := er.injector.DB(ctx).Exec(ctx, query,
		uuid.UUID(ledgerRecord.ID()),
		uuid.UUID(ledgerRecord.Transaction()),
		uuid.UUID(ledgerRecord.Account()),
		ledgerRecord.Money().Amount(),
		ledgerRecord.Money().Currency(),
		ledgerRecord.Time(),
	)
	if err != nil {
		return fmt.Errorf("executing query: %w", err)
	}

	return nil
}
