package service_test

import (
	"testing"

	"minibankingplatform/internal/domain"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateExchangeAmount_USDtoEUR(t *testing.T) {
	t.Parallel()

	// Arrange
	svc := setupService(t, testPool)
	sourceAmount, _ := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)

	// Act
	result, err := svc.CalculateExchangeAmount(sourceAmount, domain.CurrencyEUR)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "100", result.SourceAmount.Amount.String())
	assert.Equal(t, "USD", result.SourceAmount.Currency)
	assert.True(t, result.TargetAmount.Amount.Equal(decimal.NewFromInt(92)))
	assert.Equal(t, "EUR", result.TargetAmount.Currency)
	assert.Equal(t, domain.CurrencyUSD, result.ExchangeRate.From())
	assert.Equal(t, domain.CurrencyEUR, result.ExchangeRate.To())
	assert.True(t, result.ExchangeRate.Rate().Equal(decimal.NewFromFloat(0.92)))
}

func TestCalculateExchangeAmount_EURtoUSD(t *testing.T) {
	t.Parallel()

	// Arrange
	svc := setupService(t, testPool)
	sourceAmount, _ := domain.NewMoney(decimal.NewFromInt(92), domain.CurrencyEUR)

	// Act
	result, err := svc.CalculateExchangeAmount(sourceAmount, domain.CurrencyUSD)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "92", result.SourceAmount.Amount.String())
	assert.Equal(t, "EUR", result.SourceAmount.Currency)
	assert.True(t, result.TargetAmount.Amount.Equal(decimal.NewFromInt(100)))
	assert.Equal(t, "USD", result.TargetAmount.Currency)
	assert.Equal(t, domain.CurrencyEUR, result.ExchangeRate.From())
	assert.Equal(t, domain.CurrencyUSD, result.ExchangeRate.To())
	// EUR to USD uses inverse rate: 1/0.92 ≈ 1.086957
	assert.True(t, result.ExchangeRate.Rate().Round(6).Equal(decimal.NewFromInt(1).Div(decimal.NewFromFloat(0.92)).Round(6)))
}

func TestCalculateExchangeAmount_DecimalPrecision(t *testing.T) {
	t.Parallel()

	// Arrange
	svc := setupService(t, testPool)
	sourceAmount, _ := domain.NewMoney(decimal.NewFromFloat(123.45), domain.CurrencyUSD)

	// Act
	result, err := svc.CalculateExchangeAmount(sourceAmount, domain.CurrencyEUR)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.SourceAmount.Amount.Equal(decimal.NewFromFloat(123.45)))

	// 123.45 * 0.92 = 113.574, rounded to 2 decimal places = 113.57
	expectedTarget := decimal.NewFromFloat(123.45).Mul(decimal.NewFromFloat(0.92)).Round(2)
	assert.True(t, result.TargetAmount.Amount.Equal(expectedTarget))
}
