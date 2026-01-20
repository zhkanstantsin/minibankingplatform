package service

import (
	"context"
	"fmt"
	"minibankingplatform/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransferCommand struct {
	From  domain.AccountID
	To    domain.AccountID
	Money domain.Money
	Time  time.Time
}

func NewTransferCommand(
	from uuid.UUID,
	to uuid.UUID,
	amount string,
	rawCurrency string,
	time time.Time,
) (*TransferCommand, error) {
	decimalAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	currency, err := domain.ParseCurrency(rawCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	money, err := domain.NewMoney(decimalAmount, currency)
	if err != nil {
		return nil, fmt.Errorf("cannot get money value: %w", err)
	}

	return &TransferCommand{
		From:  domain.AccountID(from),
		To:    domain.AccountID(to),
		Money: money,
		Time:  time,
	}, nil
}

func (s *Service) Transfer(ctx context.Context, cmd *TransferCommand) error {
	err := s.trm.Do(ctx, func(ctx context.Context) error {
		from, err := s.accounts.GetForUpdate(ctx, cmd.From)
		if err != nil {
			return fmt.Errorf("getting 'from' account: %w", err)
		}

		to, err := s.accounts.GetForUpdate(ctx, cmd.To)
		if err != nil {
			return fmt.Errorf("getting 'to' account: %w", err)
		}

		details, err := s.transfer.Execute(from, to, cmd.Money, cmd.Time)
		if err != nil {
			return fmt.Errorf("executing transfer domain service: %w", err)
		}

		err = s.transfers.Insert(ctx, details)
		if err != nil {
			return fmt.Errorf("inserting transfer domain service: %w", err)
		}

		err = s.accounts.Save(ctx, from)
		if err != nil {
			return fmt.Errorf("saving 'from' account: %w", err)
		}

		err = s.accounts.Save(ctx, to)
		if err != nil {
			return fmt.Errorf("saving 'to' account: %w", err)
		}

		err = s.CheckLedgerBalanceByCurrency(ctx)
		if err != nil {
			return fmt.Errorf("checking ledger balance by currency: %w", err)
		}

		err = s.checkAccountLedgerConsistency(ctx, from)
		if err != nil {
			return fmt.Errorf("checking 'from' account ledger consistency: %w", err)
		}

		err = s.checkAccountLedgerConsistency(ctx, to)
		if err != nil {
			return fmt.Errorf("checking 'to' account ledger consistency: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("doing atomic operation: %w", err)
	}

	return nil
}
