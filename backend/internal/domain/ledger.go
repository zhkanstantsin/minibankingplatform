package domain

import (
	"time"

	"github.com/google/uuid"
)

type Ledger struct {
	cashBookAccounts map[Currency]AccountID
}

type LedgerRecordID uuid.UUID

func NewLedgerRecordID() LedgerRecordID {
	return LedgerRecordID(uuid.New())
}

type LedgerRecord struct {
	id          LedgerRecordID
	transaction TransactionID
	account     AccountID
	money       Money
	time        time.Time
}

func NewLedgerRecord(id LedgerRecordID, transaction TransactionID, account AccountID, money Money, time time.Time) *LedgerRecord {
	return &LedgerRecord{
		id:          id,
		transaction: transaction,
		account:     account,
		money:       money,
		time:        time,
	}
}

func (l LedgerRecord) ID() LedgerRecordID {
	return l.id
}

func (l LedgerRecord) Transaction() TransactionID {
	return l.transaction
}

func (l LedgerRecord) Account() AccountID {
	return l.account
}

func (l LedgerRecord) Money() Money {
	return l.money
}

func (l LedgerRecord) Time() time.Time {
	return l.time
}

type LedgerEntry [2]*LedgerRecord

type ExchangeLedgerEntries struct {
	SourceCurrencyEntry LedgerEntry
	TargetCurrencyEntry LedgerEntry
}

func (e ExchangeLedgerEntries) Records() []*LedgerRecord {
	return []*LedgerRecord{
		e.SourceCurrencyEntry[0],
		e.SourceCurrencyEntry[1],
		e.TargetCurrencyEntry[0],
		e.TargetCurrencyEntry[1],
	}
}
