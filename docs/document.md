# Ajora Platform - Complete Documentation

```markdown
# Ajora Platform Documentation

## 📋 Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [System Design](#system-design)
4. [Services](#services)
5. [Database Schema](#database-schema)
6. [Smart Contracts](#smart-contracts)
7. [API Reference](#api-reference)
8. [Security](#security)
9. [Deployment](#deployment)
10. [Monitoring & Observability](#monitoring--observability)
11. [Development Guide](#development-guide)
12. [Performance Optimization](#performance-optimization)
13. [Troubleshooting](#troubleshooting)
14. [Contributing](#contributing)

---

## Overview

### What is Ajora?

Ajora is a production-grade, blockchain-integrated fintech platform that enables decentralized rotating savings and credit associations (ROSCAs), commonly known as "Ajo" in West Africa. The platform combines traditional savings group mechanics with modern blockchain technology to create a transparent, secure, and accessible financial ecosystem.

### Key Features

- **Decentralized Savings Pools**: Create and join savings groups with smart contract guarantees
- **Automated Payouts**: Fair and transparent distribution of funds using blockchain
- **Reputation System**: Trust scoring based on participation history
- **Multi-Factor Authentication**: Enhanced security with TOTP
- **Real-time Notifications**: Email, SMS, and push notifications
- **KYC Compliance**: Identity verification for regulatory compliance
- **Hybrid Architecture**: Go for business logic, Rust for cryptographic operations
- **Zero-Trust Security**: End-to-end encryption and strict access controls

### Business Value

- **Financial Inclusion**: Access to savings and credit for underserved populations
- **Transparency**: All transactions recorded on blockchain
- **Trust**: Reputation system incentivizes good behavior
- **Efficiency**: Automated processes reduce operational costs
- **Scalability**: Support for 100,000+ users and 10,000+ active pools

---

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Client Applications                     │
│         (Web, Mobile, API Clients)                         │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway (Go)                         │
│         - Request Routing                                   │
│         - Rate Limiting (100 req/sec)                      │
│         - API Composition                                   │
│         - WebSocket Management                              │
└─────────┬───────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────┐
│                   Service Mesh (Istio)                      │
│         - Service Discovery                                 │
│         - Load Balancing                                    │
│         - Circuit Breaking                                  │
│         - Observability                                     │
└─────────┬───────────────────────────────────────────────────┘
          │
    ┌─────┴─────┬──────────┬──────────┬──────────┐
    ▼           ▼          ▼          ▼          ▼
┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
│  Auth   │ │  User   │ │  Pool   │ │Contrib. │ │  Rep.   │
│ Service │ │ Service │ │ Service │ │ Service │ │ Service │
│  (Go)   │ │  (Go)   │ │  (Go)   │ │  (Go)   │ │  (Go)   │
└────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘
     │           │           │           │           │
     └───────────┴───────────┴───────────┴───────────┘
                           │
                    ┌──────▼──────┐
                    │ Blockchain  │
                    │ Orchestr.   │
                    │   (Go)      │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │   Rust      │
                    │  Services   │
                    │  - Signer   │
                    │  - Validator│
                    │  - Fraud    │
                    │  - Audit    │
                    └─────────────┘
```

### Technology Stack

#### Backend Services
| Component | Technology | Purpose |
|-----------|-----------|---------|
| Primary Services | Go 1.21+ | Business logic, API orchestration |
| Critical Services | Rust 1.70+ | Crypto, validation, security |
| API Gateway | Go + Gorilla Mux | Request routing, rate limiting |
| Authentication | JWT + OAuth2 | User authentication and authorization |

#### Infrastructure
| Component | Technology | Purpose |
|-----------|-----------|---------|
| Container Orchestration | Kubernetes (EKS) | Service deployment and scaling |
| Infrastructure as Code | Terraform | AWS resource provisioning |
| Database | PostgreSQL 15 | Primary data store |
| Cache | Redis 7 | Caching, rate limiting, sessions |
| Message Queue | Apache Kafka | Event streaming and async processing |
| Service Mesh | Istio | Traffic management, observability |

#### Blockchain
| Component | Technology | Purpose |
|-----------|-----------|---------|
| Network | Polygon (EVM compatible) | Smart contract deployment |
| Smart Contracts | Solidity 0.8.19 | Savings pool logic |
| Key Management | AWS KMS | Cryptographic key storage |
| Wallet Integration | Web3.js | Blockchain interaction |

#### Observability
| Component | Technology | Purpose |
|-----------|-----------|---------|
| Metrics | Prometheus | Performance monitoring |
| Visualization | Grafana | Dashboards and alerts |
| Tracing | OpenTelemetry | Distributed tracing |
| Logging | ELK/Loki | Centralized log management |

---

## System Design

### Domain Model

```
┌─────────────────────────────────────────────────────────────┐
│                         User                                │
├─────────────────────────────────────────────────────────────┤
│ - id: UUID                                                  │
│ - email: String                                             │
│ - phone: String                                             │
│ - password_hash: String                                     │
│ - first_name: String                                        │
│ - last_name: String                                         │
│ - kyc_status: KYCStatus                                     │
│ - mfa_enabled: Boolean                                      │
│ - role: UserRole                                            │
│ - reputation_score: Decimal                                 │
└────────────┬────────────────────────────────────────────────┘
             │ 1
             │
             │ has
             │
             ▼
┌─────────────────────────────────────────────────────────────┐
│                         Wallet                              │
├─────────────────────────────────────────────────────────────┤
│ - id: UUID                                                  │
│ - user_id: UUID                                             │
│ - address: String                                           │
│ - public_key: String                                        │
│ - encrypted_private_key: String                             │
│ - balance: Decimal                                          │
│ - chain: ChainType                                          │
└────────────┬────────────────────────────────────────────────┘
             │
             │
             ▼
┌─────────────────────────────────────────────────────────────┐
│                      Savings Pool                           │
├─────────────────────────────────────────────────────────────┤
│ - id: UUID                                                  │
│ - contract_address: String                                  │
│ - creator_id: UUID                                          │
│ - name: String                                              │
│ - pool_type: PoolType                                       │
│ - total_slots: Integer                                      │
│ - contribution_amount: Decimal                              │
│ - contribution_frequency: Frequency                         │
│ - total_rounds: Integer                                     │
│ - current_round: Integer                                    │
│ - start_date: DateTime                                      │
│ - end_date: DateTime                                        │
│ - interest_rate: Decimal                                    │
│ - status: PoolStatus                                        │
└────────────┬────────────────────────────────────────────────┘
             │ 1
             │
             │ contains
             │
             ▼
┌─────────────────────────────────────────────────────────────┐
│                      Pool Member                            │
├─────────────────────────────────────────────────────────────┤
│ - pool_id: UUID                                             │
│ - user_id: UUID                                             │
│ - wallet_id: UUID                                           │
│ - slot_number: Integer                                      │
│ - join_date: DateTime                                       │
│ - status: MemberStatus                                      │
│ - total_contributed: Decimal                                │
│ - total_payouts: Decimal                                    │
│ - penalty_points: Integer                                   │
└────────────┬────────────────────────────────────────────────┘
             │
             │ makes
             │
             ▼
┌─────────────────────────────────────────────────────────────┐
│                     Contribution                            │
├─────────────────────────────────────────────────────────────┤
│ - id: UUID                                                  │
│ - pool_id: UUID                                             │
│ - member_id: UUID                                           │
│ - wallet_id: UUID                                           │
│ - transaction_hash: String                                  │
│ - amount: Decimal                                           │
│ - round_number: Integer                                     │
│ - status: ContributionStatus                                │
│ - block_number: BigInteger                                  │
│ - gas_used: BigInteger                                      │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ 1. Request
       ▼
┌─────────────┐    2. Validate     ┌─────────────┐
│API Gateway  │──────────────────▶ │ Auth Service│
└──────┬──────┘                    └──────┬──────┘
       │ 3. Route                        │
       ▼                                   │
┌─────────────┐    4. Process            │
│Pool Service │──────────────────┐       │
└──────┬──────┘                  │       │
       │                        ▼       │
       │ 5. Create Pool    ┌─────────────┐
       └─────────────────▶│Blockchain   │
                           │Service      │
                           └──────┬──────┘
                                  │ 6. Sign TX
                                  ▼
                           ┌─────────────┐
                           │Wallet Signer│
                           │  (Rust)     │
                           └──────┬──────┘
                                  │ 7. Deploy Contract
                                  ▼
                           ┌─────────────┐
                           │  Polygon    │
                           │ Blockchain  │
                           └─────────────┘
```

