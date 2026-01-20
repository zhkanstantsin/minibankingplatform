import { apiClient } from '@shared/api';
import type { TransactionsResponse, TransactionFilters } from '../model/types';

export const transactionApi = {
  getAll: (filters: TransactionFilters) =>
    apiClient
      .get<TransactionsResponse>('/transactions', {
        params: {
          type: filters.type,
          page: filters.page,
          limit: filters.limit,
        },
      })
      .then((r) => r.data),
};
