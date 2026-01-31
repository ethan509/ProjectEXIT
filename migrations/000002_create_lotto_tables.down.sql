-- 000002_create_lotto_tables.down.sql
-- Rollback: 로또 관련 테이블 삭제

DROP INDEX IF EXISTS idx_lotto_reappear_stats_number;
DROP INDEX IF EXISTS idx_lotto_number_stats_number;
DROP INDEX IF EXISTS idx_lotto_draws_draw_date;
DROP INDEX IF EXISTS idx_lotto_draws_draw_no;

DROP TABLE IF EXISTS lotto_reappear_stats;
DROP TABLE IF EXISTS lotto_number_stats;
DROP TABLE IF EXISTS lotto_draws;