### Event-Driven Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Event Flow                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  User Registration ──▶ user.registered ──▶ Welcome Email   │
│                                                             │
│  Pool Creation ──────▶ pool.created ──────▶ Notify Members │
│                                                             │
│  Contribution ────────▶ contribution.submitted             │
│                          │                                   │
│                          ▼                                   │
│                   Validate Transaction                      │
│                          │                                   │
│                          ▼                                   │
│                   Update Blockchain                         │
│                          │                                   │
│                          ▼                                   │
│                   Send Notification                         │
│                                                             │
│  Payout ─────────────▶ payout.processed                    │
│                          │                                   │
│                          ▼                                   │
│                   Distribute Funds                          │
│                          │                                   │
│                          ▼                                   │
│                   Notify Recipients                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Services

### Go Services (8 Services)

#### 1. API Gateway
**Purpose**: Entry point for all client requests

**Responsibilities**:
- Request routing and load balancing
- Rate limiting (100 req/sec per user)
- API composition and aggregation
- WebSocket connections for real-time updates
- CORS and security headers

**Endpoints**:
```
GET  /health                    - Health check
GET  /api/v1/docs              - OpenAPI documentation
WS   /api/v1/ws                - WebSocket connection
```

**Configuration**:
```yaml
port: 8080
rate_limit: 100
timeout: 30s
max_body_size: 10MB
cors_origins:
  - https://app.ajora.com
  - https://api.ajora.com
```

#### 2. Auth Service
**Purpose**: Authentication and authorization

**Responsibilities**:
- JWT token generation and validation
- OAuth2 integration (Google, GitHub)
- MFA (TOTP) setup and verification
- Password management (reset, change)
- Session management
- Role-Based Access Control (RBAC)

**Endpoints**:
```
POST /api/v1/auth/register      - Register new user
POST /api/v1/auth/login         - User login
POST /api/v1/auth/refresh       - Refresh access token
POST /api/v1/auth/logout        - User logout
POST /api/v1/auth/verify-email  - Verify email address
POST /api/v1/auth/reset-password - Request password reset
PUT  /api/v1/auth/change-password - Change password
POST /api/v1/auth/mfa/enable    - Enable MFA
POST /api/v1/auth/mfa/verify    - Verify MFA code
POST /api/v1/auth/mfa/disable   - Disable MFA
GET  /api/v1/auth/sessions      - List active sessions
DELETE /api/v1/auth/sessions/{id} - Revoke session
```

**JWT Claims**:
```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "role": "USER",
  "permissions": ["read", "write"],
  "exp": 1234567890,
  "iat": 1234567890,
  "iss": "ajora-auth"
}
```

#### 3. User Service
**Purpose**: User management and profiles

**Responsibilities**:
- User profile CRUD operations
- KYC verification
- User preferences
- Wallet management
- Account deactivation

**Endpoints**:
```
GET    /api/v1/users/{id}        - Get user profile
PUT    /api/v1/users/{id}        - Update user profile
DELETE /api/v1/users/{id}        - Deactivate account
POST   /api/v1/users/{id}/kyc    - Submit KYC documents
GET    /api/v1/users/{id}/kyc    - Get KYC status
POST   /api/v1/users/{id}/wallet - Create wallet
GET    /api/v1/users/{id}/wallet - Get wallet info
```

**KYC Flow**:
```
1. User submits documents
2. Document validation
3. Manual or automated review
4. Status update (PENDING → VERIFIED/REJECTED)
5. Notification of result
```

#### 4. Pool Management Service
**Purpose**: Savings pool lifecycle management

**Responsibilities**:
- Pool creation and configuration
- Member management (join/leave)
- Pool lifecycle (draft → active → completed)
- Pool statistics
- Round management

**Endpoints**:
```
POST   /api/v1/pools              - Create pool
GET    /api/v1/pools              - List pools
GET    /api/v1/pools/{id}         - Get pool details
PUT    /api/v1/pools/{id}         - Update pool
DELETE /api/v1/pools/{id}         - Cancel pool
POST   /api/v1/pools/{id}/join    - Join pool
POST   /api/v1/pools/{id}/leave   - Leave pool
GET    /api/v1/pools/{id}/members - Get pool members
GET    /api/v1/pools/{id}/rounds  - Get round history
```

**Pool Types**:
- **FIXED**: Fixed contribution amount and duration
- **FLEXIBLE**: Variable contributions and duration
- **INSTANT**: Immediate payout upon completion

**Pool Status Flow**:
```
DRAFT → (Configure) → ACTIVE → (Complete) → COMPLETED
                    ↓
                (Cancel)
                    ↓
              CANCELLED
```

#### 5. Contribution Service
**Purpose**: Contribution processing

**Responsibilities**:
- Process contributions
- Validate contribution amounts
- Track contribution history
- Handle retries and failures
- Reconciliation with blockchain

**Endpoints**:
```
POST   /api/v1/contributions          - Submit contribution
GET    /api/v1/contributions/{id}     - Get contribution
GET    /api/v1/users/{id}/contributions - Get user contributions
GET    /api/v1/pools/{id}/contributions - Get pool contributions
POST   /api/v1/contributions/retry    - Retry failed contribution
```

**Contribution Flow**:
```
1. User submits contribution
2. Validate pool and user
3. Check wallet balance
4. Create blockchain transaction
5. Wait for confirmation
6. Update database
7. Send notification
```

#### 6. Notification Service
**Purpose**: Multi-channel notifications

**Responsibilities**:
- Email notifications
- SMS notifications (Twilio)
- Push notifications (Firebase)
- Notification templates
- Delivery tracking

**Endpoints**:
```
POST /api/v1/notifications/email     - Send email
POST /api/v1/notifications/sms       - Send SMS
POST /api/v1/notifications/push      - Send push notification
GET  /api/v1/notifications           - List notifications
GET  /api/v1/notifications/{id}      - Get notification status
```

