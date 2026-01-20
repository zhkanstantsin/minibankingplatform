import { apiClient } from '@shared/api';
import type { AuthResponse } from '@entities/user';
import type { LoginFormData, RegisterFormData } from '../model/schemas';

export const authApi = {
  login: (data: LoginFormData) =>
    apiClient.post<AuthResponse>('/auth/login', data).then((r) => r.data),

  register: (data: RegisterFormData) =>
    apiClient.post<AuthResponse>('/auth/register', data).then((r) => r.data),
};
