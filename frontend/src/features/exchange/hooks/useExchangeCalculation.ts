import { useQuery } from '@tanstack/react-query';
import type { Currency } from '@shared/types';
import { exchangeApi } from '../api/exchangeApi';

export function useExchangeCalculation(
  amount: string,
  sourceCurrency: Currency | undefined,
  targetCurrency: Currency | undefined
) {
  return useQuery({
    queryKey: ['exchangeCalculation', amount, sourceCurrency, targetCurrency],
    queryFn: () => exchangeApi.calculate(amount, sourceCurrency!, targetCurrency!),
    enabled:
      !!amount &&
      parseFloat(amount) > 0 &&
      !!sourceCurrency &&
      !!targetCurrency &&
      sourceCurrency !== targetCurrency,
  });
}
