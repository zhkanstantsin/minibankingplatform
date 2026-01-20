import {
  Card,
  CardHeader,
  CardContent,
  CardFooter,
  Stack,
  Table,
  TableHead,
  TableBody,
  TableRow,
  TableCell,
  Select,
  Pagination,
  Spinner,
  Alert,
} from '@shared/ui';
import type { TransactionType } from '@shared/types';
import { useTransactions, TransactionRow } from '@entities/transaction';
import { useAccounts } from '@entities/account';
import { useTransactionFilters } from '@features/transactions';

interface TransactionListProps {
  limit?: number;
  showFilters?: boolean;
  showPagination?: boolean;
}

export function TransactionList({
  limit = 20,
  showFilters = false,
  showPagination = false,
}: TransactionListProps) {
  const { filters, setType, setPage } = useTransactionFilters(limit);
  const { data, isLoading, isError } = useTransactions(filters);
  const { data: accounts } = useAccounts();

  const userAccountIds = accounts?.map((a) => a.id) ?? [];

  return (
    <Card>
      <CardHeader>
        <Stack direction="row" justify="between" align="center">
          <h3 className="font-semibold">
            {showFilters ? 'Transaction History' : 'Recent Transactions'}
          </h3>
          {showFilters && (
            <Select
              value={filters.type || ''}
              onChange={(e) =>
                setType((e.target.value as TransactionType) || undefined)
              }
              options={[
                { value: '', label: 'All Types' },
                { value: 'transfer', label: 'Transfers' },
                { value: 'exchange', label: 'Exchanges' },
              ]}
              className="w-40"
            />
          )}
        </Stack>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex justify-center py-8">
            <Spinner />
          </div>
        ) : isError ? (
          <Alert variant="error">Failed to load transactions</Alert>
        ) : data?.transactions.length === 0 ? (
          <p className="text-center text-gray-500 py-8">No transactions yet</p>
        ) : (
          <Table>
            <TableHead>
              <TableRow>
                <TableCell header>Date</TableCell>
                <TableCell header>Type</TableCell>
                <TableCell header>Details</TableCell>
                <TableCell header>Amount</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {data?.transactions.map((tx) => (
                <TransactionRow key={tx.id} transaction={tx} userAccountIds={userAccountIds} />
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
      {showPagination && data?.pagination && data.pagination.totalPages > 1 && (
        <CardFooter>
          <Pagination
            current={data.pagination.page}
            total={data.pagination.totalPages}
            onChange={setPage}
          />
        </CardFooter>
      )}
    </Card>
  );
}
