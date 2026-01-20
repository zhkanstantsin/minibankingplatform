package service

import (
	"context"
	"fmt"
	"minibankingplatform/internal/domain"
	"time"

	"github.com/shopspring/decimal"
)

type ReconciliationReport struct {
	Timestamp            time.Time
	IsConsistent         bool
	LedgerBalances       []LedgerCurrencyStatus
	AccountMismatches    []AccountMismatch
	TotalAccountsChecked int
}

type LedgerCurrencyStatus struct {
	Currency   domain.Currency
	TotalSum   decimal.Decimal
	IsBalanced bool
}

type AccountMismatch struct {
	AccountID      domain.AccountID
	Currency       domain.Currency
	AccountBalance decimal.Decimal
	LedgerBalance  decimal.Decimal
	Difference     decimal.Decimal
}

func (s *Service) CheckLedgerBalanceByCurrency(ctx context.Context) error {
	totals, err := s.ledger.GetTotalBalanceByCurrency(ctx)
	if err != nil {
		return fmt.Errorf("getting ledger totals by currency: %w", err)
	}

	for currency, total := range totals {
		if !total.IsZero() {
			return domain.NewLedgerImbalanceError(currency, total.Amount())
		}
	}

	return nil
}

func (s *Service) checkAccountLedgerConsistency(ctx context.Context, account *domain.Account) error {
	ledgerBalance, err := s.ledger.GetAccountBalance(ctx, account.ID(), account.Balance().Currency())
	if err != nil {
		return fmt.Errorf("getting ledger balance for account %v: %w", account.ID(), err)
	}

	accountBalance := account.Balance()

	if !ledgerBalance.Amount().Equal(accountBalance.Amount()) {
		return domain.NewAccountBalanceMismatchError(account.ID(), accountBalance.Amount(), ledgerBalance.Amount())
	}

	return nil
}

func (s *Service) CheckAllAccountBalances(ctx context.Context) error {
	mismatches, err := s.ledger.GetAccountBalanceMismatches(ctx)
	if err != nil {
		return fmt.Errorf("getting account balance mismatches: %w", err)
	}

	if len(mismatches) > 0 {
		m := mismatches[0]
		return domain.NewAccountBalanceMismatchError(m.AccountID, m.AccountBalance, m.LedgerBalance)
	}

	return nil
}

func (s *Service) Reconcile(ctx context.Context) (*ReconciliationReport, error) {
	report := &ReconciliationReport{
		Timestamp:    time.Now(),
		IsConsistent: true,
	}

	totals, err := s.ledger.GetTotalBalanceByCurrency(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting ledger totals by currency: %w", err)
	}

	for currency, total := range totals {
		status := LedgerCurrencyStatus{
			Currency:   currency,
			TotalSum:   total.Amount(),
			IsBalanced: total.IsZero(),
		}
		report.LedgerBalances = append(report.LedgerBalances, status)

		if !status.IsBalanced {
			report.IsConsistent = false
		}
	}

	mismatches, err := s.ledger.GetAccountBalanceMismatches(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting account balance mismatches: %w", err)
	}

	for _, m := range mismatches {
		mismatch := AccountMismatch{
			AccountID:      m.AccountID,
			Currency:       m.Currency,
			AccountBalance: m.AccountBalance,
			LedgerBalance:  m.LedgerBalance,
			Difference:     m.AccountBalance.Sub(m.LedgerBalance),
		}
		report.AccountMismatches = append(report.AccountMismatches, mismatch)
		report.IsConsistent = false
	}

	accountsCount, err := s.accounts.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting accounts: %w", err)
	}
	report.TotalAccountsChecked = accountsCount

	return report, nil
}
