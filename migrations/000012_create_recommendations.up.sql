-- 추천 이력 테이블
CREATE TABLE IF NOT EXISTS lotto_recommendations (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT REFERENCES users(id) ON DELETE SET NULL,
    method_codes    TEXT[] NOT NULL,
    numbers         INTEGER[] NOT NULL,
    bonus_number    INTEGER,
    confidence      DECIMAL(5,4),
    created_at      TIMESTAMP DEFAULT NOW()
);

-- 인덱스
CREATE INDEX IF NOT EXISTS idx_recommendations_user ON lotto_recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_recommendations_created ON lotto_recommendations(created_at DESC);