**Notification Types**:
| Event | Email | SMS | Push |
|-------|-------|-----|------|
| Registration | ✅ | ✅ | ❌ |
| Pool Creation | ✅ | ✅ | ✅ |
| Contribution | ✅ | ✅ | ✅ |
| Payout | ✅ | ✅ | ✅ |
| Member Join | ✅ | ✅ | ✅ |
| Member Leave | ✅ | ✅ | ❌ |
| Reminder | ✅ | ✅ | ✅ |
| KYC Update | ✅ | ❌ | ❌ |

#### 7. Blockchain Orchestration Service
**Purpose**: Smart contract interaction

**Responsibilities**:
- Contract deployment
- Transaction submission
- Event listening
- Gas optimization
- Error handling

**Endpoints**:
```
POST   /api/v1/blockchain/deploy     - Deploy contract
POST   /api/v1/blockchain/transact   - Send transaction
GET    /api/v1/blockchain/tx/{hash}  - Get transaction status
GET    /api/v1/blockchain/events     - Get contract events
GET    /api/v1/blockchain/gas        - Get gas price
```

**Smart Contract Methods**:
```solidity
function joinPool() external payable
function contribute() external payable
function processPayout() external
function getMemberCount() external view returns (uint256)
function getMembers() external view returns (address[] memory)
```

#### 8. Reputation Service
**Purpose**: Trust scoring system

**Responsibilities**:
- Calculate reliability scores
- Track participation history
- Handle penalties
- Generate trust metrics
- Provide fraud indicators

**Endpoints**:
```
GET /api/v1/reputation/users/{id}    - Get user reputation
GET /api/v1/reputation/top          - Get top users
POST /api/v1/reputation/update      - Manual update
GET /api/v1/reputation/metrics      - Get system metrics
```

**Reputation Algorithm**:
```python
reliability_score = (
    completed_pools * 10 +
    on_time_contributions * 5 -
    missed_contributions * 10 -
    defaults * 50
)

trust_score = min(100, max(0, reliability_score / 10))

# Penalty multipliers
- 1 missed contribution: 10 points
- 1 default: 50 points
- 1 complaint: 25 points
- Early withdrawal: 20 points
```

### Rust Services (4 Services)

#### 1. Wallet Signer
**Purpose**: Cryptographic signing operations

**Responsibilities**:
- Sign transactions using AWS KMS
- Generate key pairs
- Verify signatures
- Secure key management
- Audit trail

**Endpoints**:
```
POST /api/v1/signer/sign           - Sign transaction
POST /api/v1/signer/verify         - Verify signature
POST /api/v1/signer/generate-keypair - Generate key pair
```

**Security Features**:
- Hardware Security Module (HSM) integration
- AWS KMS for key storage
- No private key exposure
- Audit logging

#### 2. Transaction Validator
**Purpose**: Transaction validation engine

**Responsibilities**:
- Validate transaction structure
- Check balance sufficiency
- Verify signatures
- Replay attack protection
- Double-spend prevention

**Validation Rules**:
```rust
// Validate transaction
fn validate_transaction(tx: &Transaction) -> Result<(), ValidationError> {
    // 1. Check transaction structure
    if tx.amount <= 0 {
        return Err(ValidationError::InvalidAmount);
    }
    
    // 2. Verify signature
    if !verify_signature(&tx.signature, &tx.public_key, &tx.hash) {
        return Err(ValidationError::InvalidSignature);
    }
    
    // 3. Check for replay attack
    if is_replay_attack(&tx.nonce, &tx.from) {
        return Err(ValidationError::ReplayAttack);
    }
    
    // 4. Check balance
    if !has_sufficient_balance(&tx.from, tx.amount) {
        return Err(ValidationError::InsufficientBalance);
    }
    
    Ok(())
}
```

#### 3. Fraud Detection Engine
**Purpose**: Anomaly detection and fraud prevention

**Responsibilities**:
- Rule-based fraud detection
- Machine learning anomaly detection
- Real-time analysis
- Alert generation
- Risk scoring

**Fraud Rules**:
```yaml
rules:
  - name: "Rapid Contributions"
    condition: "contributions_per_minute > 10"
    action: "block"
    
  - name: "Large Withdrawal"
    condition: "withdrawal_amount > 10000"
    action: "review"
    
  - name: "Suspicious Timing"
    condition: "transaction_time BETWEEN 00:00 AND 04:00"
    action: "alert"
    
  - name: "Multiple Accounts"
    condition: "same_ip_address AND different_users"
    action: "investigate"
```

**Anomaly Detection**:
- Statistical methods (Z-score, IQR)
- Machine learning (Isolation Forest)
- Behavioral analysis
- Graph analysis for network fraud

#### 4. Cryptographic Audit Service
**Purpose**: Audit logging and verification

**Responsibilities**:
- Immutable audit logging
- Cryptographic verification
- Compliance reporting
- Chain of custody

**Audit Log Structure**:
```rust
struct AuditLog {
    id: UUID,
    timestamp: DateTime,
    user_id: UUID,
    action: String,
    resource_type: String,
    resource_id: String,
    details: JSON,
    signature: String,  // Cryptographic signature
    hash: String,       // Hash of previous log
    status: String,     // SUCCESS/FAILURE
}
```

---

## Database Schema

### Complete ER Diagram

