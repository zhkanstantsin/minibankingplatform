import { useEffect } from 'react';
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
  MoneyDisplay,
} from '@shared/ui';
import { getErrorMessage } from '@shared/api';
import { formatMoney } from '@shared/lib';
import { useAccounts } from '@entities/account';
import {
  useExchange,
  useExchangeCalculation,
  exchangeSchema,
  type ExchangeFormData,
} from '@features/exchange';

export function ExchangeForm() {
  const { data: accounts } = useAccounts();
  const exchange = useExchange();

  const form = useForm<ExchangeFormData>({
    resolver: zodResolver(exchangeSchema),
    defaultValues: {
      sourceAccountId: '',
      targetAccountId: '',
      amount: '',
    },
  });

  const watchAmount = form.watch('amount');
  const watchSourceId = form.watch('sourceAccountId');

  const sourceAccount = accounts?.find((a) => a.id === watchSourceId);
  const targetAccount = accounts?.find(
    (a) => a.balance.currency !== sourceAccount?.balance.currency
  );

  // Auto-select target account when source changes
  useEffect(() => {
    if (targetAccount && form.getValues('targetAccountId') !== targetAccount.id) {
      form.setValue('targetAccountId', targetAccount.id);
    }
  }, [targetAccount, form]);

  const { data: calculation, isFetching: isCalculating } = useExchangeCalculation(
    watchAmount,
    sourceAccount?.balance.currency,
    targetAccount?.balance.currency
  );

  const onSubmit = form.handleSubmit((data) => {
    exchange.mutate(data, {
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
        <h3 className="font-semibold">Exchange Currency</h3>
      </CardHeader>
      <CardContent>
        <form onSubmit={onSubmit}>
          <Stack direction="column" gap="md">
            <FormField
              label="From Account"
              error={form.formState.errors.sourceAccountId?.message}
            >
              <Select
                options={accountOptions}
                error={!!form.formState.errors.sourceAccountId}
                {...form.register('sourceAccountId')}
              />
            </FormField>

            {sourceAccount && (
              <p className="text-sm text-gray-500">
                Available: {formatMoney(sourceAccount.balance)}
              </p>
            )}

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

            <FormField
              label="To Account"
              error={form.formState.errors.targetAccountId?.message}
            >
              <Select
                options={accountOptions.filter(
                  (a) => a.value !== watchSourceId
                )}
                error={!!form.formState.errors.targetAccountId}
                {...form.register('targetAccountId')}
              />
            </FormField>

            {calculation && !isCalculating && (
              <Alert variant="info">
                <Stack direction="column" gap="xs">
                  <span>
                    You will receive:{' '}
                    <MoneyDisplay
                      amount={calculation.targetAmount.amount}
                      currency={calculation.targetAmount.currency}
                      className="font-semibold"
                    />
                  </span>
                  <span className="text-xs">
                    Rate: 1 {calculation.exchangeRate.sourceCurrency} ={' '}
                    {calculation.exchangeRate.rate}{' '}
                    {calculation.exchangeRate.targetCurrency}
                  </span>
                </Stack>
              </Alert>
            )}

            {exchange.isError && (
              <Alert variant="error">{getErrorMessage(exchange.error)}</Alert>
            )}

            {exchange.isSuccess && (
              <Alert variant="success">Exchange completed successfully!</Alert>
            )}

            <Button type="submit" loading={exchange.isPending}>
              Exchange
            </Button>
          </Stack>
        </form>
      </CardContent>
    </Card>
  );
}
