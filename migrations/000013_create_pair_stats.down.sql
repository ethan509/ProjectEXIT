-- 013_create_pair_stats.down.sql
-- 번호 쌍 동시출현 통계 테이블 삭제

DROP INDEX IF EXISTS idx_pair_stats_prob;
DROP INDEX IF EXISTS idx_pair_stats_numbers;
DROP INDEX IF EXISTS idx_pair_stats_draw_no;
DROP TABLE IF EXISTS lotto_pair_stats;
