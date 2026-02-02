-- 009_add_total_prob_column.down.sql
-- 번호별 출현 확률 컬럼 제거

ALTER TABLE lotto_analysis_stats
DROP COLUMN IF EXISTS total_prob;
