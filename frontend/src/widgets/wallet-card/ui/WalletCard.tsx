import { Card, CardContent, Spinner } from '@shared/ui';
import type { Currency } from '@shared/types';
import { useAccounts, AccountCard } from '@entities/account';

interface WalletCardProps {
  currency: Currency;
}

export function WalletCard({ currency }: WalletCardProps) {
  const { data: accounts, isLoading, isError } = useAccounts();
  const account = accounts?.find((a) => a.balance.currency === currency);

  if (isLoading) {
    return (
      <Card className="min-w-[200px]">
        <CardContent className="flex items-center justify-center py-8">
          <Spinner />
        </CardContent>
      </Card>
    );
  }

  if (isError || !account) {
    return null;
  }

  return <AccountCard account={account} />;
}
