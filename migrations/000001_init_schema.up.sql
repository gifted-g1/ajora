-- Ajora Core Database Schema (PostgreSQL 16)
-- All financial amounts in Kobo (Minor Unit, 1 Naira = 100 Kobo)

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1. Users Table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    transaction_pin_hash VARCHAR(255),
    role VARCHAR(30) DEFAULT 'MEMBER' NOT NULL CHECK (role IN ('MEMBER', 'ORGANIZER', 'ADMIN', 'SUPER_ADMIN')),
    is_phone_verified BOOLEAN DEFAULT FALSE NOT NULL,
    status VARCHAR(30) DEFAULT 'ACTIVE' NOT NULL CHECK (status IN ('ACTIVE', 'SUSPENDED', 'DEACTIVATED')),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone_number);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- 2. User Profiles Table
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    full_name VARCHAR(150) NOT NULL,
    avatar_url TEXT,
    occupation VARCHAR(100),
    market_location VARCHAR(150),
    state VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- 3. Wallets Table
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    available_balance_kobo BIGINT DEFAULT 0 NOT NULL CHECK (available_balance_kobo >= 0),
    locked_balance_kobo BIGINT DEFAULT 0 NOT NULL CHECK (locked_balance_kobo >= 0),
    currency VARCHAR(5) DEFAULT 'NGN' NOT NULL,
    virtual_account_bank VARCHAR(100),
    virtual_account_number VARCHAR(20),
    virtual_account_name VARCHAR(150),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_wallets_user ON wallets(user_id);

-- 4. Wallet Transactions Table (Double-Entry / Balance Audit)
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
    reference VARCHAR(100) UNIQUE NOT NULL,
    transaction_type VARCHAR(50) NOT NULL CHECK (transaction_type IN ('DEPOSIT', 'WITHDRAWAL', 'AJO_CONTRIBUTION', 'AJO_PAYOUT', 'PENALTY', 'REFUND')),
    amount_kobo BIGINT NOT NULL,
    balance_before_kobo BIGINT NOT NULL,
    balance_after_kobo BIGINT NOT NULL,
    status VARCHAR(30) DEFAULT 'CONFIRMED' NOT NULL CHECK (status IN ('PENDING', 'PROCESSING', 'CONFIRMED', 'FAILED', 'REVERSED')),
    description TEXT,
    idempotency_key VARCHAR(100) UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_wallet_tx_wallet ON wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_tx_ref ON wallet_transactions(reference);
CREATE INDEX IF NOT EXISTS idx_wallet_tx_idempotency ON wallet_transactions(idempotency_key);

-- 5. Ajo Circles Table
CREATE TABLE IF NOT EXISTS ajos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    name VARCHAR(150) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    invite_code VARCHAR(10) UNIQUE NOT NULL,
    is_private BOOLEAN DEFAULT FALSE NOT NULL,
    contribution_amount_kobo BIGINT NOT NULL CHECK (contribution_amount_kobo > 0),
    frequency VARCHAR(30) NOT NULL CHECK (frequency IN ('DAILY', 'WEEKLY', 'BIWEEKLY', 'MONTHLY')),
    total_rounds INT NOT NULL CHECK (total_rounds > 1),
    max_members INT NOT NULL CHECK (max_members >= 2),
    min_members INT NOT NULL CHECK (min_members >= 2),
    payout_strategy VARCHAR(50) NOT NULL CHECK (payout_strategy IN ('ORGANIZER_DEFINED', 'FIRST_COME', 'RANDOMIZED')),
    grace_period_hours INT DEFAULT 24 NOT NULL CHECK (grace_period_hours >= 0),
    late_fee_kobo BIGINT DEFAULT 0 NOT NULL CHECK (late_fee_kobo >= 0),
    status VARCHAR(30) DEFAULT 'OPEN' NOT NULL CHECK (status IN ('DRAFT', 'OPEN', 'ACTIVE', 'COMPLETED', 'CANCELLED')),
    smart_contract_circle_id VARCHAR(100),
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ajos_organizer ON ajos(organizer_id);
CREATE INDEX IF NOT EXISTS idx_ajos_invite_code ON ajos(invite_code);
CREATE INDEX IF NOT EXISTS idx_ajos_status ON ajos(status);

-- 6. Ajo Members Table
CREATE TABLE IF NOT EXISTS ajo_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ajo_id UUID NOT NULL REFERENCES ajos(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    payout_position INT NOT NULL CHECK (payout_position > 0),
    status VARCHAR(30) DEFAULT 'ACTIVE' NOT NULL CHECK (status IN ('ACTIVE', 'SUSPENDED', 'COMPLETED', 'LEFT')),
    joined_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    UNIQUE (ajo_id, user_id),
    UNIQUE (ajo_id, payout_position)
);

CREATE INDEX IF NOT EXISTS idx_ajo_members_ajo ON ajo_members(ajo_id);
CREATE INDEX IF NOT EXISTS idx_ajo_members_user ON ajo_members(user_id);