```sql
-- =============================================
-- 1. USERS TABLE
-- =============================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    date_of_birth DATE,
    kyc_status VARCHAR(20) DEFAULT 'PENDING',
    kyc_data JSONB,
    mfa_enabled BOOLEAN DEFAULT false,
    mfa_secret VARCHAR(255),
    role VARCHAR(20) DEFAULT 'USER',
    status VARCHAR(20) DEFAULT 'ACTIVE',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Indexes
    INDEX idx_users_email (email),
    INDEX idx_users_phone (phone),
    INDEX idx_users_kyc_status (kyc_status),
    INDEX idx_users_status (status)
);

-- =============================================
-- 2. WALLETS TABLE
-- =============================================
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    address VARCHAR(100) UNIQUE NOT NULL,
    public_key VARCHAR(100) NOT NULL,
    encrypted_private_key TEXT NOT NULL,
    key_version INT DEFAULT 1,
    chain VARCHAR(20) DEFAULT 'POLYGON',
    balance DECIMAL(20,8) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_transaction_at TIMESTAMP,
    metadata JSONB,
    
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_wallets_user_id (user_id),
    INDEX idx_wallets_address (address),
    INDEX idx_wallets_active (is_active)
);

-- =============================================
-- 3. SAVINGS POOLS TABLE
-- =============================================
CREATE TABLE savings_pools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_address VARCHAR(100) UNIQUE,
    creator_id UUID NOT NULL,
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
    status VARCHAR(30) DEFAULT 'DRAFT',
    smart_contract_version VARCHAR(10),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMP,
    
    FOREIGN KEY (creator_id) REFERENCES users(id),
    INDEX idx_pools_creator_id (creator_id),
    INDEX idx_pools_status (status),
    INDEX idx_pools_start_date (start_date),
    INDEX idx_pools_end_date (end_date)
);

-- =============================================
-- 4. POOL MEMBERS TABLE
-- =============================================
CREATE TABLE pool_members (
    pool_id UUID NOT NULL,
    user_id UUID NOT NULL,
    wallet_id UUID NOT NULL,
    slot_number INT,
    join_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    total_contributed DECIMAL(20,2) DEFAULT 0,
    total_payouts DECIMAL(20,2) DEFAULT 0,
    payout_round INT,
    penalty_points INT DEFAULT 0,
    metadata JSONB,
    
    PRIMARY KEY (pool_id, user_id),
    FOREIGN KEY (pool_id) REFERENCES savings_pools(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (wallet_id) REFERENCES wallets(id),
    INDEX idx_pool_members_user_id (user_id),
    INDEX idx_pool_members_status (status)
);

-- =============================================
-- 5. CONTRIBUTIONS TABLE
-- =============================================
CREATE TABLE contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pool_id UUID NOT NULL,
    member_id UUID NOT NULL,
    wallet_id UUID NOT NULL,
    transaction_hash VARCHAR(100) UNIQUE,
    amount DECIMAL(20,2) NOT NULL,
    round_number INT NOT NULL,
    status VARCHAR(30) DEFAULT 'PENDING',
    block_number BIGINT,
    block_timestamp TIMESTAMP,
    gas_used BIGINT,
    gas_price DECIMAL(20,10),
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    confirmed_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (pool_id) REFERENCES savings_pools(id),
    FOREIGN KEY (member_id) REFERENCES users(id),
    FOREIGN KEY (wallet_id) REFERENCES wallets(id),
    INDEX idx_contributions_pool_id (pool_id),
    INDEX idx_contributions_member_id (member_id),
    INDEX idx_contributions_status (status),
    INDEX idx_contributions_created_at (created_at DESC)
);

-- =============================================
-- 6. PAYOUTS TABLE
-- =============================================
CREATE TABLE payouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pool_id UUID NOT NULL,
    receiver_id UUID NOT NULL,
    wallet_id UUID NOT NULL,
    transaction_hash VARCHAR(100) UNIQUE,
    amount DECIMAL(20,2) NOT NULL,
    round_number INT NOT NULL,
    total_contributions DECIMAL(20,2),
    interest_earned DECIMAL(20,2),
    status VARCHAR(30) DEFAULT 'PENDING',
    block_number BIGINT,
    block_timestamp TIMESTAMP,
    gas_used BIGINT,
    gas_price DECIMAL(20,10),
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (pool_id) REFERENCES savings_pools(id),
    FOREIGN KEY (receiver_id) REFERENCES users(id),
    FOREIGN KEY (wallet_id) REFERENCES wallets(id),
    INDEX idx_payouts_pool_id (pool_id),
    INDEX idx_payouts_receiver_id (receiver_id),
    INDEX idx_payouts_status (status)
);

-- =============================================
-- 7. REPUTATION SCORES TABLE
-- =============================================
CREATE TABLE reputation_scores (
    user_id UUID PRIMARY KEY,
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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- =============================================
-- 8. AUDIT LOGS TABLE
-- =============================================
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(100),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    request_id UUID,
    status VARCHAR(20) DEFAULT 'SUCCESS',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_audit_logs_user_id (user_id),
    INDEX idx_audit_logs_created_at (created_at DESC),
    INDEX idx_audit_logs_action (action),
    INDEX idx_audit_logs_status (status)
);

-- =============================================
-- 9. NOTIFICATIONS TABLE
-- =============================================
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    title VARCHAR(255),
    content TEXT,
    metadata JSONB,
    status VARCHAR(20) DEFAULT 'PENDING',
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_notifications_user_id (user_id),
    INDEX idx_notifications_status (status),
    INDEX idx_notifications_created_at (created_at DESC)
);

-- =============================================
-- 10. SETTINGS TABLE
-- =============================================
CREATE TABLE settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    key VARCHAR(100) NOT NULL,
    value JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE (user_id, key),
    INDEX idx_settings_user_id (user_id)
);
```

---

## Smart Contracts

### AjoraPool.sol - Complete Contract

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";

/**
 * @title AjoraPool
 * @dev Decentralized rotating savings and credit association (ROSCA)
 */
