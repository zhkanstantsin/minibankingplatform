# AGENTS.md - Frontend Implementation Guide

This document provides step-by-step instructions for AI agents implementing the Mini Banking Platform frontend.

## Overview

Build a minimalist React 19 frontend using Feature Sliced Design (FSD) architecture with React Query for server state. Focus on maximum component reuse - pages should compose high-level widgets, not contain raw HTML.

## Tech Stack

- **Framework**: React 19 with Vite
- **Language**: TypeScript (strict mode)
- **Styling**: Tailwind CSS
- **Server State**: TanStack Query (React Query) v5
- **Client State**: Zustand (only for auth token)
- **Routing**: React Router v7
- **Forms**: React Hook Form with Zod validation
- **HTTP Client**: Axios (wrapped in React Query)

## Feature Sliced Design Architecture

```
frontend/
├── src/
│   ├── app/                    # App initialization layer
│   │   ├── providers/          # All providers (Query, Router, Auth)
│   │   ├── router/             # Route configuration
│   │   ├── styles/             # Global styles
│   │   └── index.tsx           # App entry composition
│   │
│   ├── pages/                  # Page compositions (NO raw HTML here!)
│   │   ├── login/
│   │   ├── register/
│   │   ├── dashboard/
│   │   └── transactions/
│   │
│   ├── widgets/                # Self-contained UI blocks
│   │   ├── wallet-card/        # Account balance display
│   │   ├── transaction-list/   # Transaction list with items
│   │   ├── transfer-form/      # Complete transfer form
│   │   ├── exchange-form/      # Complete exchange form
│   │   ├── auth-form/          # Login/Register form (reused!)
│   │   └── header/             # App header with nav
│   │
│   ├── features/               # User interactions (business logic)
│   │   ├── auth/               # Login, register, logout
│   │   ├── transfer/           # Transfer money mutation
│   │   ├── exchange/           # Exchange currency mutation
│   │   └── transactions/       # Transaction filters, pagination
│   │
│   ├── entities/               # Business entities
│   │   ├── account/            # Account type, hooks, card component
│   │   ├── transaction/        # Transaction type, hooks, row component
│   │   └── user/               # User type, hooks
│   │
│   └── shared/                 # Shared code (no business logic)
│       ├── api/                # Axios instance, base config
│       ├── ui/                 # UI Kit components
│       ├── lib/                # Utilities, helpers
│       └── types/              # Global types
│
├── .env.example
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js
```

## FSD Layer Rules

| Layer | Can Import From | Cannot Import From |
|-------|-----------------|-------------------|
| app | pages, widgets, features, entities, shared | - |
| pages | widgets, features, entities, shared | app, other pages |
| widgets | features, entities, shared | app, pages, other widgets |
| features | entities, shared | app, pages, widgets, other features |
| entities | shared | app, pages, widgets, features, other entities |
| shared | - | anything above |

## Implementation Phases

### Phase 1: Project Setup

```bash
npm create vite@latest . -- --template react-ts
npm install @tanstack/react-query axios react-router-dom zustand react-hook-form @hookform/resolvers zod
npm install -D tailwindcss postcss autoprefixer @types/node
npx tailwindcss init -p
```

### Phase 2: Shared Layer (`src/shared/`)

#### 2.1 UI Kit (`shared/ui/`)

Create reusable, composable UI components. **These are the building blocks for everything.**

```typescript
// shared/ui/index.ts - Public API
export { Button } from './button';
export { Input } from './input';
export { Card, CardHeader, CardContent, CardFooter } from './card';
export { Select } from './select';
export { Alert } from './alert';
export { Spinner } from './spinner';
export { Badge } from './badge';
export { Table, TableHead, TableBody, TableRow, TableCell } from './table';
export { Pagination } from './pagination';
export { FormField } from './form-field';
export { PageLayout } from './page-layout';
export { Stack } from './stack';
export { MoneyDisplay } from './money-display';
```

**Key UI Components:**

