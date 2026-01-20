import { useQuery } from '@tanstack/react-query';
import { transactionApi } from '../api/transactionApi';
import type { TransactionFilters } from '../model/types';

export function useTransactions(filters: TransactionFilters) {
  return useQuery({
    queryKey: ['transactions', filters],
    queryFn: () => transactionApi.getAll(filters),
  });
}
