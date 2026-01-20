package service

import (
	"minibankingplatform/internal/domain"
	"minibankingplatform/internal/infrastructure"
	jwtpkg "minibankingplatform/pkg/jwt"
	"minibankingplatform/pkg/trm"

	"github.com/jackc/pgx/v5"
)

type Service struct {
	transfer domain.TransferService
	exchange domain.ExchangeService

	trm *trm.TransactionManager[pgx.Tx, pgx.TxOptions]

	users                *infrastructure.UsersRepository
	accounts             *infrastructure.AccountsRepository
	transfers            *infrastructure.TransfersRepository
	exchanges            *infrastructure.ExchangesRepository
	transactions         *infrastructure.TransactionsRepository
	ledger               *infrastructure.LedgerRepository
	exchangeRateProvider domain.ExchangeRateProvider
	tokenManager         *jwtpkg.TokenManager
}

func NewService(
	trm *trm.TransactionManager[pgx.Tx, pgx.TxOptions],
	users *infrastructure.UsersRepository,
	accounts *infrastructure.AccountsRepository,
	transfers *infrastructure.TransfersRepository,
	exchanges *infrastructure.ExchangesRepository,
	transactions *infrastructure.TransactionsRepository,
	ledger *infrastructure.LedgerRepository,
	exchangeRateProvider domain.ExchangeRateProvider,
	tokenManager *jwtpkg.TokenManager,
) *Service {
	return &Service{
		transfer:             domain.TransferService{},
		exchange:             domain.ExchangeService{},
		trm:                  trm,
		users:                users,
		accounts:             accounts,
		transfers:            transfers,
		exchanges:            exchanges,
		transactions:         transactions,
		ledger:               ledger,
		exchangeRateProvider: exchangeRateProvider,
		tokenManager:         tokenManager,
	}
}
