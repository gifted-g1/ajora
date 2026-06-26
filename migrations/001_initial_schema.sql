
-- migrations/001_initial_schema.sql
-- Users Table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    date_of_birth DATE,
    kyc_status VARCHAR(20) DEFAULT PENDING,
    kyc_data JSONB,
    mfa_enabled BOOLEAN DEFAULT false,
    mfa_secret VARCHAR(255),
    role VARCHAR(20) DEFAULT USER,
    status VARCHAR(20) DEFAULT ACTIVE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Wallets Table
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    address VARCHAR(100) UNIQUE NOT NULL,
    public_key VARCHAR(100) NOT NULL,
    encrypted_private_key TEXT NOT NULL,
    key_version INT DEFAULT 1,
    chain VARCHAR(20) DEFAULT POLYGON,
    balance DECIMAL(20,8) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_transaction_at TIMESTAMP,
    metadata JSONB
);

-- Savings Pools Table
CREATE TABLE savings_pools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_address VARCHAR(100) UNIQUE,
    creator_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    pool_type VARCHAR(50) NOT NULL,
    total_slots INT NOT NULL,
    filled_slots INT DEFAULT 0,
    contribution_amount DECIMAL(20,2) NOT NULL,
    contribution_frequency VARCHAR(20) NOT NULL,
    total_rounds INT NOT NULL,
    current_round INT DEFAULT 0,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    interest_rate DECIMAL(5,2) DEFAULT 0,
    governance_token VARCHAR(100),
    status VARCHAR(30) DEFAULT DRAFT,
    smart_contract_version VARCHAR(10),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMP
);

-- Pool Members Table
CREATE TABLE pool_members (
    pool_id UUID NOT NULL REFERENCES savings_pools(id),
    user_id UUID NOT NULL REFERENCES users(id),
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    slot_number INT,
    join_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT ACTIVE,
    total_contributed DECIMAL(20,2) DEFAULT 0,
    total_payouts DECIMAL(20,2) DEFAULT 0,
    payout_round INT,
    penalty_points INT DEFAULT 0,
    metadata JSONB,
    PRIMARY KEY (pool_id, user_id)
);

-- Contributions Table
CREATE TABLE contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pool_id UUID NOT NULL REFERENCES savings_pools(id),
    member_id UUID NOT NULL REFERENCES users(id),
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    transaction_hash VARCHAR(100) UNIQUE,
    amount DECIMAL(20,2) NOT NULL,
    round_number INT NOT NULL,
    status VARCHAR(30) DEFAULT PENDING,
    block_number BIGINT,
    block_timestamp TIMESTAMP,
    gas_used BIGINT,
    gas_price DECIMAL(20,10),
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    confirmed_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Payouts Table
CREATE TABLE payouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pool_id UUID NOT NULL REFERENCES savings_pools(id),
    receiver_id UUID NOT NULL REFERENCES users(id),
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    transaction_hash VARCHAR(100) UNIQUE,
    amount DECIMAL(20,2) NOT NULL,
    round_number INT NOT NULL,
    total_contributions DECIMAL(20,2),
    interest_earned DECIMAL(20,2),
    status VARCHAR(30) DEFAULT PENDING,
    block_number BIGINT,
    block_timestamp TIMESTAMP,
    gas_used BIGINT,
    gas_price DECIMAL(20,10),
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Reputation Scores Table
CREATE TABLE reputation_scores (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    reliability_score DECIMAL(10,2) DEFAULT 0,
    completed_pools INT DEFAULT 0,
    on_time_contributions INT DEFAULT 0,
    missed_contributions INT DEFAULT 0,
    defaults INT DEFAULT 0,
    total_participated INT DEFAULT 0,
    positive_ratings INT DEFAULT 0,
    negative_ratings INT DEFAULT 0,
    trust_score DECIMAL(5,2) DEFAULT 0,
    last_calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Audit Logs Table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(100),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    request_id UUID,
    status VARCHAR(20) DEFAULT SUCCESS,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_kyc_status ON users(kyc_status);
CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_address ON wallets(address);
CREATE INDEX idx_pools_creator_id ON savings_pools(creator_id);
CREATE INDEX idx_pools_status ON savings_pools(status);
CREATE INDEX idx_pools_start_date ON savings_pools(start_date);
CREATE INDEX idx_pool_members_user_id ON pool_members(user_id);
CREATE INDEX idx_pool_members_pool_id ON pool_members(pool_id);
CREATE INDEX idx_contributions_pool_id ON contributions(pool_id);
CREATE INDEX idx_contributions_member_id ON contributions(member_id);
CREATE INDEX idx_contributions_status ON contributions(status);
CREATE INDEX idx_payouts_pool_id ON payouts(pool_id);
CREATE INDEX idx_payouts_receiver_id ON payouts(receiver_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

