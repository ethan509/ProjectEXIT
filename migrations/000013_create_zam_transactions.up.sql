-- Zam transactions table for history tracking
CREATE TABLE IF NOT EXISTS zam_transactions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount          BIGINT NOT NULL,                    -- positive for credit, negative for debit
    balance_after   BIGINT NOT NULL,                    -- balance after this transaction
    tx_type         VARCHAR(30) NOT NULL,               -- REGISTER_BONUS, DAILY_LOGIN, PURCHASE, REFUND, etc.
    description     TEXT,
    reference_id    VARCHAR(100),                       -- optional reference (e.g., order_id)
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_zam_transactions_user_id ON zam_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_zam_transactions_tx_type ON zam_transactions(tx_type);
CREATE INDEX IF NOT EXISTS idx_zam_transactions_created_at ON zam_transactions(created_at);
