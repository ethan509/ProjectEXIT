-- 009_add_total_prob_column.sql
-- 번호별 출현 확률 컬럼 추가

ALTER TABLE lotto_analysis_stats
ADD COLUMN IF NOT EXISTS total_prob DOUBLE PRECISION DEFAULT 0;  -- 해당 번호의 출현 확률 (total_count / (draw_no * 6))
