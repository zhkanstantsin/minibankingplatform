package service

import (
	"context"
	"fmt"
	"minibankingplatform/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ExchangeCommand struct {
	SourceAccount domain.AccountID
	TargetAccount domain.AccountID
	SourceAmount  domain.Money
	Time          time.Time
}

func NewExchangeCommand(
	sourceAccount uuid.UUID,
	targetAccount uuid.UUID,
	amount string,
	sourceCurrency string,
	time time.Time,
) (*ExchangeCommand, error) {
	decimalAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	currency, err := domain.ParseCurrency(sourceCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	money, err := domain.NewMoney(decimalAmount, currency)
	if err != nil {
		return nil, fmt.Errorf("cannot get money value: %w", err)
	}

	return &ExchangeCommand{
		SourceAccount: domain.AccountID(sourceAccount),
		TargetAccount: domain.AccountID(targetAccount),
		SourceAmount:  money,
		Time:          time,
	}, nil
}

func (s *Service) Exchange(ctx context.Context, cmd *ExchangeCommand) error {
	err := s.trm.Do(ctx, func(ctx context.Context) error {
		sourceAccount, err := s.accounts.GetForUpdate(ctx, cmd.SourceAccount)
		if err != nil {
			return fmt.Errorf("getting source account: %w", err)
		}

		targetAccount, err := s.accounts.GetForUpdate(ctx, cmd.TargetAccount)
		if err != nil {
			return fmt.Errorf("getting target account: %w", err)
		}

		usdCashbookID := domain.GetCashbookAccount(domain.CurrencyUSD)
		eurCashbookID := domain.GetCashbookAccount(domain.CurrencyEUR)

		usdCashbookAccount, err := s.accounts.GetForUpdate(ctx, usdCashbookID)
		if err != nil {
			return fmt.Errorf("getting USD cashbook account: %w", err)
		}

		eurCashbookAccount, err := s.accounts.GetForUpdate(ctx, eurCashbookID)
		if err != nil {
			return fmt.Errorf("getting EUR cashbook account: %w", err)
		}

		exchangeRate, err := s.exchangeRateProvider.GetRate(
			cmd.SourceAmount.Currency(),
			targetAccount.Balance().Currency(),
		)
		if err != nil {
			return fmt.Errorf("getting exchange rate: %w", err)
		}

		details, err := s.exchange.Execute(
			sourceAccount,
			targetAccount,
			usdCashbookAccount,
			eurCashbookAccount,
			cmd.SourceAmount,
			exchangeRate,
			cmd.Time,
		)
		if err != nil {
			return fmt.Errorf("executing exchange domain service: %w", err)
		}

		err = s.exchanges.Insert(ctx, details)
		if err != nil {
			return fmt.Errorf("inserting exchange: %w", err)
		}

		err = s.accounts.Save(ctx, sourceAccount)
		if err != nil {
			return fmt.Errorf("saving source account: %w", err)
		}

		err = s.accounts.Save(ctx, targetAccount)
		if err != nil {
			return fmt.Errorf("saving target account: %w", err)
		}

		err = s.accounts.Save(ctx, usdCashbookAccount)
		if err != nil {
			return fmt.Errorf("saving USD cashbook account: %w", err)
		}

		err = s.accounts.Save(ctx, eurCashbookAccount)
		if err != nil {
			return fmt.Errorf("saving EUR cashbook account: %w", err)
		}

		err = s.CheckLedgerBalanceByCurrency(ctx)
		if err != nil {
			return fmt.Errorf("checking ledger balance by currency: %w", err)
		}

		err = s.checkAccountLedgerConsistency(ctx, sourceAccount)
		if err != nil {
			return fmt.Errorf("checking source account ledger consistency: %w", err)
		}

		err = s.checkAccountLedgerConsistency(ctx, targetAccount)
		if err != nil {
			return fmt.Errorf("checking target account ledger consistency: %w", err)
		}

		err = s.checkAccountLedgerConsistency(ctx, usdCashbookAccount)
		if err != nil {
			return fmt.Errorf("checking USD cashbook ledger consistency: %w", err)
		}

		err = s.checkAccountLedgerConsistency(ctx, eurCashbookAccount)
		if err != nil {
			return fmt.Errorf("checking EUR cashbook ledger consistency: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("doing atomic operation: %w", err)
	}

	return nil
}

type ExchangeCalculation struct {
	SourceAmount Money
	TargetAmount Money
	ExchangeRate domain.ExchangeRate
}

type Money struct {
	Amount   decimal.Decimal
	Currency string
}

func (s *Service) CalculateExchangeAmount(
	sourceAmount domain.Money,
	targetCurrency domain.Currency,
) (*ExchangeCalculation, error) {
	exchangeRate, err := s.exchangeRateProvider.GetRate(sourceAmount.Currency(), targetCurrency)
	if err != nil {
		return nil, fmt.Errorf("getting exchange rate: %w", err)
	}

	targetAmount, err := domain.CalculateExchangeAmount(sourceAmount, exchangeRate)
	if err != nil {
		return nil, fmt.Errorf("calculating exchange amount: %w", err)
	}

	return &ExchangeCalculation{
		SourceAmount: Money{
			Amount:   sourceAmount.Amount(),
			Currency: string(sourceAmount.Currency()),
		},
		TargetAmount: Money{
			Amount:   targetAmount.Amount(),
			Currency: string(targetAmount.Currency()),
		},
		ExchangeRate: exchangeRate,
	}, nil
}
