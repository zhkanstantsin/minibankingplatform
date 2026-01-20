import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { reconcileApi } from '../api/reconcileApi';

export function ReconcileIndicator() {
  const [isOpen, setIsOpen] = useState(false);

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['reconcile'],
    queryFn: reconcileApi.check,
    refetchInterval: 30000, // Check every 30 seconds
    staleTime: 10000,
  });

  const isConsistent = data?.isConsistent ?? true;

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        onMouseEnter={() => setIsOpen(true)}
        onMouseLeave={() => setIsOpen(false)}
        className={`w-3 h-3 rounded-full transition-colors ${
          isLoading
            ? 'bg-gray-400 animate-pulse'
            : isError
              ? 'bg-yellow-500'
              : isConsistent
                ? 'bg-green-500'
                : 'bg-red-500 animate-pulse'
        }`}
        title="Database consistency status"
      />

      {isOpen && (
        <div className="absolute right-0 top-full mt-2 w-72 bg-white rounded-lg shadow-lg border border-gray-200 p-4 z-50">
          <div className="flex items-center justify-between mb-3">
            <h4 className="font-semibold text-sm">DB Consistency Check</h4>
            <button
              onClick={() => refetch()}
              className="text-xs text-blue-600 hover:text-blue-800"
            >
              Refresh
            </button>
          </div>

          {isLoading ? (
            <p className="text-sm text-gray-500">Checking...</p>
          ) : isError ? (
            <p className="text-sm text-yellow-600">Unable to check status</p>
          ) : data ? (
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <span
                  className={`w-2 h-2 rounded-full ${isConsistent ? 'bg-green-500' : 'bg-red-500'}`}
                />
                <span className="text-sm">
                  {isConsistent ? 'All systems consistent' : 'Issues detected'}
                </span>
              </div>

              <div className="text-xs text-gray-500 space-y-1">
                <p>Accounts checked: {data.totalAccountsChecked}</p>

                {data.ledgerBalances.map((lb) => (
                  <p key={lb.currency} className="flex items-center gap-1">
                    <span className={lb.isBalanced ? 'text-green-600' : 'text-red-600'}>
                      {lb.isBalanced ? '✓' : '✗'}
                    </span>
                    {lb.currency} ledger: {lb.totalSum}
                  </p>
                ))}
              </div>

              {data.accountMismatches.length > 0 && (
                <div className="mt-2 pt-2 border-t border-gray-200">
                  <p className="text-xs font-medium text-red-600 mb-1">
                    Account mismatches:
                  </p>
                  {data.accountMismatches.map((m) => (
                    <p key={m.accountId} className="text-xs text-gray-600">
                      {m.accountId.slice(0, 8)}... ({m.currency}): diff {m.difference}
                    </p>
                  ))}
                </div>
              )}

              <p className="text-xs text-gray-400 pt-1">
                Last check: {new Date(data.timestamp).toLocaleTimeString()}
              </p>
            </div>
          ) : null}
        </div>
      )}
    </div>
  );
}
