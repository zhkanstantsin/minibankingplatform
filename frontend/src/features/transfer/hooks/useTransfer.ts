import { useMutation, useQueryClient } from '@tanstack/react-query';
import { transferApi } from '../api/transferApi';
import type { TransferRequest } from '../model/types';

export function useTransfer() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: TransferRequest) => transferApi.transfer(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
    },
  });
}
