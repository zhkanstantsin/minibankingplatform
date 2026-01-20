package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ExchangeService struct{}

func (es *ExchangeService) Execute(
	sourceAccount *Account,
	targetAccount *Account,
	cashbookUSD *Account,
	cashbookEUR *Account,
	sourceAmount Money,
	exchangeRate ExchangeRate,
	now time.Time,
) (*ExchangeDetails, error) {
	if sourceAmount.IsNegative() {
		return nil, NewNegativeExchangeError(sourceAmount)
	}

	if sourceAmount.IsZero() {
		return nil, fmt.Errorf("exchange amount cannot be zero")
	}

	if sourceAccount.Balance().Currency() == targetAccount.Balance().Currency() {
		return nil, NewSameCurrencyExchangeError(sourceAccount.Balance().Currency())
	}

	if exchangeRate.From() != sourceAmount.Currency() {
		return nil, NewCurrencyMismatchError(exchangeRate.From(), sourceAmount.Currency())
	}
	if exchangeRate.To() != targetAccount.Balance().Currency() {
		return nil, NewCurrencyMismatchError(exchangeRate.To(), targetAccount.Balance().Currency())
	}

	targetAmount, err := CalculateExchangeAmount(sourceAmount, exchangeRate)
	if err != nil {
		return nil, fmt.Errorf("cannot calculate exchange amount: %w", err)
	}

	if err := sourceAccount.Debit(sourceAmount); err != nil {
		return nil, fmt.Errorf("cannot debit from source account %s: %w", sourceAccount.ID(), err)
	}

	if err := targetAccount.Credit(targetAmount); err != nil {
		return nil, fmt.Errorf("cannot credit to target account %s: %w", targetAccount.ID(), err)
	}

	sourceCashbook := getCashbookForCurrency(cashbookUSD, cashbookEUR, sourceAmount.Currency())
	if err := sourceCashbook.Credit(sourceAmount); err != nil {
		return nil, fmt.Errorf("cannot credit source cashbook: %w", err)
	}

	targetCashbook := getCashbookForCurrency(cashbookUSD, cashbookEUR, targetAmount.Currency())
	if err := targetCashbook.Debit(targetAmount); err != nil {
		return nil, fmt.Errorf("cannot debit target cashbook: %w", err)
	}

	exchange, err := NewExchangeDetails(
		NewExchangeDetailsID(),
		sourceAccount.ID(),
		targetAccount.ID(),
		sourceAmount,
		targetAmount,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create exchange details: %w", err)
	}

	return exchange, nil
}

func getCashbookForCurrency(cashbookUSD, cashbookEUR *Account, currency Currency) *Account {
	if currency == CurrencyUSD {
		return cashbookUSD
	}
	return cashbookEUR
}

func CalculateExchangeAmount(sourceAmount Money, exchangeRate ExchangeRate) (Money, error) {
	return exchangeRate.Convert(sourceAmount)
}

type ExchangeDetailsID uuid.UUID

func NewExchangeDetailsID() ExchangeDetailsID {
	return ExchangeDetailsID(uuid.New())
}

type ExchangeDetails struct {
	id            ExchangeDetailsID
	transaction   *Transaction
	sourceAccount AccountID
	targetAccount AccountID
	sourceAmount  Money
	targetAmount  Money
	time          time.Time
}

func NewExchangeDetails(
	id ExchangeDetailsID,
	sourceAccount AccountID,
	targetAccount AccountID,
	sourceAmount Money,
	targetAmount Money,
	time time.Time,
) (*ExchangeDetails, error) {
	if sourceAmount.Currency() == targetAmount.Currency() {
		return nil, fmt.Errorf("source and target currencies must be different")
	}

	return &ExchangeDetails{
		id:            id,
		transaction:   NewTransaction(NewTransactionID(), TransactionTypeExchange, sourceAccount, time),
		sourceAccount: sourceAccount,
		targetAccount: targetAccount,
		sourceAmount:  sourceAmount,
		targetAmount:  targetAmount,
		time:          time,
	}, nil
}

func (ed *ExchangeDetails) ID() ExchangeDetailsID {
	return ed.id
}

func (ed *ExchangeDetails) TransactionID() TransactionID {
	return ed.transaction.ID()
}

func (ed *ExchangeDetails) SourceAccount() AccountID {
	return ed.sourceAccount
}

func (ed *ExchangeDetails) TargetAccount() AccountID {
	return ed.targetAccount
}

func (ed *ExchangeDetails) SourceAmount() Money {
	return ed.sourceAmount
}

func (ed *ExchangeDetails) TargetAmount() Money {
	return ed.targetAmount
}

func (ed *ExchangeDetails) Time() time.Time {
	return ed.time
}

func (ed *ExchangeDetails) ExchangeRate() decimal.Decimal {
	return ed.targetAmount.Amount().Div(ed.sourceAmount.Amount())
}

func (ed *ExchangeDetails) GetLedgerEntries() (ExchangeLedgerEntries, error) {
	sourceCashbook := GetCashbookAccount(ed.SourceAmount().Currency())
	targetCashbook := GetCashbookAccount(ed.TargetAmount().Currency())

	sourceCurrencyEntry, err := ed.buildSourceCurrencyEntry(sourceCashbook)
	if err != nil {
		return ExchangeLedgerEntries{}, fmt.Errorf("building source currency entry: %w", err)
	}

	targetCurrencyEntry, err := ed.buildTargetCurrencyEntry(targetCashbook)
	if err != nil {
		return ExchangeLedgerEntries{}, fmt.Errorf("building target currency entry: %w", err)
	}

	return ExchangeLedgerEntries{
		SourceCurrencyEntry: sourceCurrencyEntry,
		TargetCurrencyEntry: targetCurrencyEntry,
	}, nil
}

func (ed *ExchangeDetails) buildSourceCurrencyEntry(cashbook AccountID) (LedgerEntry, error) {
	userDebit := NewLedgerRecord(
		NewLedgerRecordID(),
		ed.TransactionID(),
		ed.SourceAccount(),
		ed.SourceAmount().ToNegative(),
		ed.Time(),
	)

	cashbookCredit := NewLedgerRecord(
		NewLedgerRecordID(),
		ed.TransactionID(),
		cashbook,
		ed.SourceAmount(),
		ed.Time(),
	)

	if err := validateBalancedEntry(userDebit, cashbookCredit); err != nil {
		return LedgerEntry{}, fmt.Errorf("source currency: %w", err)
	}

	return LedgerEntry{userDebit, cashbookCredit}, nil
}

func (ed *ExchangeDetails) buildTargetCurrencyEntry(cashbook AccountID) (LedgerEntry, error) {
	cashbookDebit := NewLedgerRecord(
		NewLedgerRecordID(),
		ed.TransactionID(),
		cashbook,
		ed.TargetAmount().ToNegative(),
		ed.Time(),
	)

	userCredit := NewLedgerRecord(
		NewLedgerRecordID(),
		ed.TransactionID(),
		ed.TargetAccount(),
		ed.TargetAmount(),
		ed.Time(),
	)

	if err := validateBalancedEntry(cashbookDebit, userCredit); err != nil {
		return LedgerEntry{}, fmt.Errorf("target currency: %w", err)
	}

	return LedgerEntry{cashbookDebit, userCredit}, nil
}

func validateBalancedEntry(a, b *LedgerRecord) error {
	sum, err := a.Money().Add(b.Money())
	if err != nil {
		return fmt.Errorf("cannot sum records: %w", err)
	}
	if !sum.IsZero() {
		return fmt.Errorf("ledger entry does not balance (sum=%s)", sum.Amount())
	}
	return nil
}
