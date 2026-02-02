-- 010_add_bonus_prob_column.sql
-- 보너스 번호 출현 확률 컬럼 추가

ALTER TABLE lotto_analysis_stats
ADD COLUMN IF NOT EXISTS bonus_prob DOUBLE PRECISION DEFAULT 0;  -- 보너스 번호 출현 확률 (bonus_count / draw_no)
