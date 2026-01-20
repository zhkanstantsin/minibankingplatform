package domain

//go:generate go tool go-enum --marshal --names --values

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ENUM(transfer, exchange, deposit, withdrawal)
type TransactionType string

type TransactionID uuid.UUID

func NewTransactionID() TransactionID {
	return TransactionID(uuid.New())
}

type Transaction struct {
	id              TransactionID
	transactionType TransactionType
	account         AccountID
	time            time.Time
}

func NewTransaction(id TransactionID, transactionType TransactionType, account AccountID, time time.Time) *Transaction {
	return &Transaction{
		id:              id,
		transactionType: transactionType,
		account:         account,
		time:            time,
	}
}

func (t *Transaction) Type() TransactionType {
	return t.transactionType
}

func (t *Transaction) ID() TransactionID {
	return t.id
}

func (t *Transaction) Account() AccountID {
	return t.account
}

func (t *Transaction) Time() time.Time {
	return t.time
}

type TransferDetailsView struct {
	id               uuid.UUID
	recipientAccount AccountID
	amount           Money
}

func NewTransferDetailsView(id uuid.UUID, recipientAccount AccountID, amount Money) *TransferDetailsView {
	return &TransferDetailsView{
		id:               id,
		recipientAccount: recipientAccount,
		amount:           amount,
	}
}

func (v *TransferDetailsView) ID() uuid.UUID {
	return v.id
}

func (v *TransferDetailsView) RecipientAccount() AccountID {
	return v.recipientAccount
}

func (v *TransferDetailsView) Amount() Money {
	return v.amount
}

type ExchangeDetailsView struct {
	id            uuid.UUID
	sourceAccount AccountID
	targetAccount AccountID
	sourceAmount  Money
	targetAmount  Money
	exchangeRate  decimal.Decimal
}

func NewExchangeDetailsView(
	id uuid.UUID,
	sourceAccount AccountID,
	targetAccount AccountID,
	sourceAmount Money,
	targetAmount Money,
	exchangeRate decimal.Decimal,
) *ExchangeDetailsView {
	return &ExchangeDetailsView{
		id:            id,
		sourceAccount: sourceAccount,
		targetAccount: targetAccount,
		sourceAmount:  sourceAmount,
		targetAmount:  targetAmount,
		exchangeRate:  exchangeRate,
	}
}

func (v *ExchangeDetailsView) ID() uuid.UUID {
	return v.id
}

func (v *ExchangeDetailsView) SourceAccount() AccountID {
	return v.sourceAccount
}

func (v *ExchangeDetailsView) TargetAccount() AccountID {
	return v.targetAccount
}

func (v *ExchangeDetailsView) SourceAmount() Money {
	return v.sourceAmount
}

func (v *ExchangeDetailsView) TargetAmount() Money {
	return v.targetAmount
}

func (v *ExchangeDetailsView) ExchangeRate() decimal.Decimal {
	return v.exchangeRate
}

type TransactionWithDetails struct {
	transaction     *Transaction
	transferDetails *TransferDetailsView
	exchangeDetails *ExchangeDetailsView
}

func NewTransactionWithDetails(
	transaction *Transaction,
	transferDetails *TransferDetailsView,
	exchangeDetails *ExchangeDetailsView,
) *TransactionWithDetails {
	return &TransactionWithDetails{
		transaction:     transaction,
		transferDetails: transferDetails,
		exchangeDetails: exchangeDetails,
	}
}

func (t *TransactionWithDetails) Transaction() *Transaction {
	return t.transaction
}

func (t *TransactionWithDetails) TransferDetails() *TransferDetailsView {
	return t.transferDetails
}

func (t *TransactionWithDetails) ExchangeDetails() *ExchangeDetailsView {
	return t.exchangeDetails
}
