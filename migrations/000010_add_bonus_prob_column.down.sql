-- 010_add_bonus_prob_column.down.sql
-- 보너스 번호 출현 확률 컬럼 제거

ALTER TABLE lotto_analysis_stats
DROP COLUMN IF EXISTS bonus_prob;
