import type { Money, TransactionType, PaginationInfo } from '@shared/types';

export interface TransferDetails {
  id: string;
  recipientAccountId: string;
  amount: Money;
}

export interface ExchangeDetails {
  id: string;
  sourceAccountId: string;
  targetAccountId: string;
  sourceAmount: Money;
  targetAmount: Money;
  exchangeRate: string;
}

export interface Transaction {
  id: string;
  type: TransactionType;
  accountId: string;
  timestamp: string;
  transferDetails?: TransferDetails | null;
  exchangeDetails?: ExchangeDetails | null;
}

export interface TransactionsResponse {
  transactions: Transaction[];
  pagination: PaginationInfo;
}

export interface TransactionFilters {
  type?: TransactionType;
  page: number;
  limit: number;
}
