import { useMutation, useQueryClient } from '@tanstack/react-query';
import { exchangeApi } from '../api/exchangeApi';
import type { ExchangeRequest } from '../model/types';

export function useExchange() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: ExchangeRequest) => exchangeApi.exchange(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
    },
  });
}
