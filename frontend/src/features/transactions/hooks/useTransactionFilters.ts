import { useState } from 'react';
import type { TransactionType } from '@shared/types';
import type { TransactionFilters } from '@entities/transaction';

export function useTransactionFilters(initialLimit = 20) {
  const [filters, setFilters] = useState<TransactionFilters>({
    page: 1,
    limit: initialLimit,
  });

  return {
    filters,
    setType: (type?: TransactionType) =>
      setFilters((f) => ({ ...f, type, page: 1 })),
    setPage: (page: number) => setFilters((f) => ({ ...f, page })),
    setLimit: (limit: number) => setFilters((f) => ({ ...f, limit, page: 1 })),
  };
}
