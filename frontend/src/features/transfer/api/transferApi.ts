import { apiClient } from '@shared/api';
import type { TransferRequest, TransferResponse } from '../model/types';

export const transferApi = {
  transfer: (data: TransferRequest) =>
    apiClient.post<TransferResponse>('/transactions/transfer', data).then((r) => r.data),
};
