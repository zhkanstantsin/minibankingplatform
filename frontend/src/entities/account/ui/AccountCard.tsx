import { useState } from 'react';
import { Card, CardContent, Badge, MoneyDisplay } from '@shared/ui';
import type { Account } from '../model/types';

interface AccountCardProps {
  account: Account;
}

export function AccountCard({ account }: AccountCardProps) {
  const [copied, setCopied] = useState(false);

  const copyAccountId = async () => {
    await navigator.clipboard.writeText(account.id);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Card className="min-w-[250px]">
      <CardContent>
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <Badge>{account.balance.currency}</Badge>
          </div>
          <MoneyDisplay
            amount={account.balance.amount}
            currency={account.balance.currency}
            size="lg"
          />
          <div className="pt-2 border-t border-gray-200">
            <div className="text-xs text-gray-500 mb-1">Account ID</div>
            <div className="flex items-center gap-2">
              <code className="text-xs bg-gray-100 px-2 py-1 rounded font-mono truncate flex-1">
                {account.id}
              </code>
              <button
                onClick={copyAccountId}
                className="text-xs px-2 py-1 bg-gray-100 hover:bg-gray-200 rounded transition-colors"
                title="Copy Account ID"
              >
                {copied ? 'Copied!' : 'Copy'}
              </button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