```typescript
// shared/ui/card/Card.tsx
interface CardProps {
  children: React.ReactNode;
  className?: string;
}

// shared/ui/form-field/FormField.tsx
interface FormFieldProps {
  label: string;
  error?: string;
  children: React.ReactNode;
}

// shared/ui/money-display/MoneyDisplay.tsx
interface MoneyDisplayProps {
  amount: string;
  currency: 'USD' | 'EUR';
  size?: 'sm' | 'md' | 'lg';
  showSign?: boolean;
}

// shared/ui/page-layout/PageLayout.tsx
interface PageLayoutProps {
  title?: string;
  children: React.ReactNode;
  actions?: React.ReactNode;
}

// shared/ui/stack/Stack.tsx - Flex container helper
interface StackProps {
  direction?: 'row' | 'column';
  gap?: 'sm' | 'md' | 'lg';
  children: React.ReactNode;
}
```

#### 2.2 API Client (`shared/api/`)

```typescript
// shared/api/client.ts
import axios from 'axios';

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
});

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// shared/api/types.ts
export interface ProblemDetails {
  type: string;
  title: string;
  status: number;
  detail: string;
  instance?: string;
  [key: string]: unknown;
}

// shared/api/error-handler.ts
export function getErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error) && error.response?.data) {
    const problem = error.response.data as ProblemDetails;
    return problem.detail || problem.title || 'An error occurred';
  }
  return 'Network error';
}
```

#### 2.3 Types (`shared/types/`)

```typescript
// shared/types/api.ts
export type Currency = 'USD' | 'EUR';
export type TransactionType = 'transfer' | 'exchange' | 'deposit' | 'withdrawal';

export interface Money {
  amount: string;
  currency: Currency;
}

export interface Pagination {
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}
```

### Phase 3: Entities Layer (`src/entities/`)

Each entity has: **type**, **api**, **hooks**, **ui component**.

#### 3.1 Account Entity

```
entities/account/
├── index.ts              # Public API
├── model/
│   └── types.ts          # Account interface
├── api/
│   └── accountApi.ts     # API functions
├── hooks/
│   └── useAccounts.ts    # React Query hooks
└── ui/
    └── AccountCard.tsx   # Account display component
```

```typescript
// entities/account/model/types.ts
export interface Account {
  id: string;
  userId: string;
  balance: Money;
}

// entities/account/api/accountApi.ts
export const accountApi = {
  getAll: () => apiClient.get<Account[]>('/accounts').then(r => r.data),
  getBalance: (id: string) => apiClient.get<Balance>(`/accounts/${id}/balance`).then(r => r.data),
};

// entities/account/hooks/useAccounts.ts
export function useAccounts() {
  return useQuery({
    queryKey: ['accounts'],
    queryFn: accountApi.getAll,
  });
}

export function useAccountByCurrency(currency: Currency) {
  const { data: accounts } = useAccounts();
  return accounts?.find(a => a.balance.currency === currency);
}

// entities/account/ui/AccountCard.tsx
interface AccountCardProps {
  account: Account;
  onClick?: () => void;
}

export function AccountCard({ account, onClick }: AccountCardProps) {
  return (
    <Card onClick={onClick}>
      <CardContent>
        <Badge>{account.balance.currency}</Badge>
        <MoneyDisplay
          amount={account.balance.amount}
          currency={account.balance.currency}
          size="lg"
        />
      </CardContent>
    </Card>
  );
}

// entities/account/index.ts
export type { Account } from './model/types';
export { useAccounts, useAccountByCurrency } from './hooks/useAccounts';
export { AccountCard } from './ui/AccountCard';
```

#### 3.2 Transaction Entity

```
entities/transaction/
├── index.ts
├── model/
│   └── types.ts
├── api/
│   └── transactionApi.ts
├── hooks/
│   └── useTransactions.ts
└── ui/
    └── TransactionRow.tsx
```