contract AjoraPool is ReentrancyGuard, Pausable, Ownable {
    using EnumerableSet for EnumerableSet.AddressSet;

    // =============================================
    // STRUCTS
    // =============================================
    
    struct PoolConfig {
        uint256 contributionAmount;
        uint256 totalSlots;
        uint256 totalRounds;
        uint256 contributionFrequency;
        uint256 interestRate;
        uint256 startTime;
        uint256 endTime;
        bool isFlexible;
    }

    struct Member {
        uint256 totalContributed;
        uint256 totalPayouts;
        uint256 payoutRound;
        uint256 penaltyPoints;
        bool isActive;
        bool hasReceivedPayout;
    }

    struct RoundInfo {
        uint256 totalContributions;
        uint256 contributionsCount;
        address winner;
        bool isProcessed;
    }

    // =============================================
    // STATE VARIABLES
    // =============================================
    
    PoolConfig public config;
    EnumerableSet.AddressSet private members;
    mapping(address => Member) public membersInfo;
    mapping(uint256 => RoundInfo) public rounds;
    mapping(uint256 => mapping(address => uint256)) public roundContributions;
    mapping(address => uint256[]) public contributionHistory;
    
    uint256 public currentRound;
    bool public isComplete;
    address public governanceToken;
    
    uint256 public totalContributions;
    uint256 public totalPayouts;
    uint256 public totalMembers;
    uint256 public activeMembers;

    // =============================================
    // EVENTS
    // =============================================
    
    event PoolCreated(
        address indexed creator,
        uint256 contributionAmount,
        uint256 totalSlots,
        uint256 totalRounds
    );
    
    event MemberJoined(
        address indexed member,
        uint256 slotNumber,
        uint256 timestamp
    );
    
    event ContributionMade(
        address indexed member,
        uint256 round,
        uint256 amount,
        uint256 timestamp
    );
    
    event PayoutProcessed(
        address indexed receiver,
        uint256 round,
        uint256 amount,
        uint256 timestamp
    );
    
    event MemberDropped(
        address indexed member,
        uint256 penaltyPoints,
        uint256 timestamp
    );
    
    event PoolCompleted(
        address indexed creator,
        uint256 totalContributions,
        uint256 totalPayouts,
        uint256 timestamp
    );

    // =============================================
    // MODIFIERS
    // =============================================
    
    modifier onlyActiveMember() {
        require(members.contains(msg.sender), "Not a member");
        require(membersInfo[msg.sender].isActive, "Member inactive");
        _;
    }

    modifier onlyActivePool() {
        require(!isComplete, "Pool completed");
        require(block.timestamp <= config.endTime, "Pool ended");
        _;
    }

    // =============================================
    // CONSTRUCTOR
    // =============================================
    
    constructor(
        uint256 _contributionAmount,
        uint256 _totalSlots,
        uint256 _totalRounds,
        uint256 _contributionFrequency,
        uint256 _interestRate,
        uint256 _startTime,
        uint256 _endTime,
        address _governanceToken,
        bool _isFlexible
    ) {
        require(_totalSlots > 0 && _totalSlots <= 1000, "Invalid slots");
        require(_totalRounds > 0 && _totalRounds <= 52, "Invalid rounds");
        require(_contributionAmount > 0, "Invalid contribution");
        require(_startTime > block.timestamp, "Invalid start time");
        require(_endTime > _startTime, "Invalid end time");

        config = PoolConfig({
            contributionAmount: _contributionAmount,
            totalSlots: _totalSlots,
            totalRounds: _totalRounds,
            contributionFrequency: _contributionFrequency,
            interestRate: _interestRate,
            startTime: _startTime,
            endTime: _endTime,
            isFlexible: _isFlexible
        });

        governanceToken = _governanceToken;
        currentRound = 1;
        totalMembers = 0;
        activeMembers = 0;

        emit PoolCreated(
            msg.sender,
            _contributionAmount,
            _totalSlots,
            _totalRounds
        );
    }

    // =============================================
    // PUBLIC FUNCTIONS
    // =============================================
    
    /**
     * @dev Join the savings pool
     */
    function joinPool() external payable whenNotPaused {
        require(block.timestamp < config.startTime, "Pool started");
        require(members.length() < config.totalSlots, "Pool full");
        require(!members.contains(msg.sender), "Already joined");

        if (config.contributionAmount > 0 && !config.isFlexible) {
            require(
                msg.value == config.contributionAmount,
                "Invalid contribution amount"
            );
        }

        members.add(msg.sender);
        membersInfo[msg.sender] = Member({
            totalContributed: msg.value,
            totalPayouts: 0,
            payoutRound: 0,
            penaltyPoints: 0,
            isActive: true,
            hasReceivedPayout: false
        });

        totalMembers++;
        activeMembers++;
        totalContributions += msg.value;

        emit MemberJoined(msg.sender, members.length(), block.timestamp);
    }

    /**
     * @dev Make a contribution to the pool
     */
    function contribute() 
        external 
        payable 
        whenNotPaused 
        nonReentrant 
        onlyActiveMember 
        onlyActivePool 
    {
        require(block.timestamp >= config.startTime, "Pool not started");
        require(block.timestamp <= config.endTime, "Pool not ended");
        require(!rounds[currentRound].isProcessed, "Round processed");

        uint256 deadline = config.startTime + 
            (currentRound - 1) * config.contributionFrequency;
        
        if (!config.isFlexible) {
            require(
                block.timestamp <= deadline + 7 days,
                "Round deadline passed"
            );
            require(
                msg.value == config.contributionAmount,
                "Invalid contribution amount"
            );
        }

        // Update member info
        membersInfo[msg.sender].totalContributed += msg.value;
        roundContributions[currentRound][msg.sender] += msg.value;
        contributionHistory[msg.sender].push(block.timestamp);

        // Update round info
        rounds[currentRound].totalContributions += msg.value;
        rounds[currentRound].contributionsCount++;

        totalContributions += msg.value;

        emit ContributionMade(msg.sender, currentRound, msg.value, block.timestamp);
    }

    /**
     * @dev Process payout for the current round
     */
    function processPayout() 
        external 
        whenNotPaused 
        nonReentrant 
        onlyOwner 
        onlyActivePool 
    {
        require(!rounds[currentRound].isProcessed, "Round already processed");
        require(
            block.timestamp >= config.endTime || 
            currentRound > config.totalRounds,
            "Pool not complete"
        );

        // Determine winner
        address winner = getWinner();
        require(winner != address(0), "No winner available");

        // Calculate payout amount
        uint256 totalPool = address(this).balance;
        uint256 payoutAmount = (totalPool * (10000 + config.interestRate)) / 10000;

        require(payoutAmount > 0, "Insufficient funds");

        // Update winner info
        membersInfo[winner].totalPayouts += payoutAmount;
        membersInfo[winner].hasReceivedPayout = true;
        membersInfo[winner].payoutRound = currentRound;

        // Update round info
        rounds[currentRound].winner = winner;
        rounds[currentRound].isProcessed = true;

        // Send payout
        (bool success, ) = winner.call{value: payoutAmount}("");
        require(success, "Payout failed");

        totalPayouts += payoutAmount;

        emit PayoutProcessed(winner, currentRound, payoutAmount, block.timestamp);

        // Move to next round
        currentRound++;

        // Check if pool is complete
        if (currentRound > config.totalRounds) {
            isComplete = true;
            emit PoolCompleted(
                owner(),
                totalContributions,
                totalPayouts,
                block.timestamp
            );
        }
    }

    /**
     * @dev Drop a member from the pool (admin only)
     */
    function dropMember(address member) 
        external 
        whenNotPaused 
        onlyOwner 
    {
        require(members.contains(member), "Member not found");
        require(membersInfo[member].isActive, "Already inactive");

        membersInfo[member].isActive = false;
        membersInfo[member].penaltyPoints += 1;
        activeMembers--;

        emit MemberDropped(member, membersInfo[member].penaltyPoints, block.timestamp);
    }

    // =============================================
    // VIEW FUNCTIONS
    // =============================================
    
    /**
     * @dev Get the winner for the current round
     */
    function getWinner() public view returns (address) {
        address[] memory memberList = members.values();
        uint256 startIndex = (currentRound - 1) % memberList.length;

        for (uint256 i = 0; i < memberList.length; i++) {
            uint256 index = (startIndex + i) % memberList.length;
            address candidate = memberList[index];
            if (membersInfo[candidate].isActive && 
                !membersInfo[candidate].hasReceivedPayout) {
                return candidate;
            }
        }
        return address(0);
    }

    /**
     * @dev Get member count
     */
    function getMemberCount() external view returns (uint256) {
        return members.length();
    }

    /**
     * @dev Get all members
     */
    function getMembers() external view returns (address[] memory) {
        return members.values();
    }

    /**
     * @dev Get pool status
     */
    function getPoolStatus() external view returns (
        uint256 totalContributions,
        uint256 totalMembers,
        uint256 activeMembers,
        uint256 poolBalance,
        uint256 currentRound,
        bool isComplete
    ) {
        return (
            totalContributions,
            members.length(),
            activeMembers,
            address(this).balance,
            currentRound,
            isComplete
        );
    }

    /**
     * @dev Get member contribution history
     */
    function getContributionHistory(address member) 
        external 
        view 
        returns (uint256[] memory) 
    {
        require(members.contains(member), "Not a member");
        return contributionHistory[member];
    }

    /**
     * @dev Get round contribution for a member
     */
    function getRoundContribution(uint256 round, address member) 
        external 
        view 
        returns (uint256) 
    {
        return roundContributions[round][member];
    }

    // =============================================
    // EMERGENCY FUNCTIONS
    // =============================================
    
    /**
     * @dev Emergency withdrawal (only owner)
     */
    function emergencyWithdraw() external onlyOwner {
        require(isComplete || block.timestamp > config.endTime + 30 days, 
            "Not eligible for emergency withdrawal");
        
        uint256 balance = address(this).balance;
        require(balance > 0, "No balance");

        (bool success, ) = owner().call{value: balance}("");
        require(success, "Withdrawal failed");
    }

    /**
     * @dev Pause the contract
     */
    function pause() external onlyOwner {
        _pause();
    }

    /**
     * @dev Unpause the contract
     */
    function unpause() external onlyOwner {
        _unpause();
    }

    // =============================================
    // FALLBACK FUNCTIONS
    // =============================================
    
    receive() external payable {}
    
    fallback() external payable {}
}
```

---

## API Reference

### REST API Endpoints

#### Authentication Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/auth/register` | Register new user | No |
| POST | `/api/v1/auth/login` | User login | No |
| POST | `/api/v1/auth/refresh` | Refresh access token | Yes |
| POST | `/api/v1/auth/logout` | User logout | Yes |
| POST | `/api/v1/auth/verify-email` | Verify email address | No |
| POST | `/api/v1/auth/reset-password` | Request password reset | No |
| PUT | `/api/v1/auth/change-password` | Change password | Yes |
| POST | `/api/v1/auth/mfa/enable` | Enable MFA | Yes |
| POST | `/api/v1/auth/mfa/verify` | Verify MFA code | Yes |
| POST | `/api/v1/auth/mfa/disable` | Disable MFA | Yes |
| GET | `/api/v1/auth/sessions` | List active sessions | Yes |
| DELETE | `/api/v1/auth/sessions/{id}` | Revoke session | Yes |

