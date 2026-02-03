-- 005_create_bayesian_stats.down.sql
-- 베이지안 통계 테이블 삭제

DROP INDEX IF EXISTS idx_bayesian_stats_draw_number;
DROP INDEX IF EXISTS idx_bayesian_stats_number;
DROP INDEX IF EXISTS idx_bayesian_stats_draw_no;
DROP TABLE IF EXISTS lotto_bayesian_stats;
