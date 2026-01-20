package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type AccountID uuid.UUID
type UserID uuid.UUID
type Version uint64

type Account struct {
	id      AccountID
	userID  UserID
	balance Money
}

func NewAccount(id AccountID, userID UserID, balance Money) *Account {
	return &Account{
		id:      id,
		userID:  userID,
		balance: balance,
	}
}

func (a *Account) ID() AccountID {
	return a.id
}

func (a *Account) UserID() UserID {
	return a.userID
}

func (a *Account) Balance() Money {
	return a.balance
}

func (a *Account) IsCashbook() bool {
	return a.userID == CashbookUserID
}

func (a *Account) Credit(money Money) error {
	updated, err := a.balance.Add(money)
	if err != nil {
		return fmt.Errorf("money adding money to account: %w", err)
	}

	a.balance = updated

	return nil
}

func (a *Account) Debit(money Money) error {
	if !a.IsCashbook() && a.balance.Amount().LessThan(money.Amount()) {
		return NewInsufficientFundsError(a.id, money.Amount(), a.balance.Amount())
	}

	updated, err := a.balance.Sub(money)
	if err != nil {
		return fmt.Errorf("money adding money to account: %w", err)
	}

	a.balance = updated

	return nil
}
