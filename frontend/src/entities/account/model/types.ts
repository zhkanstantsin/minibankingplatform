import type { Money } from '@shared/types';

export interface Account {
  id: string;
  userId: string;
  balance: Money;
}

export interface Balance {
  accountId: string;
  balance: Money;
}
