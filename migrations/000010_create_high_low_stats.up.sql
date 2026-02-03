-- 010_create_high_low_stats.sql
-- 고저 비율 통계 테이블 (회차별 고번호:저번호 비율 확률 추이)
-- 고번호: 23~45, 저번호: 1~22

CREATE TABLE IF NOT EXISTS lotto_high_low_stats (
    draw_no         INTEGER PRIMARY KEY,
    actual_ratio    VARCHAR(3) NOT NULL,              -- 해당 회차의 실제 비율 (0:6, 1:5, 2:4, 3:3, 4:2, 5:1, 6:0)
    count_0_6       INTEGER NOT NULL DEFAULT 0,       -- 고0:저6 누적 횟수
    count_1_5       INTEGER NOT NULL DEFAULT 0,       -- 고1:저5 누적 횟수
    count_2_4       INTEGER NOT NULL DEFAULT 0,       -- 고2:저4 누적 횟수
    count_3_3       INTEGER NOT NULL DEFAULT 0,       -- 고3:저3 누적 횟수
    count_4_2       INTEGER NOT NULL DEFAULT 0,       -- 고4:저2 누적 횟수
    count_5_1       INTEGER NOT NULL DEFAULT 0,       -- 고5:저1 누적 횟수
    count_6_0       INTEGER NOT NULL DEFAULT 0,       -- 고6:저0 누적 횟수
    prob_0_6        DOUBLE PRECISION NOT NULL DEFAULT 0, -- 고0:저6 확률
    prob_1_5        DOUBLE PRECISION NOT NULL DEFAULT 0, -- 고1:저5 확률
    prob_2_4        DOUBLE PRECISION NOT NULL DEFAULT 0, -- 고2:저4 확률
    prob_3_3        DOUBLE PRECISION NOT NULL DEFAULT 0, -- 고3:저3 확률
    prob_4_2        DOUBLE PRECISION NOT NULL DEFAULT 0, -- 고4:저2 확률
    prob_5_1        DOUBLE PRECISION NOT NULL DEFAULT 0, -- 고5:저1 확률
    prob_6_0        DOUBLE PRECISION NOT NULL DEFAULT 0, -- 고6:저0 확률
    calculated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 조회 성능을 위한 인덱스
CREATE INDEX IF NOT EXISTS idx_high_low_stats_actual ON lotto_high_low_stats(actual_ratio);
