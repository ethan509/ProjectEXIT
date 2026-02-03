-- 010_create_high_low_stats.down.sql
-- 고저 비율 통계 테이블 삭제

DROP INDEX IF EXISTS idx_high_low_stats_actual;
DROP TABLE IF EXISTS lotto_high_low_stats;
