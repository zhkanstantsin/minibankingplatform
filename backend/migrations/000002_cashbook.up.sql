-- Cashbook is a system account used for double-entry bookkeeping.
-- All deposits come FROM cashbook, all withdrawals go TO cashbook.
-- This ensures the sum of all ledger entries is always zero.

-- System user for cashbook accounts
INSERT INTO users (id, email, password_hash)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'system@cashbook.internal',
    'SYSTEM_ACCOUNT_NO_LOGIN'
);

-- Cashbook account for USD
INSERT INTO accounts (id, user_id, balance, currency)
VALUES (
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000001',
    0,
    'USD'
);

-- Cashbook account for EUR
INSERT INTO accounts (id, user_id, balance, currency)
VALUES (
    '00000000-0000-0000-0000-000000000011',
    '00000000-0000-0000-0000-000000000001',
    0,
    'EUR'
);
