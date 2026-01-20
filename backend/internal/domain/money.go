package domain

//go:generate go tool go-enum --marshal --names --values

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// ENUM(USD, EUR)
type Currency string

type Money struct {
	amount   decimal.Decimal
	currency Currency
}

func NewMoney(amount decimal.Decimal, currency Currency) (Money, error) {
	if !currency.IsValid() {
		return Money{}, NewUnsupportedCurrencyError(currency)
	}

	return Money{
		amount:   amount,
		currency: currency,
	}, nil
}

func (m Money) CheckIsNotEqualCurrencies(other Money) error {
	if m.currency != other.currency {
		return NewCurrencyMismatchError(m.currency, other.currency)
	}

	return nil
}

func (m Money) LessThan(other Money) (bool, error) {
	if err := m.CheckIsNotEqualCurrencies(other); err != nil {
		return false, fmt.Errorf("money in different currencies is not comparable: %w", err)
	}

	return m.amount.LessThan(other.amount), nil
}

func (m Money) Add(other Money) (Money, error) {
	if err := m.CheckIsNotEqualCurrencies(other); err != nil {
		return Money{}, fmt.Errorf("money in different currencies cannot be added: %w", err)
	}

	return Money{
		currency: m.currency,
		amount:   m.amount.Add(other.amount),
	}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if err := m.CheckIsNotEqualCurrencies(other); err != nil {
		return Money{}, fmt.Errorf("money in different currencies cannot be subbed: %w", err)
	}

	return Money{
		currency: m.currency,
		amount:   m.amount.Sub(other.amount),
	}, nil
}

func (m Money) ToNegative() Money {
	return Money{
		currency: m.currency,
		amount:   m.amount.Neg(),
	}
}

func (m Money) IsNegative() bool {
	return m.amount.IsNegative()
}

func (m Money) Currency() Currency {
	return m.currency
}

func (m Money) Amount() decimal.Decimal {
	return m.amount
}

func (m Money) IsZero() bool {
	return m.amount.IsZero()
}
