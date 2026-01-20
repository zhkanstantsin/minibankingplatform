package service

import (
	"context"
	"minibankingplatform/internal/domain"
)

func (s *Service) GetUserAccounts(ctx context.Context, userID domain.UserID) ([]*domain.Account, error) {
	return s.accounts.GetByUserID(ctx, userID)
}

func (s *Service) GetAccountBalance(ctx context.Context, accountID domain.AccountID) (domain.Money, error) {
	account, err := s.accounts.Get(ctx, accountID)
	if err != nil {
		return domain.Money{}, err
	}
	return account.Balance(), nil
}
