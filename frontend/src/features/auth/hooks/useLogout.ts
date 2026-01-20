import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '../model/authStore';

export function useLogout() {
  const logout = useAuthStore((s) => s.logout);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      logout();
    },
    onSuccess: () => {
      queryClient.clear();
    },
  });
}
