-- Zam daily history: per-user, per-day, per-tx_type aggregation
CREATE TABLE IF NOT EXISTS zam_daily_history (
    user_id       BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    history_date  DATE        NOT NULL,
    tx_type       VARCHAR(30) NOT NULL,
    total_amount  BIGINT      NOT NULL DEFAULT 0,
    tx_count      INTEGER     NOT NULL DEFAULT 0,
    updated_at    TIMESTAMP   NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, history_date, tx_type)
);

CREATE INDEX IF NOT EXISTS idx_zam_daily_history_user_date ON zam_daily_history(user_id, history_date);
CREATE INDEX IF NOT EXISTS idx_zam_daily_history_date ON zam_daily_history(history_date);
