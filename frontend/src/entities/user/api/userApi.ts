import { apiClient } from '@shared/api';
import type { User } from '../model/types';

export const userApi = {
  getCurrentUser: () => apiClient.get<User>('/auth/me').then((r) => r.data),
};
