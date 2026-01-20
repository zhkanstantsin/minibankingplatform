import { Fragment } from 'react';
import { TableRow, TableCell, Badge, MoneyDisplay } from '@shared/ui';
import { formatDate } from '@shared/lib';
import type { Transaction } from '../model/types';

interface TransactionRowProps {
  transaction: Transaction;
  userAccountIds: string[];
}

export function TransactionRow({ transaction, userAccountIds }: TransactionRowProps) {
  // For exchanges, render two rows (debit and credit)
  if (transaction.type === 'exchange' && transaction.exchangeDetails) {
    const { sourceAmount, targetAmount } = transaction.exchangeDetails;
    const date = formatDate(transaction.timestamp);

    return (
      <Fragment>
        <TableRow>
          <TableCell>{date}</TableCell>
          <TableCell>
            <Badge>exchange</Badge>
          </TableCell>
          <TableCell>Exchange {sourceAmount.currency} → {targetAmount.currency}</TableCell>
          <TableCell>
            <MoneyDisplay
              amount={`-${sourceAmount.amount}`}
              currency={sourceAmount.currency}
              showSign
            />
          </TableCell>
        </TableRow>
        <TableRow>
          <TableCell>{date}</TableCell>
          <TableCell>
            <Badge>exchange</Badge>
          </TableCell>
          <TableCell>Exchange {sourceAmount.currency} → {targetAmount.currency}</TableCell>
          <TableCell>
            <MoneyDisplay
              amount={targetAmount.amount}
              currency={targetAmount.currency}
              showSign
            />
          </TableCell>
        </TableRow>
      </Fragment>
    );
  }

  // For transfers
  if (transaction.type === 'transfer' && transaction.transferDetails) {
    const isIncoming = userAccountIds.includes(transaction.transferDetails.recipientAccountId);
    const amount = transaction.transferDetails.amount.amount;
    const description = isIncoming
      ? `Received from ${transaction.accountId.slice(0, 8)}...`
      : `Sent to ${transaction.transferDetails.recipientAccountId.slice(0, 8)}...`;

    return (
      <TableRow>
        <TableCell>{formatDate(transaction.timestamp)}</TableCell>
        <TableCell>
          <Badge>transfer</Badge>
        </TableCell>
        <TableCell>{description}</TableCell>
        <TableCell>
          <MoneyDisplay
            amount={isIncoming ? amount : `-${amount}`}
            currency={transaction.transferDetails.amount.currency}
            showSign
          />
        </TableCell>
      </TableRow>
    );
  }

  // Fallback for other transaction types
  return (
    <TableRow>
      <TableCell>{formatDate(transaction.timestamp)}</TableCell>
      <TableCell>
        <Badge>{transaction.type}</Badge>
      </TableCell>
      <TableCell>{transaction.type}</TableCell>
      <TableCell>—</TableCell>
    </TableRow>
  );
}