```typescript
// entities/transaction/ui/TransactionRow.tsx
interface TransactionRowProps {
  transaction: Transaction;
}

export function TransactionRow({ transaction }: TransactionRowProps) {
  return (
    <TableRow>
      <TableCell>{formatDate(transaction.timestamp)}</TableCell>
      <TableCell><Badge>{transaction.type}</Badge></TableCell>
      <TableCell>{getTransactionDescription(transaction)}</TableCell>
      <TableCell>
        <MoneyDisplay {...getTransactionAmount(transaction)} showSign />
      </TableCell>
    </TableRow>
  );
}
```

#### 3.3 User Entity

```typescript
// entities/user/model/types.ts
export interface User {
  userId: string;
  email: string;
}

// entities/user/hooks/useCurrentUser.ts
export function useCurrentUser() {
  return useQuery({
    queryKey: ['currentUser'],
    queryFn: () => apiClient.get<User>('/auth/me').then(r => r.data),
    enabled: !!localStorage.getItem('token'),
  });
}
```

### Phase 4: Features Layer (`src/features/`)

Features contain **user interactions** - forms, actions, mutations.

#### 4.1 Auth Feature

```
features/auth/
├── index.ts
├── model/
│   ├── authStore.ts      # Zustand store for token only
│   └── schemas.ts        # Zod validation schemas
├── api/
│   └── authApi.ts
└── hooks/
    ├── useLogin.ts
    ├── useRegister.ts
    └── useLogout.ts
```

```typescript
// features/auth/model/authStore.ts
interface AuthStore {
  token: string | null;
  setToken: (token: string | null) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthStore>((set) => ({
  token: localStorage.getItem('token'),
  setToken: (token) => {
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
    set({ token });
  },
  logout: () => {
    localStorage.removeItem('token');
    set({ token: null });
  },
}));

// features/auth/hooks/useLogin.ts
export function useLogin() {
  const setToken = useAuthStore(s => s.setToken);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: LoginRequest) => authApi.login(data),
    onSuccess: (response) => {
      setToken(response.token);
      queryClient.invalidateQueries({ queryKey: ['currentUser'] });
    },
  });
}

// features/auth/model/schemas.ts
export const loginSchema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(1, 'Password required'),
});

export const registerSchema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(8, 'Minimum 8 characters'),
});
```

#### 4.2 Transfer Feature

```
features/transfer/
├── index.ts
├── model/
│   └── schemas.ts
├── hooks/
│   └── useTransfer.ts
└── ui/
    └── TransferFields.tsx   # Form fields only, not full form
```

```typescript
// features/transfer/hooks/useTransfer.ts
export function useTransfer() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: TransferRequest) =>
      apiClient.post<TransferResponse>('/transactions/transfer', data).then(r => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
    },
  });
}

// features/transfer/ui/TransferFields.tsx
// Just the form fields, reusable in any form context
interface TransferFieldsProps {
  control: Control<TransferFormData>;
  accounts: Account[];
}

export function TransferFields({ control, accounts }: TransferFieldsProps) {
  return (
    <Stack direction="column" gap="md">
      <FormField label="From Account" error={errors.fromAccountId?.message}>
        <Controller
          name="fromAccountId"
          control={control}
          render={({ field }) => (
            <Select {...field} options={accounts.map(a => ({
              value: a.id,
              label: `${a.balance.currency} - ${formatMoney(a.balance)}`
            }))} />
          )}
        />
      </FormField>
      {/* ... other fields */}
    </Stack>
  );
}
```

#### 4.3 Exchange Feature

```typescript
// features/exchange/hooks/useExchange.ts
export function useExchange() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: ExchangeRequest) =>
      apiClient.post<ExchangeResponse>('/transactions/exchange', data).then(r => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
    },
  });
}

// features/exchange/hooks/useExchangeCalculation.ts
export function useExchangeCalculation(amount: string, from: Currency, to: Currency) {
  return useQuery({
    queryKey: ['exchangeCalculation', amount, from, to],
    queryFn: () => apiClient.get<ExchangeCalculation>('/transactions/exchange/calculate', {
      params: { amount, sourceCurrency: from, targetCurrency: to },
    }).then(r => r.data),
    enabled: !!amount && parseFloat(amount) > 0 && from !== to,
  });
}
```

#### 4.4 Transactions Feature (filters, pagination)

