import { useMutation, useQueryClient } from '@tanstack/react-query';
import { authApi } from '../api/authApi';
import { useAuthStore } from '../model/authStore';
import type { LoginFormData } from '../model/schemas';

export function useLogin() {
  const setToken = useAuthStore((s) => s.setToken);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: LoginFormData) => authApi.login(data),
    onSuccess: (response) => {
      setToken(response.token);
      queryClient.invalidateQueries({ queryKey: ['currentUser'] });
    },
  });
}
