import { PageLayout, Stack } from '@shared/ui';
import { WalletCard } from '@widgets/wallet-card';
import { TransferForm } from '@widgets/transfer-form';
import { ExchangeForm } from '@widgets/exchange-form';
import { TransactionList } from '@widgets/transaction-list';

export function DashboardPage() {
  return (
    <PageLayout title="Dashboard">
      <Stack direction="column" gap="lg">
        {/* Wallet Cards */}
        <Stack direction="row" gap="md" wrap>
          <WalletCard currency="USD" />
          <WalletCard currency="EUR" />
        </Stack>

        {/* Transfer and Exchange Forms */}
        <Stack direction="row" gap="md" wrap className="items-start">
          <TransferForm />
          <ExchangeForm />
        </Stack>

        {/* Recent Transactions */}
        <TransactionList limit={5} />
      </Stack>
    </PageLayout>
  );
}
