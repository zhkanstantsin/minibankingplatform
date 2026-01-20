package domain

import "github.com/shopspring/decimal"

type ExchangeRate struct {
	from Currency
	to   Currency
	rate decimal.Decimal
}

func NewExchangeRate(from, to Currency, rate decimal.Decimal) (ExchangeRate, error) {
	if from == to {
		return ExchangeRate{}, NewSameCurrencyExchangeRateError(from)
	}

	if !from.IsValid() {
		return ExchangeRate{}, NewUnsupportedCurrencyError(from)
	}

	if !to.IsValid() {
		return ExchangeRate{}, NewUnsupportedCurrencyError(to)
	}

	if rate.IsNegative() || rate.IsZero() {
		return ExchangeRate{}, NewInvalidExchangeRateError(rate)
	}

	return ExchangeRate{
		from: from,
		to:   to,
		rate: rate,
	}, nil
}

func (e ExchangeRate) From() Currency {
	return e.from
}

func (e ExchangeRate) To() Currency {
	return e.to
}

func (e ExchangeRate) Rate() decimal.Decimal {
	return e.rate
}

func (e ExchangeRate) Convert(amount Money) (Money, error) {
	if amount.Currency() != e.from {
		return Money{}, NewCurrencyMismatchError(e.from, amount.Currency())
	}

	convertedAmount := amount.Amount().Mul(e.rate).Round(2)

	return NewMoney(convertedAmount, e.to)
}

type ExchangeRateProvider interface {
	GetRate(from Currency, to Currency) (ExchangeRate, error)
}

type ExchangeRateNotFoundError struct {
	From Currency
	To   Currency
}

func NewExchangeRateNotFoundError(from, to Currency) *ExchangeRateNotFoundError {
	return &ExchangeRateNotFoundError{From: from, To: to}
}

func (e *ExchangeRateNotFoundError) Error() string {
	return "exchange rate not found for " + string(e.From) + " to " + string(e.To)
}
