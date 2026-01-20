import type { Money, Currency } from '@shared/types';

export interface ExchangeRequest {
  sourceAccountId: string;
  targetAccountId: string;
  amount: string;
}

export interface ExchangeResponse {
  transactionId: string;
  sourceAccountId: string;
  targetAccountId: string;
  sourceAmount: Money;
  targetAmount: Money;
  exchangeRate: string;
  timestamp: string;
}

export interface ExchangeCalculation {
  sourceAmount: Money;
  targetAmount: Money;
  exchangeRate: {
    sourceCurrency: Currency;
    targetCurrency: Currency;
    rate: string;
  };
}
