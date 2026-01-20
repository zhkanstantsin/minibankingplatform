# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a mini banking platform with a Go backend and React frontend. The backend follows Domain-Driven Design (DDD) principles with double-entry bookkeeping. The system handles money transfers with strong consistency guarantees through ledger validation. The frontend is a minimalist React 19 application focused on functionality and correct data flow.

## Project Structure

```
minibankingplatform/
├── backend/           # Go backend application
│   ├── api/           # OpenAPI spec and generated code config
│   ├── cmd/server/    # Application entry point
│   ├── internal/      # Private application code
│   │   ├── api/       # HTTP handlers and middleware
│   │   ├── domain/    # Business logic and entities
│   │   ├── infrastructure/  # Database repositories
│   │   └── service/   # Application services
│   ├── migrations/    # Database migrations
│   ├── pkg/           # Public packages (jwt, trm)
│   ├── Dockerfile     # Backend container image
│   ├── go.mod
│   └── go.sum
├── frontend/          # React frontend (Feature Sliced Design)
│   ├── src/
│   │   ├── app/       # App layer: providers, router, layouts
│   │   ├── pages/     # Page compositions (only widgets, no raw HTML)
│   │   ├── widgets/   # Self-contained UI blocks (forms, lists)
│   │   ├── features/  # User interactions (auth, transfer, exchange)
│   │   ├── entities/  # Business entities (account, transaction, user)
│   │   └── shared/    # UI kit, API client, utilities
│   ├── Dockerfile     # Frontend container image
│   └── AGENTS.md      # Detailed implementation guide for AI agents
├── docker-compose.yml # Full stack orchestration
└── Taskfile.yml       # Task automation
```

## Common Commands

### Docker (Full Stack)
```bash
# Start all services (postgres, backend, frontend)
task docker:up

# Stop all services
task docker:down

# Build all docker images
task docker:build

# View logs
task docker:logs
task docker:logs:backend
task docker:logs:frontend
```

### Database Management
```bash
# Start PostgreSQL container only
docker-compose up -d postgres

# Apply all migrations
task migrate:up

# Apply one migration
task migrate:up:one

# Rollback one migration
task migrate:down:one

# Create new migration
task migrate:create -- migration_name

# Check current migration version
task migrate:version

# Full database reset
task db:reset
```

### Backend Development
```bash
# Run the backend server locally
task backend:run

# Build the backend binary
task backend:build

# Run all tests
task backend:test

# Run tests with verbose output
task backend:test:verbose

# Or run Go commands directly from backend/
cd backend
go test ./...
go test -v -run TestTransfer ./internal/service
go run ./cmd/server
```

Tests use testcontainers-go to spin up PostgreSQL automatically. Each test suite applies migrations programmatically.

### Code Generation
```bash
# Generate all code (enums + API)
task generate

# Generate API code from OpenAPI spec
task api:generate
```

The project uses `go-enum` to generate methods for enums defined with `// ENUM(...)` comments in domain types.

## Architecture

### Domain Layer (`backend/internal/domain/`)

Core business logic with no external dependencies:

- **Money**: Value object with `decimal.Decimal` amounts and `Currency` enum (USD, EUR). All operations validate currency matching.
- **Account**: Aggregate root representing user accounts. Has `Credit()` and `Debit()` methods that modify balance.
- **TransferService**: Domain service that orchestrates transfers between accounts, creating `TransferDetails`.
- **Ledger**: Double-entry bookkeeping system. Each transfer creates two `LedgerRecord` entries (debit and credit) that sum to zero.

### Infrastructure Layer (`backend/internal/infrastructure/`)

Database repositories that implement the DBTX interface:

- All repositories accept either `pgx.Tx` (transaction) or `*pgxpool.Pool` (direct connection)
- Queries use `GetForUpdate()` for row-level locking during transfers
- `LedgerRepository` provides validation queries: `GetTotalBalanceByCurrency()` and `GetAccountBalanceMismatches()`

