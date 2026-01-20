import { useQuery } from '@tanstack/react-query';
import type { Currency } from '@shared/types';
import { accountApi } from '../api/accountApi';

export function useAccounts() {
  return useQuery({
    queryKey: ['accounts'],
    queryFn: accountApi.getAll,
  });
}

export function useAccountByCurrency(currency: Currency) {
  const { data: accounts } = useAccounts();
  return accounts?.find((a) => a.balance.currency === currency);
}