```typescript
// features/transactions/model/types.ts
export interface TransactionFilters {
  type?: TransactionType;
  page: number;
  limit: number;
}

// features/transactions/hooks/useTransactionFilters.ts
export function useTransactionFilters() {
  const [filters, setFilters] = useState<TransactionFilters>({
    page: 1,
    limit: 20,
  });

  return {
    filters,
    setType: (type?: TransactionType) => setFilters(f => ({ ...f, type, page: 1 })),
    setPage: (page: number) => setFilters(f => ({ ...f, page })),
    setLimit: (limit: number) => setFilters(f => ({ ...f, limit, page: 1 })),
  };
}
```

### Phase 5: Widgets Layer (`src/widgets/`)

Widgets are **self-contained UI blocks** that compose features and entities.

#### 5.1 Wallet Card Widget (reuses AccountCard)

```
widgets/wallet-card/
├── index.ts
└── ui/
    └── WalletCard.tsx
```

```typescript
// widgets/wallet-card/ui/WalletCard.tsx
interface WalletCardProps {
  currency: Currency;
}

export function WalletCard({ currency }: WalletCardProps) {
  const { data: accounts, isLoading } = useAccounts();
  const account = accounts?.find(a => a.balance.currency === currency);

  if (isLoading) return <Card><Spinner /></Card>;
  if (!account) return null;

  return <AccountCard account={account} />;
}
```

#### 5.2 Auth Form Widget (REUSED for login AND register!)

```
widgets/auth-form/
├── index.ts
└── ui/
    └── AuthForm.tsx
```

```typescript
// widgets/auth-form/ui/AuthForm.tsx
interface AuthFormProps {
  mode: 'login' | 'register';
  onSuccess?: () => void;
}

export function AuthForm({ mode, onSuccess }: AuthFormProps) {
  const login = useLogin();
  const register = useRegister();
  const mutation = mode === 'login' ? login : register;
  const schema = mode === 'login' ? loginSchema : registerSchema;

  const form = useForm({
    resolver: zodResolver(schema),
  });

  const onSubmit = form.handleSubmit((data) => {
    mutation.mutate(data, { onSuccess });
  });

  return (
    <Card>
      <CardHeader>
        <h2>{mode === 'login' ? 'Sign In' : 'Create Account'}</h2>
      </CardHeader>
      <CardContent>
        <form onSubmit={onSubmit}>
          <Stack direction="column" gap="md">
            <FormField label="Email" error={form.formState.errors.email?.message}>
              <Input type="email" {...form.register('email')} />
            </FormField>
            <FormField label="Password" error={form.formState.errors.password?.message}>
              <Input type="password" {...form.register('password')} />
            </FormField>
            {mutation.isError && (
              <Alert variant="error">{getErrorMessage(mutation.error)}</Alert>
            )}
            <Button type="submit" loading={mutation.isPending}>
              {mode === 'login' ? 'Sign In' : 'Create Account'}
            </Button>
          </Stack>
        </form>
      </CardContent>
      <CardFooter>
        <Link to={mode === 'login' ? '/register' : '/login'}>
          {mode === 'login' ? 'Need an account?' : 'Already have an account?'}
        </Link>
      </CardFooter>
    </Card>
  );
}
```

#### 5.3 Transfer Form Widget

```typescript
// widgets/transfer-form/ui/TransferForm.tsx
export function TransferForm() {
  const { data: accounts } = useAccounts();
  const transfer = useTransfer();
  const form = useForm<TransferFormData>({
    resolver: zodResolver(transferSchema),
  });

  const onSubmit = form.handleSubmit((data) => {
    transfer.mutate(data, {
      onSuccess: () => {
        form.reset();
        // Show success toast or message
      },
    });
  });

  return (
    <Card>
      <CardHeader>Transfer Money</CardHeader>
      <CardContent>
        <form onSubmit={onSubmit}>
          <Stack direction="column" gap="md">
            <TransferFields control={form.control} accounts={accounts || []} />
            {transfer.isError && (
              <Alert variant="error">{getErrorMessage(transfer.error)}</Alert>
            )}
            {transfer.isSuccess && (
              <Alert variant="success">Transfer completed!</Alert>
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
```

