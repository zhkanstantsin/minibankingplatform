# Mini Banking Platform

A full-stack mini banking platform with a **Go backend** implementing **double-entry bookkeeping** and a **React 19 frontend** following **Feature Sliced Design (FSD)** architecture. The system handles money transfers and currency exchanges with strong consistency guarantees through ledger validation.

## Table of Contents

- [Setup Instructions](#setup-instructions)
- [User Management Approach](#user-management-approach)
- [Architecture Overview](#architecture-overview)
- [Double-Entry Ledger Design](#double-entry-ledger-design)
- [Design Decisions and Trade-offs](#design-decisions-and-trade-offs)
- [Known Limitations](#known-limitations)
- [Technical Questions Answered](#technical-questions-answered)
- [Technical Evaluation Checklist](#technical-evaluation-checklist)

---

## Setup Instructions

### Prerequisites

- **Docker** and **Docker Compose** (for containerized setup)
- **Go 1.22+** (for local backend development)
- **Node.js 20+** and **npm** (for local frontend development)
- **Task** (task runner) - optional but recommended
- **golang-migrate** CLI (for database migrations)

### Quick Start with Docker (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd minibankingplatform
   ```

2. **Create backend environment file**
   ```bash
   cp backend/.env.example backend/.env
   ```

3. **Configure environment variables** in `backend/.env`:
   ```env
   # Database Configuration
   POSTGRES_USER=bankuser
   POSTGRES_PASSWORD=bankpass123
   POSTGRES_DB=minibankingdb
   POSTGRES_HOST=localhost
   POSTGRES_PORT=5432

   # JWT Configuration
   JWT_SECRET=your_jwt_secret_key_change_this_in_production

   # Database URL for migrations
   DATABASE_URL=postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable
   ```

4. **Start all services**
   ```bash
   task docker:up
   # Or without Task:
   docker-compose up -d
   ```

5. **Apply database migrations**
   ```bash
   task migrate:up
   ```

6. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080

### Local Development Setup

#### Backend

```bash
cd backend

# Create environment file
cp .env.example .env

# Configure environment variables in backend/.env
# Make sure to set POSTGRES_HOST=localhost for local development
# and update JWT_SECRET to a secure random value

# Start PostgreSQL (if not using Docker)
cd ..
docker-compose up -d postgres

# Apply migrations
task migrate:up

# Run the server
task backend:run
# Or: cd backend && go run ./cmd/server
```

#### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Create environment file
cp .env.example .env

# Start development server
npm run dev
```

### Running Tests

```bash
# Backend tests (uses testcontainers for PostgreSQL)
task backend:test

# Frontend type checking
cd frontend && npm run typecheck
```

---

## User Management Approach

### Chosen Approach: Registration

This project implements **user registration** (not pre-seeded users). New users can register through the `/auth/register` endpoint or the frontend registration form.

### How It Works

1. User provides email and password (minimum 8 characters)
2. System creates the user with hashed password (bcrypt)
3. System automatically creates **two accounts** for the user:
   - USD account with **$1,000.00** initial balance
   - EUR account with **€500.00** initial balance
4. Initial balances are funded from the **system cashbook** to maintain double-entry integrity
5. JWT token is returned for immediate authentication

### Why Registration?

- **Simplicity**: No need for separate seed scripts or pre-configured test accounts
- **Realistic flow**: Mimics real-world banking onboarding
- **Clean data**: Each test run can start fresh without conflicting data
- **Self-contained**: All user setup happens within a single atomic transaction

---

## Architecture Overview

### Backend (Go)

```
backend/
├── api/                    # OpenAPI spec (swagger.yaml)
├── cmd/server/             # Application entry point
├── internal/
│   ├── api/                # HTTP handlers, middleware, validation
│   ├── domain/             # Business logic, entities, value objects
│   ├── infrastructure/     # Database repositories
│   └── service/            # Application services (orchestration)
├── migrations/             # SQL migrations
└── pkg/
    ├── jwt/                # JWT utilities
    └── trm/                # Transaction manager
```

### Frontend (React)

```
frontend/src/
├── app/                    # Providers, router, layouts
├── pages/                  # Page compositions (widgets only)
├── widgets/                # Self-contained UI blocks
├── features/               # User interactions (mutations)
├── entities/               # Business objects (queries)
└── shared/                 # UI kit, API client, utilities
```

---

## Double-Entry Ledger Design

### Core Concept

Every financial operation creates **paired ledger entries** that sum to zero, ensuring the accounting equation always balances:

```
Assets = Liabilities + Equity
```

For every debit entry, there's an equal credit entry. The system enforces this invariant at the database and application layers.

### Database Schema

```sql
CREATE TABLE ledger (
    id UUID PRIMARY KEY,
    transaction UUID NOT NULL REFERENCES transactions(id),
    account UUID NOT NULL REFERENCES accounts(id),
    amount DECIMAL(19, 4) NOT NULL,  -- Positive = credit, Negative = debit
    currency currency NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_ledger_transaction ON ledger(transaction);
CREATE INDEX idx_ledger_account ON ledger(account);
CREATE INDEX idx_ledger_timestamp ON ledger(timestamp);
```

### Ledger Entry Examples

**Transfer: User A sends $100 to User B**

| Account | Amount | Entry Type |
|---------|--------|------------|
| User A USD | -100.00 | Debit |
| User B USD | +100.00 | Credit |
| **Sum** | **0.00** | ✓ Balanced |

**Exchange: User exchanges $100 USD for EUR (rate: 0.92)**

| Account | Currency | Amount | Entry Type |
|---------|----------|--------|------------|
| User USD Account | USD | -100.00 | Debit |
| Cashbook USD | USD | +100.00 | Credit |
| Cashbook EUR | EUR | -92.00 | Debit |
| User EUR Account | EUR | +92.00 | Credit |
| **USD Sum** | | **0.00** | ✓ Balanced |
| **EUR Sum** | | **0.00** | ✓ Balanced |

### Cashbook System

For currency exchanges, the system uses **cashbook accounts** (one per currency). These are special system accounts that act as the counterparty for cross-currency operations:

```sql
-- System user (cannot login)
INSERT INTO users (id, email, password_hash)
VALUES ('00000000-0000-0000-0000-000000000001', 'system@cashbook.internal', 'SYSTEM_ACCOUNT_NO_LOGIN');

-- USD Cashbook
INSERT INTO accounts (id, user_id, balance, currency)
VALUES ('00000000-0000-0000-0000-000000000010', '00000000-...', 0, 'USD');

-- EUR Cashbook
INSERT INTO accounts (id, user_id, balance, currency)
VALUES ('00000000-0000-0000-0000-000000000011', '00000000-...', 0, 'EUR');
```

### Three Key Invariants

The system enforces these invariants after every transaction:

1. **Zero-sum per currency**: `SUM(ledger.amount) GROUP BY currency = 0`
2. **Account-ledger consistency**: `account.balance = SUM(ledger.amount WHERE account = account.id)`
3. **Positive amounts only**: Transfer/exchange amounts must be > 0

```go
// From transfer.go - invariant checks within the transaction
err = s.CheckLedgerBalanceByCurrency(ctx)        // Invariant 1
err = s.checkAccountLedgerConsistency(ctx, from) // Invariant 2
err = s.checkAccountLedgerConsistency(ctx, to)   // Invariant 2
```

### Reconciliation Endpoint

The system provides a `/system/reconcile` endpoint that:

1. Calculates total ledger balance per currency (should all be zero)
2. Compares each account's balance with its ledger sum
3. Returns a detailed report with any mismatches

```json
{
  "timestamp": "2025-01-20T10:30:00Z",
  "isConsistent": true,
  "ledgerBalances": [
    { "currency": "USD", "totalSum": "0.00", "isBalanced": true },
    { "currency": "EUR", "totalSum": "0.00", "isBalanced": true }
  ],
  "accountMismatches": [],
  "totalAccountsChecked": 42
}
```

---

## Design Decisions and Trade-offs

### 1. Domain-Driven Design (DDD)

**Decision**: Implement core business logic in a pure domain layer.

**Benefits**:
- Business rules are isolated and testable
- No dependencies on infrastructure
- Clear separation of concerns

**Trade-off**: More code and abstractions compared to a simple CRUD approach.

### 2. Generic Transaction Manager

**Decision**: Custom `trm.TransactionManager[Tx, Opts]` with context-based injection.

**Benefits**:
- Type-safe transaction handling
- Repositories don't need to know about transactions
- Automatic commit/rollback

**Trade-off**: Additional complexity vs using raw `pgx.Tx` directly.

### 3. Fixed Exchange Rate

**Decision**: Use fixed rate (1 USD = 0.92 EUR) instead of external API.

**Benefits**:
- Predictable behavior for testing
- No external dependencies
- Simpler implementation

**Trade-off**: Not suitable for production without modification.

### 4. Feature Sliced Design (Frontend)

**Decision**: Strict FSD architecture with layered imports.

**Benefits**:
- Maximum component reuse
- Clear dependency direction
- Scalable structure

**Trade-off**: Learning curve, more files/folders.

### 5. React Query for Server State

**Decision**: TanStack Query instead of Redux/context for server state.

**Benefits**:
- Built-in caching, refetching, invalidation
- No manual loading/error state
- Optimistic updates support

**Trade-off**: Another dependency, specific patterns to learn.

### 6. Pessimistic Locking (SELECT FOR UPDATE)

**Decision**: Use row-level locking for concurrent access.

**Benefits**:
- Simple correctness guarantees
- No lost updates
- PostgreSQL native support

**Trade-off**: Potential deadlocks (mitigated by consistent lock order), reduced throughput under high contention.

---

## Known Limitations

### Backend

1. **Fixed exchange rate**: The rate is hardcoded (1 USD = 0.92 EUR). Production would need an external rate provider.

2. **Two currencies only**: System only supports USD and EUR. Adding more currencies requires:
   - Database enum update
   - New cashbook accounts
   - Exchange rate matrix

3. **No deposit/withdrawal**: While the schema supports these transaction types, they're not implemented. Initial balances come from registration.

4. **Single database**: No read replicas or sharding. Would need horizontal scaling for millions of users.

5. **No rate limiting**: API endpoints lack rate limiting protection.

6. **Basic password hashing**: Uses bcrypt, but no password complexity rules beyond minimum length.

### Frontend

1. **No real-time updates**: Balances update on page refresh or after mutations, not via WebSockets.

2. **Limited transaction details modal**: Shows basic info without full audit trail.

3. **No transaction receipts**: Can't export or print transaction confirmations.

4. **Basic responsive design**: Works on mobile but not optimized.

5. **No accessibility audit**: ARIA labels present but no full a11y compliance testing.

### Incomplete Features Due to Time Constraints

1. **WebSocket real-time updates**: Planned but not implemented.
2. **Transaction receipts/export**: Schema supports it, UI not built.
3. **Audit log UI**: Backend has full ledger trail, no admin UI to view it.
4. **Password reset**: Not implemented.
5. **Account statements**: No monthly/periodic statement generation.

---

## Technical Questions Answered

### How do you ensure transaction atomicity?

**Database transactions with automatic rollback:**

```go
func (s *Service) Transfer(ctx context.Context, cmd *TransferCommand) error {
    err := s.trm.Do(ctx, func(ctx context.Context) error {
        // All operations here are in a single transaction
        from, _ := s.accounts.GetForUpdate(ctx, cmd.From)
        to, _ := s.accounts.GetForUpdate(ctx, cmd.To)
        
        // Domain logic
        details, _ := s.transfer.Execute(from, to, cmd.Money, cmd.Time)
        
        // Persist changes
        s.transfers.Insert(ctx, details)
        s.accounts.Save(ctx, from)
        s.accounts.Save(ctx, to)
        
        // Validate invariants
        s.CheckLedgerBalanceByCurrency(ctx)
        s.checkAccountLedgerConsistency(ctx, from)
        s.checkAccountLedgerConsistency(ctx, to)
        
        return nil
    })
    // Any error triggers automatic rollback
    return err
}
```

The `trm.Do()` wrapper ensures:
- All operations run in a single PostgreSQL transaction
- Any error (including invariant violations) triggers rollback
- No partial state is ever committed

### How do you prevent double-spending?

**Three-layer protection:**

1. **Pessimistic locking**: `SELECT ... FOR UPDATE` locks the account row
   ```go
   func (ar *AccountsRepository) GetForUpdate(ctx context.Context, accountID domain.AccountID) (*domain.Account, error) {
       const query = `SELECT ... FROM accounts WHERE id = $1 FOR UPDATE`
       // Row is locked until transaction commits/rollbacks
   }
   ```

2. **Domain validation**: Balance check before debit
   ```go
   func (a *Account) Debit(amount Money, tx *Transaction, timestamp time.Time) error {
       if a.balance.Amount().LessThan(amount.Amount()) {
           return NewInsufficientFundsError(...)
       }
       // Only proceeds if sufficient funds
   }
   ```

3. **Invariant verification**: Post-operation checks
   ```go
   // After transfer, verify account balance matches ledger sum
   err = s.checkAccountLedgerConsistency(ctx, from)
   ```

### How do you maintain consistency between ledger entries and account balances?

**Dual-write with immediate verification:**

1. Every balance change creates a ledger entry in the same transaction
2. Invariant check runs before commit:
   ```go
   func (s *Service) checkAccountLedgerConsistency(ctx context.Context, account *domain.Account) error {
       ledgerBalance, _ := s.ledger.GetAccountBalance(ctx, account.ID(), account.Balance().Currency())
       
       if !ledgerBalance.Amount().Equal(account.Balance().Amount()) {
           return domain.NewAccountBalanceMismatchError(...)
       }
       return nil
   }
   ```
3. Mismatch triggers rollback, preventing inconsistent state

### How would you handle decimal precision for different currencies?

**Current implementation:**

- Database: `DECIMAL(19, 2)` for amounts (up to 2 decimal places)
- Application: `shopspring/decimal` library for arbitrary precision
- Display: Frontend formats to 2 decimal places for USD/EUR


### What indexing strategy would you use for the ledger table?

**Current indexes:**

```sql
CREATE INDEX idx_ledger_transaction ON ledger(transaction);  -- Join with transactions
CREATE INDEX idx_ledger_account ON ledger(account);          -- Account balance queries
CREATE INDEX idx_ledger_timestamp ON ledger(timestamp);       -- Time-range queries
```

### How would you verify that balances are correctly synchronized?

**Runtime verification:**

```go
// Called after every transaction
func (s *Service) CheckLedgerBalanceByCurrency(ctx context.Context) error {
    totals, _ := s.ledger.GetTotalBalanceByCurrency(ctx)
    for currency, total := range totals {
        if !total.IsZero() {
            return domain.NewLedgerImbalanceError(currency, total.Amount())
        }
    }
    return nil
}
```

**On-demand reconciliation:**

```go
// GET /system/reconcile - full audit
func (s *Service) Reconcile(ctx context.Context) (*ReconciliationReport, error) {
    // 1. Check ledger balance by currency
    // 2. Find all accounts where balance != ledger sum
    // 3. Return detailed report
}
```

**SQL for mismatch detection:**

```sql
SELECT 
    a.id,
    a.balance AS account_balance,
    COALESCE(l.ledger_sum, 0) AS ledger_balance,
    a.currency
FROM accounts a
LEFT JOIN (
    SELECT account, SUM(amount) as ledger_sum
    FROM ledger
    GROUP BY account
) l ON a.id = l.account
WHERE a.balance != COALESCE(l.ledger_sum, 0)
```

### How would you scale this system for millions of users?

1. **Vertical scaling** (simplest approach):
   - Powerful PostgreSQL server with NVMe SSD
   - Sufficient for 100K-1M active users

2. **Read replicas + CQRS**:
   - All writes go to primary
   - Balance and history reads from replicas
   - Eventual consistency for read model (acceptable for display)

3. **Period Closing**:
   - Close periods (daily/monthly) with fixed opening balances
   - Archive old ledger entries to cold storage
   - Current operations work only with current period data
   ```sql
   -- Opening balance at period start
   CREATE TABLE period_balances (
       account_id UUID,
       period_start DATE,
       opening_balance DECIMAL(19,2),
       currency currency,
       PRIMARY KEY (account_id, period_start)
   );
   
   -- Balance = opening_balance + SUM(ledger for period)
   ```

**Connection pooling and optimizations:**

- PgBouncer for connection pooling
- Prepared statements for frequent queries
- Ledger partitioning by date for faster archival

---

## Technical Evaluation Checklist

### Must Have (70%)

- [x] **Functional double-entry ledger implementation** - Every transaction creates balanced ledger entries
- [x] **Ledger entries properly maintained as audit trail** - Full history preserved, never modified
- [x] **Account balances kept in sync with ledger** - Verified after every operation
- [x] **All transaction types working correctly** - Transfer and exchange fully implemented
- [x] **Proper handling of concurrent operations** - SELECT FOR UPDATE with pessimistic locking
- [x] **Prevention of invalid states** - Negative balances prevented, currency mismatches blocked
- [x] **Clean, organized code structure** - DDD backend, FSD frontend
- [x] **Authentication working** - JWT-based auth with login/register
- [x] **Functional user interface with working forms and data display** - React 19 with React Query

### Should Have (20%)

- [x] **Comprehensive error messages** - RFC 7807 Problem Details with contextual info
- [x] **Proper API error handling** - Frontend parses and displays backend errors
- [x] **Loading states in UI** - Spinners and disabled states during mutations
- [x] **Database migrations/seeders** - golang-migrate with cashbook seeding
- [x] **Environment configuration (.env)** - Docker and local development support
- [x] **Input validation on both frontend and backend** - Zod (frontend), go-playground/validator (backend)
- [ ] **Transaction confirmation before processing** - Not implemented (direct submission)

### Nice to Have (10%)

- [x] **Unit tests for critical financial logic** - Transfer, exchange, reconciliation tests with testcontainers
- [x] **Balance reconciliation/verification endpoint** - GET /system/reconcile with detailed report
- [x] **API documentation (Swagger/OpenAPI)** - Full swagger.yaml with examples
- [x] **Docker setup** - docker-compose for full stack deployment
- [x] **Polished UI with good user experience** - Tailwind CSS with clean design
- [ ] **Real-time balance updates (WebSockets)** - Not implemented
- [ ] **Transaction receipts/details modal** - Basic display only, no export
- [ ] **Audit log for all operations** - Backend ledger exists, no admin UI

---

## API Documentation

Full OpenAPI specification available at `backend/api/swagger.yaml`.

### Key Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /auth/register | Register new user |
| POST | /auth/login | Authenticate user |
| GET | /auth/me | Get current user info |
| GET | /accounts | List user's accounts |
| POST | /transactions/transfer | Transfer money |
| POST | /transactions/exchange | Exchange currency |
| GET | /transactions/exchange/calculate | Preview exchange rate |
| GET | /transactions | List transactions |
| GET | /system/reconcile | Run reconciliation check |

