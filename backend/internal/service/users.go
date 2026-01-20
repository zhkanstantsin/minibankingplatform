package service

import (
	"context"
	"errors"
	"fmt"
	"minibankingplatform/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type RegisterCommand struct {
	Email    string
	Password string
}

type AuthResult struct {
	UserID uuid.UUID
	Email  string
	Token  string
}

func (s *Service) Register(ctx context.Context, cmd *RegisterCommand) (*AuthResult, error) {
	var result *AuthResult

	err := s.trm.Do(ctx, func(ctx context.Context) error {
		exists, err := s.users.ExistsByEmail(ctx, cmd.Email)
		if err != nil {
			return fmt.Errorf("checking user existence: %w", err)
		}
		if exists {
			return domain.NewUserAlreadyExistsError(cmd.Email)
		}

		userID := domain.GenerateUserID()
		user, err := domain.NewUser(userID, cmd.Email, cmd.Password)
		if err != nil {
			return fmt.Errorf("creating user: %w", err)
		}

		err = s.users.Save(ctx, user)
		if err != nil {
			return fmt.Errorf("saving user: %w", err)
		}

		zeroUSD, err := domain.NewMoney(decimal.Zero, domain.CurrencyUSD)
		if err != nil {
			return fmt.Errorf("creating zero USD balance: %w", err)
		}
		usdAccount := domain.NewAccount(domain.GenerateAccountID(), userID, zeroUSD)
		err = s.accounts.Save(ctx, usdAccount)
		if err != nil {
			return fmt.Errorf("saving USD account: %w", err)
		}

		zeroEUR, err := domain.NewMoney(decimal.Zero, domain.CurrencyEUR)
		if err != nil {
			return fmt.Errorf("creating zero EUR balance: %w", err)
		}
		eurAccount := domain.NewAccount(domain.GenerateAccountID(), userID, zeroEUR)
		err = s.accounts.Save(ctx, eurAccount)
		if err != nil {
			return fmt.Errorf("saving EUR account: %w", err)
		}

		usdCashbookID := domain.GetCashbookAccount(domain.CurrencyUSD)
		usdCashbook, err := s.accounts.GetForUpdate(ctx, usdCashbookID)
		if err != nil {
			return fmt.Errorf("getting USD cashbook: %w", err)
		}

		initialUSD, err := domain.NewMoney(decimal.NewFromInt(1000), domain.CurrencyUSD)
		if err != nil {
			return fmt.Errorf("creating initial USD amount: %w", err)
		}

		usdTransferDetails, err := s.transfer.Execute(usdCashbook, usdAccount, initialUSD, time.Now())
		if err != nil {
			return fmt.Errorf("transferring initial USD: %w", err)
		}

		err = s.transfers.Insert(ctx, usdTransferDetails)
		if err != nil {
			return fmt.Errorf("inserting USD transfer: %w", err)
		}

		err = s.accounts.Save(ctx, usdCashbook)
		if err != nil {
			return fmt.Errorf("saving USD cashbook: %w", err)
		}

		err = s.accounts.Save(ctx, usdAccount)
		if err != nil {
			return fmt.Errorf("saving funded USD account: %w", err)
		}

		eurCashbookID := domain.GetCashbookAccount(domain.CurrencyEUR)
		eurCashbook, err := s.accounts.GetForUpdate(ctx, eurCashbookID)
		if err != nil {
			return fmt.Errorf("getting EUR cashbook: %w", err)
		}

		initialEUR, err := domain.NewMoney(decimal.NewFromInt(500), domain.CurrencyEUR)
		if err != nil {
			return fmt.Errorf("creating initial EUR amount: %w", err)
		}

		eurTransferDetails, err := s.transfer.Execute(eurCashbook, eurAccount, initialEUR, time.Now())
		if err != nil {
			return fmt.Errorf("transferring initial EUR: %w", err)
		}

		err = s.transfers.Insert(ctx, eurTransferDetails)
		if err != nil {
			return fmt.Errorf("inserting EUR transfer: %w", err)
		}

		err = s.accounts.Save(ctx, eurCashbook)
		if err != nil {
			return fmt.Errorf("saving EUR cashbook: %w", err)
		}

		err = s.accounts.Save(ctx, eurAccount)
		if err != nil {
			return fmt.Errorf("saving funded EUR account: %w", err)
		}

		err = s.CheckLedgerBalanceByCurrency(ctx)
		if err != nil {
			return fmt.Errorf("checking ledger balance: %w", err)
		}

		err = s.checkAccountLedgerConsistency(ctx, usdAccount)
		if err != nil {
			return fmt.Errorf("checking USD account ledger consistency: %w", err)
		}

		err = s.checkAccountLedgerConsistency(ctx, eurAccount)
		if err != nil {
			return fmt.Errorf("checking EUR account ledger consistency: %w", err)
		}

		token, err := s.tokenManager.GenerateToken(uuid.UUID(userID), cmd.Email)
		if err != nil {
			return fmt.Errorf("generating token: %w", err)
		}

		result = &AuthResult{
			UserID: uuid.UUID(userID),
			Email:  cmd.Email,
			Token:  token,
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("registering user: %w", err)
	}

	return result, nil
}

type LoginCommand struct {
	Email    string
	Password string
}

func (s *Service) Login(ctx context.Context, cmd *LoginCommand) (*AuthResult, error) {
	user, err := s.users.GetByEmail(ctx, cmd.Email)
	if err != nil {
		var notFoundErr *domain.UserNotFoundError
		if errors.As(err, &notFoundErr) {
			return nil, domain.NewInvalidCredentialsError()
		}
		return nil, fmt.Errorf("getting user: %w", err)
	}

	if !user.CheckPassword(cmd.Password) {
		return nil, domain.NewInvalidCredentialsError()
	}

	token, err := s.tokenManager.GenerateToken(uuid.UUID(user.ID()), user.Email())
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	return &AuthResult{
		UserID: uuid.UUID(user.ID()),
		Email:  user.Email(),
		Token:  token,
	}, nil
}
