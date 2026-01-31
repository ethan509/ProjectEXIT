-- 000004_create_unclaimed_prizes.up.sql
-- 미수령 당첨금 테이블 생성

CREATE TABLE IF NOT EXISTS unclaimed_prizes (
    id              BIGSERIAL PRIMARY KEY,
    draw_no         INTEGER NOT NULL,
    prize_rank      SMALLINT NOT NULL,  -- 1등 또는 2등
    amount          BIGINT NOT NULL,    -- 당첨금
    winner_name     VARCHAR(100),       -- 당첨자명 (마스킹됨, 예: 김**)
    winning_date    DATE NOT NULL,      -- 당첨일
    expiration_date DATE NOT NULL,      -- 만기일 (미수령 기한)
    status          VARCHAR(20) DEFAULT 'unclaimed',  -- 상태: unclaimed, claimed 등
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (draw_no) REFERENCES lotto_draws(draw_no) ON DELETE CASCADE
);

-- 인덱스
CREATE INDEX IF NOT EXISTS idx_unclaimed_prizes_draw_no ON unclaimed_prizes(draw_no);
CREATE INDEX IF NOT EXISTS idx_unclaimed_prizes_prize_rank ON unclaimed_prizes(prize_rank);
CREATE INDEX IF NOT EXISTS idx_unclaimed_prizes_expiration_date ON unclaimed_prizes(expiration_date);
CREATE INDEX IF NOT EXISTS idx_unclaimed_prizes_status ON unclaimed_prizes(status);