#### 5.4 Exchange Form Widget

```typescript
// widgets/exchange-form/ui/ExchangeForm.tsx
export function ExchangeForm() {
  const { data: accounts } = useAccounts();
  const exchange = useExchange();
  const form = useForm<ExchangeFormData>({
    resolver: zodResolver(exchangeSchema),
  });

  const watchAmount = form.watch('amount');
  const watchSource = form.watch('sourceAccountId');
  const sourceAccount = accounts?.find(a => a.id === watchSource);
  const targetCurrency = sourceAccount?.balance.currency === 'USD' ? 'EUR' : 'USD';

  const { data: calculation } = useExchangeCalculation(
    watchAmount,
    sourceAccount?.balance.currency || 'USD',
    targetCurrency
  );

  return (
    <Card>
      <CardHeader>Exchange Currency</CardHeader>
      <CardContent>
        <form onSubmit={form.handleSubmit(data => exchange.mutate(data))}>
          <Stack direction="column" gap="md">
            <ExchangeFields control={form.control} accounts={accounts || []} />

            {calculation && (
              <Alert variant="info">
                You will receive: <MoneyDisplay {...calculation.targetAmount} />
                <br />
                Rate: 1 {calculation.exchangeRate.sourceCurrency} = {calculation.exchangeRate.rate} {calculation.exchangeRate.targetCurrency}
              </Alert>
            )}

            {exchange.isError && (
              <Alert variant="error">{getErrorMessage(exchange.error)}</Alert>
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
```

#### 5.5 Transaction List Widget (reuses TransactionRow)

```typescript
// widgets/transaction-list/ui/TransactionList.tsx
interface TransactionListProps {
  limit?: number;
  showFilters?: boolean;
  showPagination?: boolean;
}

export function TransactionList({
  limit = 20,
  showFilters = false,
  showPagination = false
}: TransactionListProps) {
  const { filters, setType, setPage } = useTransactionFilters();
  const { data, isLoading } = useTransactions({ ...filters, limit });

  if (isLoading) return <Spinner />;

  return (
    <Card>
      <CardHeader>
        <Stack direction="row" justify="between">
          <span>Transactions</span>
          {showFilters && (
            <Select
              value={filters.type || ''}
              onChange={(e) => setType(e.target.value as TransactionType || undefined)}
              options={[
                { value: '', label: 'All' },
                { value: 'transfer', label: 'Transfers' },
                { value: 'exchange', label: 'Exchanges' },
              ]}
            />
          )}
        </Stack>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Date</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>Details</TableCell>
              <TableCell>Amount</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data?.transactions.map(tx => (
              <TransactionRow key={tx.id} transaction={tx} />
            ))}
          </TableBody>
        </Table>
      </CardContent>
      {showPagination && data?.pagination && (
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
```

#### 5.6 Header Widget

```typescript
// widgets/header/ui/Header.tsx
export function Header() {
  const { data: user } = useCurrentUser();
  const logout = useLogout();

  return (
    <header className="border-b bg-white">
      <Stack direction="row" justify="between" className="container mx-auto p-4">
        <Link to="/" className="font-bold text-xl">Mini Bank</Link>
        <nav>
          <Stack direction="row" gap="md">
            <Link to="/">Dashboard</Link>
            <Link to="/transactions">History</Link>
            <span>{user?.email}</span>
            <Button variant="secondary" onClick={() => logout.mutate()}>
              Logout
            </Button>
          </Stack>
        </nav>
      </Stack>
    </header>
  );
}
```

### Phase 6: Pages Layer (`src/pages/`)

**Pages only compose widgets. NO raw HTML, NO direct API calls.**

