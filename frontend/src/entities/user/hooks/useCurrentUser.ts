import { useQuery } from '@tanstack/react-query';
import { userApi } from '../api/userApi';

export function useCurrentUser() {
  return useQuery({
    queryKey: ['currentUser'],
    queryFn: userApi.getCurrentUser,
    enabled: !!localStorage.getItem('token'),
    retry: false,
  });
}
