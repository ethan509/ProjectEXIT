-- 002_create_lotto_tables.sql
-- 로또 당첨번호 및 분석 결과 테이블

-- lotto_draws: 로또 당첨번호 테이블
CREATE TABLE IF NOT EXISTS lotto_draws (
    draw_no         INTEGER PRIMARY KEY,
    draw_date       DATE NOT NULL,
    num1            SMALLINT NOT NULL,
    num2            SMALLINT NOT NULL,
    num3            SMALLINT NOT NULL,
    num4            SMALLINT NOT NULL,
    num5            SMALLINT NOT NULL,
    num6            SMALLINT NOT NULL,
    bonus_num       SMALLINT NOT NULL,
    first_prize     BIGINT,
    first_winners   INTEGER,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

-- lotto_number_stats: 각 번호별 통계 테이블
CREATE TABLE IF NOT EXISTS lotto_number_stats (
    id              BIGSERIAL PRIMARY KEY,
    number          SMALLINT UNIQUE NOT NULL,
    total_count     INTEGER DEFAULT 0,
    bonus_count     INTEGER DEFAULT 0,
    last_draw_no    INTEGER,
    calculated_at   TIMESTAMP DEFAULT NOW(),
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

-- lotto_reappear_stats: 번호 재등장 확률 통계 테이블
CREATE TABLE IF NOT EXISTS lotto_reappear_stats (
    id              BIGSERIAL PRIMARY KEY,
    number          SMALLINT NOT NULL,
    total_appear    INTEGER DEFAULT 0,
    reappear_count  INTEGER DEFAULT 0,
    probability     DECIMAL(5,4) DEFAULT 0,
    calculated_at   TIMESTAMP DEFAULT NOW(),
    UNIQUE(number)
);

-- 인덱스
CREATE INDEX IF NOT EXISTS idx_lotto_draws_draw_no ON lotto_draws(draw_no);
CREATE INDEX IF NOT EXISTS idx_lotto_draws_draw_date ON lotto_draws(draw_date);
CREATE INDEX IF NOT EXISTS idx_lotto_number_stats_number ON lotto_number_stats(number);
CREATE INDEX IF NOT EXISTS idx_lotto_reappear_stats_number ON lotto_reappear_stats(number);
