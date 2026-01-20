import { useMutation, useQueryClient } from '@tanstack/react-query';
import { authApi } from '../api/authApi';
import { useAuthStore } from '../model/authStore';
import type { RegisterFormData } from '../model/schemas';

export function useRegister() {
  const setToken = useAuthStore((s) => s.setToken);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: RegisterFormData) => authApi.register(data),
    onSuccess: (response) => {
      setToken(response.token);
      queryClient.invalidateQueries({ queryKey: ['currentUser'] });
    },
  });
}
