package infrastructure

import (
	"minibankingplatform/internal/domain"

	"github.com/shopspring/decimal"
)

type FixedExchangeRateProvider struct {
	usdToEurRate decimal.Decimal
}

func NewFixedExchangeRateProvider(usdToEurRate decimal.Decimal) *FixedExchangeRateProvider {
	return &FixedExchangeRateProvider{
		usdToEurRate: usdToEurRate,
	}
}

func (p *FixedExchangeRateProvider) GetRate(from domain.Currency, to domain.Currency) (domain.ExchangeRate, error) {
	if from == to {
		return domain.ExchangeRate{}, domain.NewSameCurrencyExchangeRateError(from)
	}

	if from == domain.CurrencyUSD && to == domain.CurrencyEUR {
		return domain.NewExchangeRate(from, to, p.usdToEurRate)
	}

	if from == domain.CurrencyEUR && to == domain.CurrencyUSD {
		inverseRate := decimal.NewFromInt(1).Div(p.usdToEurRate).Round(6)
		return domain.NewExchangeRate(from, to, inverseRate)
	}

	return domain.ExchangeRate{}, domain.NewExchangeRateNotFoundError(from, to)
}
