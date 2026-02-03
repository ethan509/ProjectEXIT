-- 014_create_consecutive_stats.down.sql
-- 연번 통계 테이블 삭제

DROP INDEX IF EXISTS idx_consecutive_stats_actual;
DROP TABLE IF EXISTS lotto_consecutive_stats;
