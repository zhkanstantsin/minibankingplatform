package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TransferService struct{}

func (ts *TransferService) Execute(
	from *Account,
	to *Account,
	money Money,
	now time.Time,
) (*TransferDetails, error) {
	if money.IsNegative() {
		return nil, NewNegativeTransferError(money)
	}

	if err := from.Debit(money); err != nil {
		return nil, fmt.Errorf("cannot debit from %s: %w", from.ID(), err)
	}

	if err := to.Credit(money); err != nil {
		return nil, fmt.Errorf("cannot credit to %s: %w", to.ID(), err)
	}

	transfer, err := NewTransferDetails(NewTransferDetailsID(), from.ID(), to.ID(), money, now)
	if err != nil {
		return nil, fmt.Errorf("cannot create transfer details: %w", err)
	}

	return transfer, nil
}

type TransferDetailsID uuid.UUID

func NewTransferDetailsID() TransferDetailsID {
	return TransferDetailsID(uuid.New())
}

type TransferDetails struct {
	id          TransferDetailsID
	transaction *Transaction
	recipient   AccountID
	money       Money
	time        time.Time
}

func NewTransferDetails(id TransferDetailsID, from AccountID, to AccountID, money Money, time time.Time) (*TransferDetails, error) {
	return &TransferDetails{
		id:          id,
		transaction: NewTransaction(NewTransactionID(), TransactionTypeTransfer, from, time),
		recipient:   to,
		money:       money,
		time:        time,
	}, nil
}

func (td *TransferDetails) ID() TransferDetailsID {
	return td.id
}

func (td *TransferDetails) TransactionID() TransactionID {
	return td.transaction.ID()
}

func (td *TransferDetails) Sender() AccountID {
	return td.transaction.Account()
}

func (td *TransferDetails) Recipient() AccountID {
	return td.recipient
}

func (td *TransferDetails) Money() Money {
	return td.money
}

func (td *TransferDetails) Time() time.Time {
	return td.time
}

func (td *TransferDetails) GetLedgerEntry() (LedgerEntry, error) {
	first := NewLedgerRecord(NewLedgerRecordID(), td.TransactionID(), td.Sender(), td.Money().ToNegative(), td.Time())
	second := NewLedgerRecord(NewLedgerRecordID(), td.TransactionID(), td.Recipient(), td.Money(), td.Time())

	sum, err := first.Money().Add(second.Money())
	if err != nil {
		return LedgerEntry{}, fmt.Errorf("cannot add money to first: %w", err)
	}

	if !sum.IsZero() {
		return LedgerEntry{}, fmt.Errorf("sum of two ledger records at the same transaction is not zero")
	}

	return LedgerEntry{first, second}, nil
}
