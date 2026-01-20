import { PageLayout } from '@shared/ui';
import { TransactionList } from '@widgets/transaction-list';

export function TransactionsPage() {
  return (
    <PageLayout title="Transaction History">
      <TransactionList showFilters showPagination />
    </PageLayout>
  );
}
