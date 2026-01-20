package domain

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type CurrencyMismatchError struct {
	first  Currency
	second Currency
}

func NewCurrencyMismatchError(first Currency, second Currency) *CurrencyMismatchError {
	return &CurrencyMismatchError{first: first, second: second}
}

func (err CurrencyMismatchError) Error() string {
	return fmt.Sprintf("%s and %s are not equal currencies", err.first, err.second)
}

type NegativeTransferError struct {
	money Money
}

func NewNegativeTransferError(money Money) *NegativeTransferError {
	return &NegativeTransferError{money: money}
}

func (err NegativeTransferError) Error() string {
	return fmt.Sprintf(
		"cannot transfer a negative amount: %s",
		err.money.amount.String(),
	)
}

type UnsupportedCurrencyError struct {
	currency Currency
}

func NewUnsupportedCurrencyError(currency Currency) *UnsupportedCurrencyError {
	return &UnsupportedCurrencyError{currency: currency}
}

func (err UnsupportedCurrencyError) Error() string {
	return fmt.Sprintf("unsupported currency %s", err.currency)
}

type AccountNotFoundError struct {
	AccountID AccountID
}

func NewAccountNotFoundError(accountID AccountID) *AccountNotFoundError {
	return &AccountNotFoundError{AccountID: accountID}
}

func (err AccountNotFoundError) Error() string {
	return fmt.Sprintf("account %v not found", err.AccountID)
}

type AccountBalanceMismatchError struct {
	AccountID      AccountID
	AccountBalance decimal.Decimal
	LedgerBalance  decimal.Decimal
}

func NewAccountBalanceMismatchError(accountID AccountID, accountBalance, ledgerBalance decimal.Decimal) *AccountBalanceMismatchError {
	return &AccountBalanceMismatchError{
		AccountID:      accountID,
		AccountBalance: accountBalance,
		LedgerBalance:  ledgerBalance,
	}
}

func (err AccountBalanceMismatchError) Error() string {
	return fmt.Sprintf(
		"account %v balance mismatch: account has %s, ledger has %s",
		err.AccountID, err.AccountBalance.String(), err.LedgerBalance.String(),
	)
}

type LedgerImbalanceError struct {
	Currency Currency
	Sum      decimal.Decimal
}

func NewLedgerImbalanceError(currency Currency, sum decimal.Decimal) *LedgerImbalanceError {
	return &LedgerImbalanceError{Currency: currency, Sum: sum}
}

func (err LedgerImbalanceError) Error() string {
	return fmt.Sprintf("ledger is not balanced for %s: sum is %s, expected 0", err.Currency, err.Sum.String())
}

type NegativeExchangeError struct {
	money Money
}

func NewNegativeExchangeError(money Money) *NegativeExchangeError {
	return &NegativeExchangeError{money: money}
}

func (err NegativeExchangeError) Error() string {
	return fmt.Sprintf(
		"cannot exchange a negative amount: %s",
		err.money.amount.String(),
	)
}

type SameCurrencyExchangeError struct {
	currency Currency
}

type SameCurrencyExchangeRateError struct {
	currency Currency
}

func NewSameCurrencyExchangeRateError(currency Currency) *SameCurrencyExchangeRateError {
	return &SameCurrencyExchangeRateError{currency: currency}
}

func (err SameCurrencyExchangeRateError) Error() string {
	return fmt.Sprintf("exchange rate cannot have the same source and target currency: %s", err.currency)
}

type InvalidExchangeRateError struct {
	rate decimal.Decimal
}

func NewInvalidExchangeRateError(rate decimal.Decimal) *InvalidExchangeRateError {
	return &InvalidExchangeRateError{rate: rate}
}

func (err InvalidExchangeRateError) Error() string {
	return fmt.Sprintf("exchange rate must be positive, got: %s", err.rate.String())
}

func NewSameCurrencyExchangeError(currency Currency) *SameCurrencyExchangeError {
	return &SameCurrencyExchangeError{currency: currency}
}

func (err SameCurrencyExchangeError) Error() string {
	return fmt.Sprintf("cannot exchange within the same currency: %s", err.currency)
}

type UserNotFoundError struct {
	Email string
}

func NewUserNotFoundError(email string) *UserNotFoundError {
	return &UserNotFoundError{Email: email}
}

func (err UserNotFoundError) Error() string {
	return fmt.Sprintf("user with email %s not found", err.Email)
}

type InvalidCredentialsError struct{}

func NewInvalidCredentialsError() *InvalidCredentialsError {
	return &InvalidCredentialsError{}
}

func (err InvalidCredentialsError) Error() string {
	return "invalid credentials"
}

type UserAlreadyExistsError struct {
	Email string
}

func NewUserAlreadyExistsError(email string) *UserAlreadyExistsError {
	return &UserAlreadyExistsError{Email: email}
}

type InsufficientFundsError struct {
	AccountID        AccountID
	RequestedAmount  decimal.Decimal
	AvailableBalance decimal.Decimal
}

func NewInsufficientFundsError(accountID AccountID, requestedAmount, availableBalance decimal.Decimal) *InsufficientFundsError {
	return &InsufficientFundsError{
		AccountID:        accountID,
		RequestedAmount:  requestedAmount,
		AvailableBalance: availableBalance,
	}
}

func (err InsufficientFundsError) Error() string {
	return fmt.Sprintf(
		"insufficient funds in account %v: requested %s, available %s",
		err.AccountID, err.RequestedAmount.String(), err.AvailableBalance.String(),
	)
}

func (err UserAlreadyExistsError) Error() string {
	return fmt.Sprintf("user with email %s already exists", err.Email)
}
