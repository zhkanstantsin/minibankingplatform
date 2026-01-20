import { apiClient } from '@shared/api';
import type { Account, Balance } from '../model/types';

export const accountApi = {
  getAll: () => apiClient.get<Account[]>('/accounts').then((r) => r.data),

  getBalance: (id: string) =>
    apiClient.get<Balance>(`/accounts/${id}/balance`).then((r) => r.data),
};
