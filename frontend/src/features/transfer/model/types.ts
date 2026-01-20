import type { Money } from '@shared/types';

export interface TransferRequest {
  fromAccountId: string;
  toAccountId: string;
  amount: string;
  currency: string;
}

export interface TransferResponse {
  transactionId: string;
  fromAccountId: string;
  toAccountId: string;
  amount: Money;
  timestamp: string;
}
