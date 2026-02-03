-- 013_create_pair_stats.up.sql
-- 번호 쌍 동시출현 통계 테이블

CREATE TABLE IF NOT EXISTS lotto_pair_stats (
    draw_no     INTEGER NOT NULL,
    number1     INTEGER NOT NULL,
    number2     INTEGER NOT NULL,
    count       INTEGER NOT NULL DEFAULT 0,
    prob        DOUBLE PRECISION NOT NULL DEFAULT 0,
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (draw_no, number1, number2),
    CHECK (number1 < number2),
    CHECK (number1 >= 1 AND number1 <= 44),
    CHECK (number2 >= 2 AND number2 <= 45)
);

-- 조회 성능을 위한 인덱스
CREATE INDEX IF NOT EXISTS idx_pair_stats_draw_no ON lotto_pair_stats(draw_no);
CREATE INDEX IF NOT EXISTS idx_pair_stats_numbers ON lotto_pair_stats(number1, number2);
CREATE INDEX IF NOT EXISTS idx_pair_stats_prob ON lotto_pair_stats(draw_no, prob DESC);
