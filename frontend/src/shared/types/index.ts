export type Currency = 'USD' | 'EUR';
export type TransactionType = 'transfer' | 'exchange' | 'deposit' | 'withdrawal';

export interface Money {
  amount: string;
  currency: Currency;
}

export interface PaginationInfo {
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}

export interface ProblemDetails {
  type: string;
  title: string;
  status: number;
  detail: string;
  instance?: string;
  [key: string]: unknown;
}

export interface LedgerCurrencyStatus {
  currency: Currency;
  totalSum: string;
  isBalanced: boolean;
}

export interface AccountMismatch {
  accountId: string;
  currency: Currency;
  accountBalance: string;
  ledgerBalance: string;
  difference: string;
}

export interface ReconcileReport {
  timestamp: string;
  isConsistent: boolean;
  ledgerBalances: LedgerCurrencyStatus[];
  accountMismatches: AccountMismatch[];
  totalAccountsChecked: number;
}
