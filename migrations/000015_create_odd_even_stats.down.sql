-- 015_create_odd_even_stats.down.sql
-- 홀짝 비율 통계 테이블 삭제

DROP INDEX IF EXISTS idx_odd_even_stats_actual;
DROP TABLE IF EXISTS lotto_odd_even_stats;
