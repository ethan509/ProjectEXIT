-- 007_create_analysis_stats.sql
-- 통합 분석 통계 테이블 (회차별, 번호별)

CREATE TABLE IF NOT EXISTS lotto_analysis_stats (
    draw_no         INTEGER NOT NULL,
    number          SMALLINT NOT NULL,

    -- 기본 통계 (기존 number_stats)
    total_count     INTEGER DEFAULT 0,      -- 누적 출현 횟수
    bonus_count     INTEGER DEFAULT 0,      -- 보너스 출현 횟수

    -- 재등장 확률 (기존 reappear_stats)
    reappear_total  INTEGER DEFAULT 0,      -- 재등장 기준 총 출현
    reappear_count  INTEGER DEFAULT 0,      -- 재등장 횟수
    reappear_prob   DECIMAL(5,4) DEFAULT 0, -- 재등장 확률

    -- 베이지안 확률 (기존 bayesian_stats)
    bayesian_prior  DECIMAL(10,8),          -- 사전 확률
    bayesian_post   DECIMAL(10,8),          -- 사후 확률

    -- 공통
    appeared        BOOLEAN NOT NULL,       -- 해당 회차 출현 여부
    calculated_at   TIMESTAMP DEFAULT NOW(),
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (draw_no, number)
);

-- 인덱스
CREATE INDEX IF NOT EXISTS idx_analysis_stats_draw_no ON lotto_analysis_stats(draw_no);
CREATE INDEX IF NOT EXISTS idx_analysis_stats_number ON lotto_analysis_stats(number);
CREATE INDEX IF NOT EXISTS idx_analysis_stats_draw_desc ON lotto_analysis_stats(draw_no DESC, number);