-- 7. Ajo Rounds Table
CREATE TABLE IF NOT EXISTS ajo_rounds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ajo_id UUID NOT NULL REFERENCES ajos(id) ON DELETE CASCADE,
    round_number INT NOT NULL CHECK (round_number > 0),
    beneficiary_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    expected_amount_kobo BIGINT NOT NULL CHECK (expected_amount_kobo > 0),
    collected_amount_kobo BIGINT DEFAULT 0 NOT NULL CHECK (collected_amount_kobo >= 0),
    start_date TIMESTAMPTZ NOT NULL,
    due_date TIMESTAMPTZ NOT NULL,
    grace_period_end TIMESTAMPTZ NOT NULL,
    status VARCHAR(30) DEFAULT 'UPCOMING' NOT NULL CHECK (status IN ('UPCOMING', 'CONTRIBUTING', 'ROUND_COMPLETE', 'PAYOUT_READY', 'PAYOUT_COMPLETED')),
    completed_at TIMESTAMPTZ,
    UNIQUE (ajo_id, round_number)
);

CREATE INDEX IF NOT EXISTS idx_ajo_rounds_ajo ON ajo_rounds(ajo_id);
CREATE INDEX IF NOT EXISTS idx_ajo_rounds_beneficiary ON ajo_rounds(beneficiary_user_id);

-- 8. Contributions Table
CREATE TABLE IF NOT EXISTS contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ajo_id UUID NOT NULL REFERENCES ajos(id) ON DELETE CASCADE,
    round_id UUID NOT NULL REFERENCES ajo_rounds(id) ON DELETE CASCADE,
    member_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount_kobo BIGINT NOT NULL CHECK (amount_kobo > 0),
    fee_kobo BIGINT DEFAULT 0 NOT NULL,
    reference VARCHAR(100) UNIQUE NOT NULL,
    idempotency_key VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(30) DEFAULT 'PENDING' NOT NULL CHECK (status IN ('PENDING', 'PROCESSING', 'CONFIRMED', 'LATE', 'FAILED')),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    UNIQUE (round_id, member_user_id)
);

CREATE INDEX IF NOT EXISTS idx_contributions_ajo ON contributions(ajo_id);
CREATE INDEX IF NOT EXISTS idx_contributions_round ON contributions(round_id);
CREATE INDEX IF NOT EXISTS idx_contributions_member ON contributions(member_user_id);

-- 9. Payouts Table
CREATE TABLE IF NOT EXISTS payouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ajo_id UUID NOT NULL REFERENCES ajos(id) ON DELETE CASCADE,
    round_id UUID UNIQUE NOT NULL REFERENCES ajo_rounds(id) ON DELETE CASCADE,
    beneficiary_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    gross_amount_kobo BIGINT NOT NULL CHECK (gross_amount_kobo > 0),
    deductions_kobo BIGINT DEFAULT 0 NOT NULL,
    net_amount_kobo BIGINT NOT NULL CHECK (net_amount_kobo > 0),
    status VARCHAR(30) DEFAULT 'SCHEDULED' NOT NULL CHECK (status IN ('SCHEDULED', 'READY', 'PROCESSING', 'COMPLETED', 'FAILED')),
    claimed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_payouts_beneficiary ON payouts(beneficiary_user_id);
CREATE INDEX IF NOT EXISTS idx_payouts_ajo ON payouts(ajo_id);

-- 10. Blockchain Audit Transactions Table
CREATE TABLE IF NOT EXISTS blockchain_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ajo_id UUID REFERENCES ajos(id) ON DELETE CASCADE,
    round_id UUID REFERENCES ajo_rounds(id) ON DELETE SET NULL,
    tx_hash VARCHAR(100) UNIQUE NOT NULL,
    network VARCHAR(50) NOT NULL,
    block_number BIGINT,
    contract_address VARCHAR(100) NOT NULL,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('CIRCLE_CREATED', 'CONTRIBUTION_NOTARIZED', 'PAYOUT_RELEASED')),
    payload_hash VARCHAR(100) NOT NULL,
    status VARCHAR(30) DEFAULT 'CONFIRMED' NOT NULL CHECK (status IN ('PENDING', 'CONFIRMED', 'FAILED')),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_blockchain_tx_ajo ON blockchain_transactions(ajo_id);

-- 11. Reputation Scores Table
CREATE TABLE IF NOT EXISTS reputation_scores (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    score_percentage INT DEFAULT 100 NOT NULL CHECK (score_percentage BETWEEN 0 AND 100),
    total_contributions_made INT DEFAULT 0 NOT NULL,
    on_time_contributions INT DEFAULT 0 NOT NULL,
    late_contributions INT DEFAULT 0 NOT NULL,
    completed_circles INT DEFAULT 0 NOT NULL,
    tier VARCHAR(30) DEFAULT 'BRONZE' NOT NULL CHECK (tier IN ('BRONZE', 'SILVER', 'GOLD', 'PLATINUM')),
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- 12. Audit Logs Table (Append-Only)
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(100) NOT NULL,
    ip_address VARCHAR(50),
    device_info TEXT,
    old_values JSONB,
    new_values JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