### Service Layer (`backend/internal/service/`)

Application services coordinating domain logic with infrastructure:

- `Service.Transfer()` wraps the entire transfer operation in a database transaction
- After each transfer, validates:
  1. Ledger balance by currency equals zero
  2. Each account's balance matches its ledger sum

### Transaction Management (`backend/pkg/trm/`)

Custom transaction manager implementation:

- `TransactionManager[Tx, Opts]`: Generic wrapper for database transactions
- `Injector[T]`: Context-based transaction injection - repositories get the active transaction from context
- `Do()` and `DoTx()`: Execute functions within transactions with automatic commit/rollback

**Pattern**: Services call `trm.Do(ctx, func(ctx context.Context) error {...})`. The transaction is injected into context, and repositories retrieve it via `Injector.DB(ctx)`.

### Key Invariants

The system enforces strict double-entry bookkeeping:

1. **Zero-sum per transaction**: All ledger entries for a transaction must sum to zero
2. **Zero-sum per currency**: Total of all ledger entries in each currency must be zero
3. **Account-ledger consistency**: Each account's balance must equal the sum of its ledger entries

These are verified after every transfer in `backend/internal/service/service.go`.

## Database Schema

Key tables:
- `users`: User accounts
- `accounts`: Financial accounts with balance and currency
- `transactions`: Base transaction records (transfer, exchange, deposit, withdrawal)
- `transfer_details`: Specific transfer information
- `ledger`: Double-entry bookkeeping records

Enums: `currency` (USD, EUR), `transaction_type`

## Environment Variables

Backend environment variables are configured in `backend/.env`:
- `POSTGRES_USER` - Database user
- `POSTGRES_PASSWORD` - Database password
- `POSTGRES_DB` - Database name
- `POSTGRES_HOST` - Database host (use `localhost` for local development, `postgres` for Docker)
- `POSTGRES_PORT` - Database port (default: 5432)
- `JWT_SECRET` - Secret key for JWT token generation and validation
- `DATABASE_URL` - Full PostgreSQL connection string for migrations

To set up:
```bash
cp backend/.env.example backend/.env
# Edit backend/.env with your values
```

Frontend `.env` (in `frontend/` directory):
- `VITE_API_URL` - Backend API URL (default: `http://localhost:8080/api/v1`)

## Ports

- **8080**: Backend API
- **3000**: Frontend
- **5432**: PostgreSQL

---

## Frontend

### Tech Stack

- **Framework**: React 19 + Vite
- **Styling**: Tailwind CSS
- **Architecture**: Feature Sliced Design (FSD)
- **Server State**: TanStack Query (React Query) v5
- **Client State**: Zustand (auth token only)
- **HTTP Client**: Axios with JWT interceptors
- **Routing**: React Router v7
- **Forms**: React Hook Form + Zod validation
- **TypeScript**: Strict mode enabled

### Frontend Commands

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Type check
npm run typecheck

