-- Enum types
CREATE TYPE currency AS ENUM ('USD', 'EUR');
CREATE TYPE transaction_type AS ENUM ('transfer', 'exchange', 'deposit', 'withdrawal');

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Accounts table
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance DECIMAL(19, 2) NOT NULL DEFAULT 0,
    currency currency NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_accounts_currency ON accounts(currency);

-- Transactions table (base table for all transaction types)
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type transaction_type NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_account_id ON transactions(account_id);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_timestamp ON transactions(timestamp);

-- Transfer details table
CREATE TABLE transfer_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    recipient_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    amount DECIMAL(19, 2) NOT NULL,
    currency currency NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT transfer_positive_amount CHECK (amount > 0)
);

CREATE INDEX idx_transfer_details_transaction_id ON transfer_details(transaction_id);
CREATE INDEX idx_transfer_details_recipient_account_id ON transfer_details(recipient_account_id);

-- Exchange details table (similar to transfer but for currency exchange)
CREATE TABLE exchange_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    source_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    target_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    source_amount DECIMAL(19, 2) NOT NULL,
    source_currency currency NOT NULL,
    target_amount DECIMAL(19, 2) NOT NULL,
    target_currency currency NOT NULL,
    exchange_rate DECIMAL(19, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT exchange_positive_source_amount CHECK (source_amount > 0),
    CONSTRAINT exchange_positive_target_amount CHECK (target_amount > 0),
    CONSTRAINT exchange_positive_rate CHECK (exchange_rate > 0),
    CONSTRAINT exchange_different_currencies CHECK (source_currency != target_currency)
);

CREATE INDEX idx_exchange_details_transaction_id ON exchange_details(transaction_id);
CREATE INDEX idx_exchange_details_source_account_id ON exchange_details(source_account_id);
CREATE INDEX idx_exchange_details_target_account_id ON exchange_details(target_account_id);

-- Ledger table (double-entry bookkeeping)
-- Note: column names match the repository queries
CREATE TABLE ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    account UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    amount DECIMAL(19, 4) NOT NULL,
    currency currency NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_transaction ON ledger(transaction);
CREATE INDEX idx_ledger_account ON ledger(account);
CREATE INDEX idx_ledger_timestamp ON ledger(timestamp);
