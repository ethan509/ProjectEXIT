-- 006_create_analysis_stats.sql
-- 통합 분석 통계 테이블 (회차별, 번호별)

CREATE TABLE IF NOT EXISTS lotto_analysis_stats (
    draw_no         INTEGER NOT NULL,
    number          SMALLINT NOT NULL,

    -- 기본 통계
    total_count     INTEGER DEFAULT 0,              -- 누적 출현 횟수
    total_prob      DOUBLE PRECISION DEFAULT 0,     -- 출현 확률 (total_count / (draw_no * 6))
    bonus_count     INTEGER DEFAULT 0,              -- 보너스 출현 횟수
    bonus_prob      DOUBLE PRECISION DEFAULT 0,     -- 보너스 출현 확률 (bonus_count / draw_no)

    -- 첫번째/마지막 위치 통계
    first_count     INTEGER DEFAULT 0,              -- Num1(첫번째)로 나온 누적 횟수
    first_prob      DOUBLE PRECISION DEFAULT 0,     -- 첫번째 위치 확률
    last_count      INTEGER DEFAULT 0,              -- Num6(마지막)으로 나온 누적 횟수
    last_prob       DOUBLE PRECISION DEFAULT 0,     -- 마지막 위치 확률

    -- 재등장 확률
    reappear_total  INTEGER DEFAULT 0,              -- 재등장 기준 총 출현
    reappear_count  INTEGER DEFAULT 0,              -- 재등장 횟수
    reappear_prob   DECIMAL(5,4) DEFAULT 0,         -- 재등장 확률

    -- 베이지안 확률
    bayesian_prior  DECIMAL(10,8),                  -- 사전 확률
    bayesian_post   DECIMAL(10,8),                  -- 사후 확률

    -- 색상/행/열 통계
    color_count     INTEGER DEFAULT 0,              -- 해당 번호 색상의 총 출현 횟수
    color_prob      DOUBLE PRECISION DEFAULT 0,     -- 색상 출현 확률
    row_count       INTEGER DEFAULT 0,              -- 해당 번호 행의 총 출현 횟수
    row_prob        DOUBLE PRECISION DEFAULT 0,     -- 행 출현 확률
    col_count       INTEGER DEFAULT 0,              -- 해당 번호 열의 총 출현 횟수
    col_prob        DOUBLE PRECISION DEFAULT 0,     -- 열 출현 확률

    -- 공통
    appeared        BOOLEAN NOT NULL,               -- 해당 회차 출현 여부
    calculated_at   TIMESTAMP DEFAULT NOW(),
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (draw_no, number)
);

-- 인덱스
CREATE INDEX IF NOT EXISTS idx_analysis_stats_draw_no ON lotto_analysis_stats(draw_no);
CREATE INDEX IF NOT EXISTS idx_analysis_stats_number ON lotto_analysis_stats(number);
CREATE INDEX IF NOT EXISTS idx_analysis_stats_draw_desc ON lotto_analysis_stats(draw_no DESC, number);
