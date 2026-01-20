import { apiClient } from '@shared/api';
import type { Currency } from '@shared/types';
import type { ExchangeRequest, ExchangeResponse, ExchangeCalculation } from '../model/types';

export const exchangeApi = {
  exchange: (data: ExchangeRequest) =>
    apiClient.post<ExchangeResponse>('/transactions/exchange', data).then((r) => r.data),

  calculate: (amount: string, sourceCurrency: Currency, targetCurrency: Currency) =>
    apiClient
      .get<ExchangeCalculation>('/transactions/exchange/calculate', {
        params: { amount, sourceCurrency, targetCurrency },
      })
      .then((r) => r.data),
};
