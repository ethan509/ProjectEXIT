-- 디바이스 토큰 (FCM push용)
CREATE TABLE IF NOT EXISTS device_tokens (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token       TEXT NOT NULL,
    platform    VARCHAR(20) NOT NULL DEFAULT 'android',  -- android, ios, web
    is_active   BOOLEAN DEFAULT true,
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, token)
);

-- 당첨 확인 결과
CREATE TABLE IF NOT EXISTS winning_checks (
    id                  BIGSERIAL PRIMARY KEY,
    recommendation_id   BIGINT NOT NULL REFERENCES lotto_recommendations(id) ON DELETE CASCADE,
    user_id             BIGINT REFERENCES users(id) ON DELETE SET NULL,
    draw_no             INT NOT NULL,
    matched_numbers     INTEGER[] NOT NULL,
    matched_count       INT NOT NULL,
    bonus_matched       BOOLEAN DEFAULT false,
    prize_rank          INT,
    created_at          TIMESTAMP DEFAULT NOW()
);

-- 푸시 알림 기록
CREATE TABLE IF NOT EXISTS push_notifications (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT REFERENCES users(id) ON DELETE SET NULL,
    type            VARCHAR(50) NOT NULL,
    title           TEXT NOT NULL,
    body            TEXT NOT NULL,
    data            JSONB,
    status          VARCHAR(20) DEFAULT 'pending',
    error_message   TEXT,
    created_at      TIMESTAMP DEFAULT NOW(),
    sent_at         TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_device_tokens_user ON device_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_winning_checks_draw ON winning_checks(draw_no);
CREATE INDEX IF NOT EXISTS idx_winning_checks_user ON winning_checks(user_id);
CREATE INDEX IF NOT EXISTS idx_winning_checks_rank ON winning_checks(prize_rank) WHERE prize_rank IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_push_notifications_user ON push_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_push_notifications_status ON push_notifications(status);
