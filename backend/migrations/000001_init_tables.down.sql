-- Drop tables in reverse order due to foreign key constraints
DROP TABLE IF EXISTS ledger;
DROP TABLE IF EXISTS exchange_details;
DROP TABLE IF EXISTS transfer_details;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS users;

-- Drop enum types
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS currency;