# Lint
npm run lint
```

Or use Taskfile from project root:
```bash
task frontend:dev      # Start dev server
task frontend:build    # Production build
task frontend:install  # Install dependencies
```

### Feature Sliced Design Architecture

The frontend follows FSD - a layered architecture for scalable frontend applications.

#### Layer Hierarchy (imports only go DOWN)

```
app → pages → widgets → features → entities → shared
```

| Layer | Purpose | Examples |
|-------|---------|----------|
| **app** | App initialization, providers, router | QueryProvider, ProtectedLayout |
| **pages** | Page compositions (NO raw HTML) | DashboardPage, TransactionsPage |
| **widgets** | Self-contained UI blocks | WalletCard, TransferForm, AuthForm |
| **features** | User interactions, mutations | useTransfer, useExchange, useLogin |
| **entities** | Business objects, queries | Account, Transaction, User |
| **shared** | UI kit, utilities, API client | Button, Card, apiClient |

#### Key Layers

**shared/ui/** - Reusable UI Kit:
- `Button`, `Input`, `Select` - Form controls
- `Card`, `CardHeader`, `CardContent`, `CardFooter` - Container
- `FormField` - Label + input + error wrapper
- `Stack` - Flex container helper
- `MoneyDisplay` - Currency-aware money formatting
- `Table`, `Pagination`, `Alert`, `Badge`, `Spinner`

**entities/** - Business objects with types, hooks, UI:
- `account/`: Account type, `useAccounts()`, `AccountCard`
- `transaction/`: Transaction type, `useTransactions()`, `TransactionRow`
- `user/`: User type, `useCurrentUser()`

**features/** - User interactions:
- `auth/`: `useLogin()`, `useRegister()`, `useLogout()`, authStore
- `transfer/`: `useTransfer()` mutation, transfer schemas
- `exchange/`: `useExchange()`, `useExchangeCalculation()`
- `transactions/`: Filters, pagination state

**widgets/** - Complete UI blocks:
- `auth-form/`: **Reused** for login AND register pages
- `wallet-card/`: Account balance display
- `transfer-form/`: Complete transfer form with validation
- `exchange-form/`: Exchange form with live rate preview
- `transaction-list/`: **Reused** on dashboard (limit=5) and history (full)
- `header/`: Navigation with user info

**pages/** - Pure compositions:
```typescript
// Pages compose widgets, NO raw HTML
export function DashboardPage() {
  return (
    <PageLayout title="Dashboard">
      <Stack direction="column" gap="lg">
        <Stack direction="row" gap="md">
          <WalletCard currency="USD" />
          <WalletCard currency="EUR" />
        </Stack>
        <Stack direction="row" gap="md">
          <TransferForm />
          <ExchangeForm />
        </Stack>
        <TransactionList limit={5} />
      </Stack>
    </PageLayout>
  );
}
```

### Component Reuse Principle

**Every UI element is a reusable component.** Pages never contain raw HTML.

| Component | Reused In |
|-----------|-----------|
| `AuthForm` | LoginPage, RegisterPage |
| `TransactionList` | DashboardPage, TransactionsPage |
| `Card` | All widgets |
| `FormField` | All forms |
| `MoneyDisplay` | AccountCard, TransactionRow, ExchangeForm |

### React Query Usage

All server state is managed by React Query:

```typescript
// Queries (GET)
useAccounts()           // Fetch user accounts
useTransactions(filters) // Fetch transactions
useCurrentUser()        // Fetch current user
useExchangeCalculation() // Calculate exchange preview

// Mutations (POST)
useLogin()    // Login mutation
useRegister() // Register mutation
useTransfer() // Transfer mutation (invalidates accounts, transactions)
useExchange() // Exchange mutation (invalidates accounts, transactions)
```

### Key Features

1. **Dashboard Page**:
   - USD and EUR wallet balances (WalletCard x2)
   - Last 5 transactions (TransactionList limit=5)
   - Transfer form (TransferForm widget)
   - Exchange form with live calculation (ExchangeForm widget)

2. **Transaction History**:
   - Full transaction list (TransactionList showFilters showPagination)
   - Filter by type (transfer, exchange)
   - Pagination controls

3. **Error Handling**:
   - RFC 7807 Problem Details parsing
   - User-friendly error messages in Alert components

4. **Authentication**:
   - JWT in localStorage + Zustand sync
   - ProtectedLayout redirects to login
   - React Query enabled flag for auth queries

### Exchange Rate

Fixed rate: **1 USD = 0.92 EUR**

Exchange form calls `/transactions/exchange/calculate` for live preview.

### API Integration

Backend base URL: `http://localhost:8080/api/v1`

All authenticated endpoints require `Authorization: Bearer <token>` header.

See `backend/api/swagger.yaml` for full API specification.
See `frontend/AGENTS.md` for detailed implementation guide.
