CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS bakers (
    address VARCHAR(50) PRIMARY KEY,
    first_seen TIMESTAMP NOT NULL,
    last_seen TIMESTAMP NOT NULL,
    total_delegations_received BIGINT DEFAULT 0,
    unique_delegators INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS delegations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    delegator VARCHAR(50) NOT NULL,
    baker_id VARCHAR(50) NOT NULL,

    amount BIGINT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    level BIGINT NOT NULL,
    operation_hash VARCHAR(100) UNIQUE,

    is_new_delegation BOOLEAN DEFAULT FALSE,
    previous_baker VARCHAR(50),

    created_at TIMESTAMP DEFAULT NOW(),
    indexed_at TIMESTAMP DEFAULT NOW()
);

ALTER TABLE delegations ADD CONSTRAINT fk_baker_id FOREIGN KEY (baker_id) REFERENCES bakers(address);

CREATE INDEX idx_delegations_timestamp ON delegations(timestamp DESC );
CREATE INDEX idx_delegations_delegator ON delegations(delegator);
CREATE INDEX idx_delegations_baker ON delegations(baker_id);
CREATE INDEX idx_delegations_level ON delegations(level);
CREATE INDEX idx_delegations_amount ON delegations(amount DESC);

CREATE INDEX idx_delegations_date ON delegations(DATE(timestamp));

