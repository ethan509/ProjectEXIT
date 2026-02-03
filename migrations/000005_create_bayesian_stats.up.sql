-- 005_create_bayesian_stats.sql
-- 베이지안 추론 통계 테이블 (회차별, 번호별 누적 확률)

CREATE TABLE IF NOT EXISTS lotto_bayesian_stats (
    id              BIGSERIAL PRIMARY KEY,
    draw_no         INTEGER NOT NULL,        -- 회차 번호
    number          SMALLINT NOT NULL,       -- 번호 (1~45)
    total_count     INTEGER NOT NULL,        -- 1회차부터 해당 회차까지 누적 출현 횟수
    total_draws     INTEGER NOT NULL,        -- 총 회차 수 (draw_no와 동일)
    prior           DECIMAL(10,8) NOT NULL,  -- 사전 확률 (이전 회차의 posterior)
    posterior       DECIMAL(10,8) NOT NULL,  -- 사후 확률 (%)
    appeared        BOOLEAN NOT NULL,        -- 해당 회차에 출현했는지
    calculated_at   TIMESTAMP DEFAULT NOW(),
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW(),
    UNIQUE(draw_no, number)
);

-- 인덱스
CREATE INDEX IF NOT EXISTS idx_bayesian_stats_draw_no ON lotto_bayesian_stats(draw_no);
CREATE INDEX IF NOT EXISTS idx_bayesian_stats_number ON lotto_bayesian_stats(number);
CREATE INDEX IF NOT EXISTS idx_bayesian_stats_draw_number ON lotto_bayesian_stats(draw_no DESC, number);
