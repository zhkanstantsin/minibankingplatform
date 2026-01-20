package service

import (
	"context"
	"fmt"
	"minibankingplatform/internal/domain"
	"minibankingplatform/internal/infrastructure"
)

type GetTransactionsCommand struct {
	UserID          domain.UserID
	TransactionType *domain.TransactionType
	Limit           int
	Offset          int
}

type TransactionsResult struct {
	Transactions []*domain.TransactionWithDetails
	Total        int
	Limit        int
	Offset       int
}

func (s *Service) GetTransactions(ctx context.Context, cmd *GetTransactionsCommand) (*TransactionsResult, error) {
	filter := infrastructure.TransactionsFilter{
		UserID:          cmd.UserID,
		TransactionType: cmd.TransactionType,
		Limit:           cmd.Limit,
		Offset:          cmd.Offset,
	}

	transactions, err := s.transactions.GetList(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("getting transactions list: %w", err)
	}

	total, err := s.transactions.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("counting transactions: %w", err)
	}

	return &TransactionsResult{
		Transactions: transactions,
		Total:        total,
		Limit:        cmd.Limit,
		Offset:       cmd.Offset,
	}, nil
}
