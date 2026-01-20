import { apiClient } from '@shared/api';
import type { ReconcileReport } from '@shared/types';

export const reconcileApi = {
  check: async (): Promise<ReconcileReport> => {
    const response = await apiClient.get('/system/reconcile');
    return response.data;
  },
};