#### User Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/v1/users/{id}` | Get user profile | Yes |
| PUT | `/api/v1/users/{id}` | Update user profile | Yes |
| DELETE | `/api/v1/users/{id}` | Deactivate account | Yes |
| POST | `/api/v1/users/{id}/kyc` | Submit KYC documents | Yes |
| GET | `/api/v1/users/{id}/kyc` | Get KYC status | Yes |
| POST | `/api/v1/users/{id}/wallet` | Create wallet | Yes |
| GET | `/api/v1/users/{id}/wallet` | Get wallet info | Yes |

#### Pool Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/pools` | Create pool | Yes |
| GET | `/api/v1/pools` | List pools | Yes |
| GET | `/api/v1/pools/{id}` | Get pool details | Yes |
| PUT | `/api/v1/pools/{id}` | Update pool | Yes |
| DELETE | `/api/v1/pools/{id}` | Cancel pool | Yes |
| POST | `/api/v1/pools/{id}/join` | Join pool | Yes |
| POST | `/api/v1/pools/{id}/leave` | Leave pool | Yes |
| GET | `/api/v1/pools/{id}/members` | Get pool members | Yes |
| GET | `/api/v1/pools/{id}/rounds` | Get round history | Yes |

#### Contribution Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/contributions` | Submit contribution | Yes |
| GET | `/api/v1/contributions/{id}` | Get contribution | Yes |
| GET | `/api/v1/users/{id}/contributions` | Get user contributions | Yes |
| GET | `/api/v1/pools/{id}/contributions` | Get pool contributions | Yes |
| POST | `/api/v1/contributions/retry` | Retry failed contribution | Yes |

#### Blockchain Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/blockchain/deploy` | Deploy contract | Yes |
| POST | `/api/v1/blockchain/transact` | Send transaction | Yes |
| GET | `/api/v1/blockchain/tx/{hash}` | Get transaction status | Yes |
| GET | `/api/v1/blockchain/events` | Get contract events | Yes |
| GET | `/api/v1/blockchain/gas` | Get gas price | Yes |

### WebSocket API

#### Connection
```
ws://api.ajora.com/api/v1/ws
```

#### Authentication
```json
{
  "type": "auth",
  "token": "jwt_access_token"
}
```

#### Subscribe to Events
```json
{
  "type": "subscribe",
  "channels": ["pool_updates", "notifications", "contributions"]
}
```

#### Event Types
| Event | Description | Data |
|-------|-------------|------|
| `pool.created` | New pool created | Pool details |
| `pool.updated` | Pool updated | Updated fields |
| `pool.completed` | Pool completed | Final statistics |
| `member.joined` | User joined pool | Member details |
| `member.left` | User left pool | Member details |
| `contribution.made` | Contribution made | Contribution details |
| `payout.processed` | Payout processed | Payout details |
| `notification.new` | New notification | Notification content |

### Request/Response Examples

#### Register User
**Request**:
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "john.doe@example.com",
  "phone": "+2348012345678",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe",
  "date_of_birth": "1990-01-01"
}
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "kyc_status": "PENDING",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Create Pool
**Request**:
```http
POST /api/v1/pools
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "Community Savings Group",
  "description": "Monthly savings for community projects",
  "pool_type": "FIXED",
  "total_slots": 20,
  "contribution_amount": 100,
  "contribution_frequency": "MONTHLY",
  "total_rounds": 12,
  "start_date": "2024-02-01T00:00:00Z",
  "end_date": "2025-01-31T23:59:59Z",
  "interest_rate": 5.0
}
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "pool_id": "123e4567-e89b-12d3-a456-426614174001",
    "contract_address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "name": "Community Savings Group",
    "status": "DRAFT",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Submit Contribution
**Request**:
```http
POST /api/v1/contributions
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "pool_id": "123e4567-e89b-12d3-a456-426614174001",
  "amount": 100,
  "round_number": 1,
  "wallet_address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
}
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "contribution_id": "123e4567-e89b-12d3-a456-426614174002",
    "transaction_hash": "0xabc123...",
    "amount": 100,
    "status": "CONFIRMED",
    "confirmed_at": "2024-01-15T10:35:00Z"
  }
}
```

---

## Security

### Authentication & Authorization

#### JWT Token Structure
```json
{
  "header": {
    "alg": "RS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "user_id",
    "email": "user@example.com",
    "role": "USER",
    "permissions": ["read", "write", "admin"],
    "iat": 1516239022,
    "exp": 1516242622,
    "iss": "ajora-auth",
    "aud": "ajora-api"
  },
  "signature": "RSASHA256 signature"
}
```

#### RBAC Permissions
```yaml
roles:
  USER:
    permissions:
      - "users:read:self"
      - "users:update:self"
      - "pools:read"
      - "pools:create"
      - "pools:join"
      - "contributions:create"
      - "contributions:read"
      
  ADMIN:
    permissions:
      - "users:*"
      - "pools:*"
      - "contributions:*"
      - "blockchain:*"
      - "audit:*"
      
  SUPER_ADMIN:
    permissions:
      - "*"
```

### Encryption Standards

#### Data at Rest
- **Database**: AES-256 encryption
- **S3**: Server-side encryption (SSE-S3)
- **Secrets**: AWS Secrets Manager
- **Backups**: Encrypted snapshots

#### Data in Transit
- **TLS 1.3**: All external communications
- **mTLS**: Service-to-service communication
- **IPSec**: VPN connections

#### Key Management
- **AWS KMS**: Master keys
- **Key Rotation**: 90 days
- **HSM**: Hardware security module
- **Key Hierarchy**: 
  ```
  Master Key (KMS)
    ├── Database Key
    ├── Application Key
    ├── API Key
    └── JWT Key
  ```

### Security Headers
```http
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
Referrer-Policy: strict-origin-when-cross-origin
```

### Rate Limiting
```yaml
general:
  requests_per_second: 100
  burst_limit: 20
  authenticated: 1000/hour
  unauthenticated: 100/hour

endpoints:
  /auth/login:
    per_second: 10
    per_minute: 60
    per_hour: 1000
    
  /auth/register:
    per_hour: 50
    
  /api/v1/pools:
    per_second: 50
    per_minute: 300
    
  /api/v1/contributions:
    per_second: 100
    per_minute: 1000
```

---

## Deployment

### Local Development

#### Prerequisites
```bash
# Install required tools
brew install go rust terraform kubectl helm docker
brew install awscli jq postgresql redis

