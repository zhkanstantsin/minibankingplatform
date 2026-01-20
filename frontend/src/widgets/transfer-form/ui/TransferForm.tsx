import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import {
  Card,
  CardHeader,
  CardContent,
  Stack,
  FormField,
  Input,
  Select,
  Button,
  Alert,
} from '@shared/ui';
import { getErrorMessage } from '@shared/api';
import { formatMoney } from '@shared/lib';
import { useAccounts } from '@entities/account';
import { useTransfer, transferSchema, type TransferFormData } from '@features/transfer';

export function TransferForm() {
  const { data: accounts } = useAccounts();
  const transfer = useTransfer();

  const form = useForm<TransferFormData>({
    resolver: zodResolver(transferSchema),
    defaultValues: {
      fromAccountId: '',
      toAccountId: '',
      amount: '',
      currency: 'USD',
    },
  });

  const selectedCurrency = form.watch('currency');
  const sourceAccount = accounts?.find(
    (a) => a.balance.currency === selectedCurrency
  );

  const onSubmit = form.handleSubmit((data) => {
    transfer.mutate(data, {
      onSuccess: () => {
        form.reset();
      },
    });
  });

  const accountOptions =
    accounts?.map((a) => ({
      value: a.id,
      label: `${a.balance.currency} - ${formatMoney(a.balance)}`,
    })) ?? [];

  return (
    <Card className="flex-1">
      <CardHeader>
        <h3 className="font-semibold">Transfer Money</h3>
      </CardHeader>
      <CardContent>
        <form onSubmit={onSubmit}>
          <Stack direction="column" gap="md">
            <FormField
              label="Currency"
              error={form.formState.errors.currency?.message}
            >
              <Select
                options={[
                  { value: 'USD', label: 'USD' },
                  { value: 'EUR', label: 'EUR' },
                ]}
                error={!!form.formState.errors.currency}
                {...form.register('currency', {
                  onChange: (e) => {
                    const currency = e.target.value;
                    const account = accounts?.find(
                      (a) => a.balance.currency === currency
                    );
                    if (account) {
                      form.setValue('fromAccountId', account.id);
                    }
                  },
                })}
              />
            </FormField>

            <FormField
              label="From Account"
              error={form.formState.errors.fromAccountId?.message}
            >
              <Select
                options={accountOptions.filter(
                  (a) =>
                    accounts?.find((acc) => acc.id === a.value)?.balance
                      .currency === selectedCurrency
                )}
                error={!!form.formState.errors.fromAccountId}
                {...form.register('fromAccountId')}
              />
            </FormField>

            {sourceAccount && (
              <p className="text-sm text-gray-500">
                Available: {formatMoney(sourceAccount.balance)}
              </p>
            )}

            <FormField
              label="Recipient Account ID"
              error={form.formState.errors.toAccountId?.message}
            >
              <Input
                placeholder="Enter recipient's account UUID"
                error={!!form.formState.errors.toAccountId}
                {...form.register('toAccountId')}
              />
            </FormField>

            <FormField
              label="Amount"
              error={form.formState.errors.amount?.message}
            >
              <Input
                type="number"
                step="0.01"
                min="0.01"
                placeholder="0.00"
                error={!!form.formState.errors.amount}
                {...form.register('amount')}
              />
            </FormField>

            {transfer.isError && (
              <Alert variant="error">{getErrorMessage(transfer.error)}</Alert>
            )}

            {transfer.isSuccess && (
              <Alert variant="success">Transfer completed successfully!</Alert>
            )}

            <Button type="submit" loading={transfer.isPending}>
              Send Transfer
            </Button>
          </Stack>
        </form>
      </CardContent>
    </Card>
  );
}
