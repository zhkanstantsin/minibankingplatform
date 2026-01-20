# AI Usage

This document describes how AI tools were used during development. My approach: **I designed the architecture and made technical decisions, AI helped write the code.**

For backend work, I focused on DDD principles, double-entry bookkeeping, and transaction safety. For frontend (outside my primary expertise), AI handled most of the implementation based on my requirements.

---

## AI-1

Purpose: Set up development infrastructure

Tool & Model: Cursor & Claude Sonnet 4.5

Prompt:
```
I'm starting a banking platform project. Need dev infrastructure:
- Docker Compose with PostgreSQL 17
- Credentials in .env, not hardcoded in compose
- golang-migrate for migrations
- Taskfile with all migration commands (up, down, create, version)

Make sure Taskfile reads DB credentials from .env so I don't duplicate them.
```

How the response was used: Used as-is for local development setup.

---

## AI-2

Purpose: Create domain enums with code generation

Tool & Model: Cursor & Claude Sonnet 4.5

Prompt:
```
Need proper enum types in the domain layer:

1. Currency enum (USD, EUR) - use go-enum for generation, need IsValid() and String() for validation/serialization

2. TransactionType enum - transfer, exchange, deposit, withdrawal. Deposit/withdrawal are needed so I can fund accounts on registration while keeping ledger consistent.

Enums must match the PostgreSQL enum types in migrations.
```

How the response was used: Set up go-enum with go:generate, created both enums, updated DB migrations.

---

## AI-3

Purpose: Design Exchange Rate as Value Object

Tool & Model: Cursor & Claude Sonnet 4.5

Prompt:
```
For currency exchange I need ExchangeRate as a proper Value Object following DDD:
- Immutable, stores from/to currencies and the rate itself
- Convert(amount) method, round to 2 decimals (it's money)
- Validate that rate > 0 and from != to currency
- Rate direction matters (USD->EUR might differ from EUR->USD in real systems)

Also need ExchangeRateProvider interface - service layer calls it to get the rate, passes the VO to domain. Implementation just returns fixed 1 USD = 0.92 EUR for now, but the interface allows swapping in a real provider later.

Goal: domain has no external dependencies, rate source is abstracted.
```

How the response was used: Created ExchangeRate VO with Convert(), ExchangeRateProvider interface, FixedExchangeRateProvider in infrastructure.

---

## AI-4

Purpose: Implement user registration with automatic account funding

Tool & Model: Cursor & Claude Opus 4.5

Prompt:
```
Need user registration and auth:

Register(email, password):
- Create user with bcrypt hash
- Auto-create USD account ($1000) and EUR account (€500)
- Important: initial balances must come FROM the system cashbook as deposits, not appear from nowhere. Double-entry means every credit needs a debit.
- Return JWT

Login just verifies credentials and returns JWT.

Put JWT logic in pkg/jwt, include user_id in claims.
```

How the response was used: Implemented Register/Login in service layer, created JWT package, account creation with proper deposit ledger entries from cashbook.

---

## AI-5

Purpose: Refactor service layer

Tool & Model: Cursor & Claude Opus 4.5

Prompt:
```
Service layer is getting big. Split it by business capability:
- users.go (register, login)
- accounts.go (get accounts, balances)
- transfer.go
- exchange.go
- transactions.go (history)
- reconcile.go (ledger validation)

Tests should match - transfer_test.go for transfer.go etc.

Don't over-engineer it, just logical grouping.
```

How the response was used: Split into 6 files with corresponding test files.

---

## AI-6

Purpose: Create OpenAPI spec and HTTP handlers

Tool & Model: Cursor & Claude Opus 4.5

Prompt:
```
All business logic is done in the service layer. Now need the REST API.

1. swagger.yaml with all endpoints (auth, accounts, transactions, transfer, exchange, reconcile)
   - JWT Bearer auth
   - Errors as RFC 7807 ProblemDetails - I already have domain errors, just need proper HTTP mapping
   - Add validation tags via x-oapi-codegen-extra-tags (using go-playground/validator)

2. Use oapi-codegen to generate server interface

3. Implement handlers that just translate between HTTP and service calls. Don't touch service layer.
```

How the response was used: Created swagger.yaml with RFC 7807 errors, configured oapi-codegen, implemented handlers.go and middleware.go.

---

## AI-7

Purpose: Write integration tests with testcontainers

Tool & Model: Cursor & Claude Sonnet 4.5

Prompt:
```
Need tests for Transfer using testcontainers-go (classical testing, real DB, no mocks).

Setup: spin up postgres container, apply migrations, create test users/accounts.

Cover:
- Successful transfer
- Insufficient funds
- Same account transfer (should fail)
- Currency mismatch
- Concurrent transfers (check for races)

SUT is the service layer. After each test verify ledger invariants: zero-sum by currency, account balance equals its ledger sum.
```

How the response was used: Created comprehensive test suite with testcontainers, all edge cases covered with ledger validation.

---

## AI-8

Purpose: Complete README documentation

Tool & Model: Cursor & Claude Opus 4.5

Prompt:
```
I wrote the technical Q&A section in README explaining double-entry design, atomicity, scaling, limitations etc. Need to complete the rest:
- Setup instructions (Docker and local)
- Architecture overview
- Assessment checklist

Keep my technical answers as-is, you can add code examples to illustrate them.
```

How the response was used: Added setup guide, architecture section, limitations. My technical explanations preserved.

---

## AI-9

Purpose: Implement frontend

Tool & Model: Claude Code & Claude Opus 4.5

Prompt:
```
I'm a backend dev, need to implement frontend for a full-stack assessment.

Stack: React 19, Vite, TypeScript, Tailwind

Architecture: Feature Sliced Design
- Layers: app -> pages -> widgets -> features -> entities -> shared
- Pages only compose widgets, no raw HTML in page components

Requirements:
- React Query for server state
- Zustand only for auth token
- Zod for validation
- Maximum component reuse

Need: UI kit, API client with JWT, auth flow, dashboard with wallets/transfer/exchange forms, transaction history.

This is outside my expertise so I need a complete working implementation.
```

How the response was used: Generated complete React app with FSD structure and all features.

---

## AI-10

Purpose: Fix frontend UX issues

Tool & Model: Claude Code & Claude Opus 4.5

Prompt:
```
Found some UX problems testing the frontend:

1. User can't see their account ID anywhere - how do they tell someone where to send money? Add it to wallet cards with copy button.

2. All transactions show as red/negative. Incoming transfers should be green/positive - determine direction by checking if transaction account matches user's account.

3. Exchange should show as two lines (debit from one currency, credit to another) since it affects two accounts.

4. Add reconcile status indicator in header - green dot if ledger is consistent, red if not. Show details on hover, don't need a whole page for this.
```

How the response was used: Updated TransactionRow, added account ID display, created ReconcileIndicator component.
