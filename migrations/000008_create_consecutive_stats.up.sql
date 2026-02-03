-- 008_create_consecutive_stats.sql
-- 연번 통계 테이블 (회차별 연번 개수 확률 추이)

CREATE TABLE IF NOT EXISTS lotto_consecutive_stats (
    draw_no         INTEGER PRIMARY KEY,
    actual_count    INTEGER NOT NULL,           -- 해당 회차의 실제 연번 개수 (0,2,3,4,5,6)
    count_0         INTEGER NOT NULL DEFAULT 0, -- 연번 0개 누적 횟수
    count_2         INTEGER NOT NULL DEFAULT 0, -- 연번 2개 누적 횟수
    count_3         INTEGER NOT NULL DEFAULT 0, -- 연번 3개 누적 횟수
    count_4         INTEGER NOT NULL DEFAULT 0, -- 연번 4개 누적 횟수
    count_5         INTEGER NOT NULL DEFAULT 0, -- 연번 5개 누적 횟수
    count_6         INTEGER NOT NULL DEFAULT 0, -- 연번 6개 누적 횟수
    prob_0          DOUBLE PRECISION NOT NULL DEFAULT 0, -- 연번 0개 확률
    prob_2          DOUBLE PRECISION NOT NULL DEFAULT 0, -- 연번 2개 확률
    prob_3          DOUBLE PRECISION NOT NULL DEFAULT 0, -- 연번 3개 확률
    prob_4          DOUBLE PRECISION NOT NULL DEFAULT 0, -- 연번 4개 확률
    prob_5          DOUBLE PRECISION NOT NULL DEFAULT 0, -- 연번 5개 확률
    prob_6          DOUBLE PRECISION NOT NULL DEFAULT 0, -- 연번 6개 확률
    calculated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 조회 성능을 위한 인덱스
CREATE INDEX IF NOT EXISTS idx_consecutive_stats_actual ON lotto_consecutive_stats(actual_count);