# Clone repository
git clone https://github.com/ajora/ajora.git
cd ajora
```

#### Setup Development Environment
```bash
# Run setup script
./scripts/setup-dev.sh

# This will:
# 1. Install dependencies
# 2. Start PostgreSQL, Redis, Kafka
# 3. Run migrations
# 4. Build services
```

#### Run Services Locally
```bash
# Build all services
make build

# Run with Docker Compose
docker-compose up -d

# Or run individually
./scripts/run-services.sh
```

### Docker Deployment

#### Build Images
```bash
# Build all services
docker-compose build

# Build specific service
docker build -t ajora-auth-service:latest -f services/auth-service/Dockerfile services/auth-service/
```

#### Run with Docker Compose
```bash
# Start services
docker-compose up -d

# Check logs
docker-compose logs -f

# Scale services
docker-compose up -d --scale auth-service=3

# Stop services
docker-compose down
```

### Kubernetes Deployment

#### Prerequisites
```bash
# Install tools
brew install kubectl helm terraform

# Configure AWS credentials
aws configure

# Update kubeconfig
aws eks update-kubeconfig --region us-west-2 --name ajora-production
```

#### Deploy Infrastructure
```bash
# Initialize Terraform
cd infrastructure/terraform
terraform init

# Plan infrastructure
terraform plan -var-file=production.tfvars

# Apply infrastructure
terraform apply -var-file=production.tfvars
```

#### Deploy Applications
```bash
# Create namespace
kubectl create namespace ajora-production

# Install with Helm
helm upgrade --install ajora ./deployments/helm/ajora \
  --namespace ajora-production \
  --values ./deployments/helm/ajora/values/production.yaml

# Check deployment status
kubectl get pods -n ajora-production
kubectl get svc -n ajora-production
```

### CI/CD Pipeline

#### GitHub Actions Workflow
```yaml
name: Deploy to Production

on:
  push:
    branches: [main]
    
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build and Push Images
        run: |
          docker build -t ${{ secrets.ECR_REPO }}/ajora:${{ github.sha }} .
          docker push ${{ secrets.ECR_REPO }}/ajora:${{ github.sha }}
          
      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/ajora \
            ajora=${{ secrets.ECR_REPO }}/ajora:${{ github.sha }}
          
      - name: Verify Deployment
        run: |
          kubectl rollout status deployment/ajora
```

---

## Monitoring & Observability

### Metrics (Prometheus)

#### System Metrics
```yaml
# CPU Usage
container_cpu_usage_seconds_total

# Memory Usage
container_memory_usage_bytes

# Network I/O
container_network_receive_bytes_total
container_network_transmit_bytes_total

# Disk I/O
container_fs_reads_bytes_total
container_fs_writes_bytes_total
```

#### Application Metrics
```yaml
# HTTP Requests
http_requests_total
http_request_duration_seconds
http_request_size_bytes
http_response_size_bytes

# Business Metrics
ajora_users_total
ajora_pools_total
ajora_contributions_total
ajora_payouts_total
ajora_transaction_volume
ajora_pool_participation_rate

# Blockchain Metrics
blockchain_transactions_total
blockchain_confirmations_duration
blockchain_gas_price
blockchain_gas_usage
```

#### Alert Rules
```yaml
groups:
  - name: service_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        annotations:
          summary: "High error rate detected"
          
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        annotations:
          summary: "Service is down"
          
      - alert: HighLatency
        expr: histogram_quantile(0.95, http_request_duration_seconds) > 1
        annotations:
          summary: "API latency above threshold"
```

### Logging (ELK Stack)

#### Log Format (JSON)
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "service": "auth-service",
  "trace_id": "abc123",
  "span_id": "def456",
  "user_id": "123e4567",
  "request_id": "789abc",
  "message": "User logged in successfully",
  "metadata": {
    "ip": "192.168.1.1",
    "user_agent": "Mozilla/5.0"
  }
}
```

#### Log Aggregation
```bash
# View logs for a specific service
kubectl logs -f deployment/auth-service -n production

# Search logs
grep "error" logs/auth-service.log

# Tail logs with filters
tail -f logs/ajora.log | grep "ERROR"
```

### Tracing (Jaeger)

#### Trace Spans
```go
// Add tracing to service
func (s *Service) ProcessRequest(ctx context.Context, req *Request) error {
    span, ctx := opentracing.StartSpanFromContext(ctx, "process_request")
    defer span.Finish()
    
    span.SetTag("user_id", req.UserID)
    span.SetTag("request_type", req.Type)
    
    // Child span
    childSpan, _ := opentracing.StartSpanFromContext(ctx, "db_query")
    defer childSpan.Finish()
    
    result := s.db.Query(ctx, req)
    
    return nil
}
```

### Dashboards (Grafana)

#### Dashboard URLs
```
System Dashboard: http://grafana:3000/d/ajora-system
Application Dashboard: http://grafana:3000/d/ajora-app
Business Dashboard: http://grafana:3000/d/ajora-business
Blockchain Dashboard: http://grafana:3000/d/ajora-blockchain
```

---

## Development Guide

### Project Structure
```
ajora/
├── services/
│   ├── api-gateway/           # Go - API Gateway
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── pkg/
│   │   └── Dockerfile
│   ├── auth-service/          # Go - Auth Service
│   ├── user-service/          # Go - User Service
│   ├── pool-service/          # Go - Pool Management
│   ├── contribution-service/  # Go - Contributions
│   ├── notification-service/  # Go - Notifications
│   ├── blockchain-service/    # Go - Blockchain Orchestration
│   ├── reputation-service/    # Go - Reputation
│   ├── wallet-signer/         # Rust - Cryptographic Signing
│   │   ├── src/
│   │   ├── Cargo.toml
│   │   └── Dockerfile
│   ├── tx-validator/          # Rust - Transaction Validation
│   ├── fraud-detection/       # Rust - Fraud Detection
│   └── crypto-audit/          # Rust - Audit & Verification
├── contracts/                 # Solidity Smart Contracts
│   ├── AjoraPool.sol
│   └── test/
├── infrastructure/
│   ├── terraform/            # Infrastructure as Code
│   └── kubernetes/           # K8s Manifests
├── deployments/
│   ├── helm/                 # Helm Charts
│   └── docker-compose/
├── migrations/               # Database Migrations
├── docs/                     # Documentation
├── scripts/                  # Utility Scripts
├── .github/workflows/        # CI/CD
├── Makefile
├── README.md
└── LICENSE
```

### Coding Standards

#### Go
```go
// Package naming: lowercase, no underscores
package auth

// Struct naming: PascalCase
type AuthService struct {
    userRepo UserRepository
}

// Function naming: PascalCase for exported, camelCase for internal
func (s *AuthService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // Implementation
}

// Error handling
if err != nil {
    return nil, fmt.Errorf("failed to create user: %w", err)
}

// Use context for cancellation
func (s *AuthService) Process(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Process
    }
}
```

#### Rust
```rust
// Module naming: snake_case
mod auth_service;

// Struct naming: PascalCase
pub struct AuthService {
    user_repo: UserRepository,
}

// Function naming: snake_case
pub fn create_user(ctx: &Context, req: CreateUserRequest) -> Result<User, Error> {
    // Implementation
}

// Error handling with Result
pub fn validate(&self) -> Result<(), ValidationError> {
    // Implementation
}
```

