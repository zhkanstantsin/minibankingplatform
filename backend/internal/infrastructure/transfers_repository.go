package infrastructure

import (
	"context"
	"fmt"
	"minibankingplatform/internal/domain"
	"minibankingplatform/pkg/trm"

	"github.com/google/uuid"
)

type TransfersRepository struct {
	injector *trm.Injector[DBTX]
}

func NewTransfersRepository(injector *trm.Injector[DBTX]) *TransfersRepository {
	return &TransfersRepository{injector: injector}
}

func (tr *TransfersRepository) Insert(ctx context.Context, transfer *domain.TransferDetails) error {
	// TODO: it's better to have nested transaction here,
	//  but pgx factory doesn't support it for now
	if !tr.injector.HasContextTransaction(ctx) {
		return fmt.Errorf("insert command must be called inside of running transaction")
	}

	err := tr.insertTransaction(ctx, transfer)
	if err != nil {
		return fmt.Errorf("inserting transaction %w", err)
	}

	err = tr.insertDetails(ctx, transfer)
	if err != nil {
		return fmt.Errorf("inserting details %w", err)
	}

	ledgerEntry, err := transfer.GetLedgerEntry()
	if err != nil {
		return fmt.Errorf("getting ledger entry %w", err)
	}

	err = tr.insertLedgerEntry(ctx, ledgerEntry)
	if err != nil {
		return fmt.Errorf("inserting ledger entry %w", err)
	}

	return nil
}

func (tr *TransfersRepository) insertTransaction(ctx context.Context, transfer *domain.TransferDetails) error {
	const query = `
		INSERT INTO transactions (id, type, account_id, timestamp)
		VALUES ($1, $2, $3, $4)
	`

	_, err := tr.injector.DB(ctx).Exec(ctx, query,
		uuid.UUID(transfer.TransactionID()),
		domain.TransactionTypeTransfer,
		uuid.UUID(transfer.Sender()),
		transfer.Time(),
	)
	if err != nil {
		return fmt.Errorf("executing query: %w", err)
	}

	return nil
}

func (tr *TransfersRepository) insertDetails(ctx context.Context, transfer *domain.TransferDetails) error {
	const query = `
		INSERT INTO transfer_details (id, transaction_id, recipient_account_id, amount, currency)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := tr.injector.DB(ctx).Exec(
		ctx,
		query,
		uuid.UUID(transfer.ID()),
		uuid.UUID(transfer.TransactionID()),
		uuid.UUID(transfer.Recipient()),
		transfer.Money().Amount(),
		transfer.Money().Currency(),
	)
	if err != nil {
		return fmt.Errorf("executing query: %w", err)
	}

	return nil
}

func (tr *TransfersRepository) insertLedgerEntry(ctx context.Context, ledgerEntry domain.LedgerEntry) error {
	first := ledgerEntry[0]
	second := ledgerEntry[1]

	err := tr.insertLedgerRecord(ctx, first)
	if err != nil {
		return fmt.Errorf("inserting ledger entry: %w", err)
	}

	err = tr.insertLedgerRecord(ctx, second)
	if err != nil {
		return fmt.Errorf("inserting ledger entry: %w", err)
	}

	return nil
}

func (tr *TransfersRepository) insertLedgerRecord(ctx context.Context, ledgerRecord *domain.LedgerRecord) error {
	const query = `
		INSERT INTO ledger (id, transaction, account, amount, currency, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := tr.injector.DB(ctx).Exec(ctx, query,
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