```typescript
// pages/login/LoginPage.tsx
export function LoginPage() {
  const navigate = useNavigate();
  return (
    <PageLayout>
      <AuthForm mode="login" onSuccess={() => navigate('/')} />
    </PageLayout>
  );
}

// pages/register/RegisterPage.tsx
export function RegisterPage() {
  const navigate = useNavigate();
  return (
    <PageLayout>
      <AuthForm mode="register" onSuccess={() => navigate('/')} />
    </PageLayout>
  );
}

// pages/dashboard/DashboardPage.tsx
export function DashboardPage() {
  return (
    <PageLayout title="Dashboard">
      <Stack direction="column" gap="lg">
        {/* Wallets */}
        <Stack direction="row" gap="md">
          <WalletCard currency="USD" />
          <WalletCard currency="EUR" />
        </Stack>

        {/* Actions */}
        <Stack direction="row" gap="md">
          <TransferForm />
          <ExchangeForm />
        </Stack>

        {/* Recent Transactions */}
        <TransactionList limit={5} />
      </Stack>
    </PageLayout>
  );
}

// pages/transactions/TransactionsPage.tsx
export function TransactionsPage() {
  return (
    <PageLayout title="Transaction History">
      <TransactionList showFilters showPagination />
    </PageLayout>
  );
}
```

### Phase 7: App Layer (`src/app/`)

```typescript
// app/providers/QueryProvider.tsx
export function QueryProvider({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(() => new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 1000 * 60, // 1 minute
        retry: 1,
      },
    },
  }));

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}

// app/providers/AuthProvider.tsx
export function AuthProvider({ children }: { children: React.ReactNode }) {
  const token = useAuthStore(s => s.token);
  const { isLoading, isError } = useCurrentUser();

  // Handle auth state...
  return <>{children}</>;
}

// app/router/routes.tsx
export const routes = [
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/register',
    element: <RegisterPage />,
  },
  {
    element: <ProtectedLayout />,
    children: [
      { path: '/', element: <DashboardPage /> },
      { path: '/transactions', element: <TransactionsPage /> },
    ],
  },
];

// app/layouts/ProtectedLayout.tsx
export function ProtectedLayout() {
  const token = useAuthStore(s => s.token);

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return (
    <>
      <Header />
      <main className="container mx-auto p-4">
        <Outlet />
      </main>
    </>
  );
}

// app/index.tsx
export function App() {
  return (
    <QueryProvider>
      <BrowserRouter>
        <Routes>
          {routes.map(route => (
            <Route key={route.path} {...route} />
          ))}
        </Routes>
      </BrowserRouter>
    </QueryProvider>
  );
}
```

## Component Reuse Summary

| Component | Used In |
|-----------|---------|
| `Card` | WalletCard, AuthForm, TransferForm, ExchangeForm, TransactionList |
| `Stack` | Every page, every widget, every form |
| `FormField` | AuthForm, TransferForm, ExchangeForm |
| `Button` | AuthForm, TransferForm, ExchangeForm, Header |
| `Alert` | AuthForm, TransferForm, ExchangeForm |
| `MoneyDisplay` | AccountCard, TransactionRow, ExchangeForm |
| `Table*` | TransactionList |
| `Select` | TransferForm, ExchangeForm, TransactionList filters |
| `AuthForm` | LoginPage, RegisterPage |
| `TransactionList` | DashboardPage (limit=5), TransactionsPage (full) |
| `TransactionRow` | TransactionList |

## Testing Checklist

1. [ ] Pages contain NO raw HTML, only widget compositions
2. [ ] AuthForm is reused for both login and register
3. [ ] TransactionList is reused on dashboard and history page
4. [ ] All forms use shared FormField component
5. [ ] All cards use shared Card component
6. [ ] Money is always displayed via MoneyDisplay
7. [ ] React Query handles all server state
8. [ ] Zustand only stores auth token
9. [ ] FSD layer rules are respected (imports only go down)

## Key Principles

1. **Pages are compositions** - They import widgets and arrange them
2. **Widgets are self-contained** - They fetch their own data if needed
3. **Features contain business logic** - Mutations, validations, filters
4. **Entities are reusable** - Types, hooks, and basic UI for each entity
5. **Shared is framework-agnostic** - UI kit, API client, utilities
6. **React Query for server state** - No manual loading/error state management
7. **Maximum reuse** - If you write HTML twice, extract a component
