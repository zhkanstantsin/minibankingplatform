import type { Money } from '@shared/types';

export function formatMoney(money: Money): string {
  const symbol = money.currency === 'USD' ? '$' : '€';
  const amount = parseFloat(money.amount).toFixed(2);
  return `${symbol}${amount}`;
}

export function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function cn(...classes: (string | boolean | undefined | null)[]): string {
  return classes.filter(Boolean).join(' ');
}