#### Solidity
```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

// Contract naming: PascalCase
contract AjoraPool {
    // Events: PascalCase
    event PoolCreated(address indexed creator);
    
    // Functions: camelCase
    function createPool() external {}
}
```

### Testing Strategy

#### Unit Tests
```go
// services/auth-service/internal/service/auth_service_test.go
func TestAuthService_Register(t *testing.T) {
    // Setup
    repo := &mockUserRepository{}
    service := NewAuthService(repo)
    
    // Test
    user, err := service.Register(context.Background(), &RegisterRequest{
        Email:    "test@example.com",
        Password: "SecurePass123!",
    })
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
}
```

#### Integration Tests
```go
// services/auth-service/integration_test.go
func TestAuthIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    // Setup test Redis
    redis := setupTestRedis(t)
    defer redis.Close()
    
    // Test flow
    user := registerUser(t, db)
    token := loginUser(t, db, redis)
    verifyToken(t, token)
}
```

#### Contract Tests
```solidity
// contracts/test/AjoraPool.test.js
const { expect } = require("chai");

describe("AjoraPool", function() {
    it("Should create a pool", async function() {
        const Pool = await ethers.getContractFactory("AjoraPool");
        const pool = await Pool.deploy(100, 10, 12, 30, 5, ...);
        await pool.deployed();
        
        expect(await pool.getMemberCount()).to.equal(0);
    });
});
```

### Performance Testing

#### Load Testing (k6)
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '2m', target: 100 },
        { duration: '5m', target: 100 },
        { duration: '2m', target: 0 },
    ],
};

export default function() {
    let res = http.get('http://localhost:8080/health');
    check(res, { 'status was 200': (r) => r.status == 200 });
    sleep(1);
}
```

---

## Performance Optimization

### Database Optimization

#### Indexes
```sql
-- Composite indexes for common queries
CREATE INDEX idx_contributions_pool_status 
ON contributions(pool_id, status, created_at DESC);

-- Partial indexes for active data
CREATE INDEX idx_active_pools 
ON savings_pools(status, start_date) 
WHERE status = 'ACTIVE';

-- Covering indexes
CREATE INDEX idx_users_kyc_cover 
ON users(kyc_status, email, first_name, last_name) 
WHERE kyc_status = 'PENDING';
```

#### Query Optimization
```sql
-- Use CTEs for complex queries
WITH pool_stats AS (
    SELECT 
        pool_id,
        COUNT(*) as member_count,
        SUM(amount) as total_contributions
    FROM contributions
    WHERE created_at > NOW() - INTERVAL '30 days'
    GROUP BY pool_id
)
SELECT 
    p.*,
    ps.member_count,
    ps.total_contributions
FROM savings_pools p
LEFT JOIN pool_stats ps ON p.id = ps.pool_id
WHERE p.status = 'ACTIVE';
```

### Caching Strategy

#### Cache Layers
```yaml
Level 1: In-Memory Cache (Ristretto/FreeCache)
  - TTL: 5 minutes
  - Size: 1GB
  - Use: Frequently accessed data

Level 2: Redis Cache
  - TTL: 1 hour
  - Size: 10GB
  - Use: Session data, API responses

Level 3: CDN
  - TTL: 24 hours
  - Use: Static assets, images
```

#### Cache Invalidation
```go
// Cache invalidation on updates
func (s *PoolService) UpdatePool(ctx context.Context, pool *Pool) error {
    // Update database
    err := s.db.UpdatePool(ctx, pool)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    key := fmt.Sprintf("pool:%s", pool.ID)
    s.cache.Delete(ctx, key)
    
    // Publish invalidation event
    s.publishEvent(ctx, "pool.updated", pool)
    
    return nil
}
```

### Load Balancing

#### Service Discovery (Istio)
```yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: auth-service
spec:
  host: auth-service
  trafficPolicy:
    loadBalancer:
      simple: ROUND_ROBIN
    connectionPool:
      tcp:
        maxConnections: 100
    outlierDetection:
      consecutiveErrors: 5
      interval: 30s
      baseEjectionTime: 30s
```

---

## Troubleshooting

### Common Issues

#### Database Connection Issues
```bash
# Check PostgreSQL status
docker ps | grep postgres
kubectl get pods -n production | grep postgres

# Test connection
psql -h localhost -U ajora_admin -d ajora -c "SELECT 1"

# Restart PostgreSQL
docker restart ajora-postgres
kubectl rollout restart deployment/postgres -n production
```

#### Service Startup Issues
```bash
# Check service logs
kubectl logs -f deployment/auth-service -n production

# Check service status
kubectl get pods -n production
kubectl describe pod auth-service-xxxxx -n production

# Check service endpoints
kubectl get svc -n production
kubectl get endpoints -n production
```

#### Blockchain Issues
```bash
# Check RPC connection
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# Check gas price
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":1}'

# Check transaction status
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["0x..."],"id":1}'
```

### Debugging Guide

#### Enable Debug Logging
```bash
# Set log level to debug
export LOG_LEVEL=debug

# Enable tracing
export TRACING_ENABLED=true

# Start service with debug
./auth-service --debug
```

#### Profile Performance
```bash
# Enable pprof
go tool pprof http://localhost:8080/debug/pprof/profile

# Analyze heap
go tool pprof http://localhost:8080/debug/pprof/heap

# Monitor metrics
curl http://localhost:8080/metrics
```

#### Database Debugging
```sql
-- Enable query logging
SET log_statement = 'all';

-- Find slow queries
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Check locks
SELECT 
    pid,
    usename,
    application_name,
    state,
    query
FROM pg_stat_activity
WHERE waiting = true;
```

---

## Contributing

### Contribution Guidelines

1. **Fork the Repository**
2. **Create a Feature Branch**
3. **Write Tests**
4. **Update Documentation**
5. **Submit Pull Request**

### Development Workflow

```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Make changes
git add .
git commit -m "feat: add new feature"

# Run tests
make test
make lint

# Push changes
git push origin feature/your-feature-name
```

### Code Review Checklist

- [ ] Code follows style guidelines
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Security scan passes
- [ ] Documentation updated
- [ ] No breaking changes
- [ ] Performance impact assessed

---

## License

MIT License

Copyright (c) 2024 Ajora Platform

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

---

## Contact & Support

### Community
- **GitHub**: https://github.com/ajora/ajora
- **Discord**: https://discord.gg/ajora
- **Twitter**: https://twitter.com/ajora

### Documentation
- **API Docs**: https://api.ajora.com/docs
- **Deployment Guide**: https://docs.ajora.com/deployment
- **FAQ**: https://docs.ajora.com/faq

### Support
- **Email**: support@ajora.com
- **Slack**: https://ajora.slack.com
- **Issue Tracker**: https://github.com/ajora/ajora/issues

---

**Documentation Version**: 1.0.0  
**Last Updated**: January 2024  
**Maintainer**: Ajora Development Team
```

This comprehensive documentation covers everything needed to understand, deploy, and maintain the Ajora platform. It includes detailed architecture diagrams, complete API specifications, security guidelines, and operational procedures. The documentation is structured for different audiences including developers, operators, and business stakeholders.