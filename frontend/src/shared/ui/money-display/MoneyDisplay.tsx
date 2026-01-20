import { cn } from '@shared/lib';
import type { Currency } from '@shared/types';

interface MoneyDisplayProps {
  amount: string;
  currency: Currency;
  size?: 'sm' | 'md' | 'lg';
  showSign?: boolean;
  className?: string;
}

export function MoneyDisplay({
  amount,
  currency,
  size = 'md',
  showSign = false,
  className,
}: MoneyDisplayProps) {
  const symbol = currency === 'USD' ? '$' : '€';
  const numericAmount = parseFloat(amount);
  const formattedAmount = Math.abs(numericAmount).toFixed(2);
  const isNegative = numericAmount < 0;

  const sizes = {
    sm: 'text-sm',
    md: 'text-base',
    lg: 'text-2xl font-bold',
  };

  const colors = showSign
    ? isNegative
      ? 'text-red-600'
      : 'text-green-600'
    : 'text-gray-900';

  return (
    <span className={cn(sizes[size], colors, className)}>
      {showSign && !isNegative && '+'}
      {isNegative && '-'}
      {symbol}
      {formattedAmount}
    </span>
  );
}
