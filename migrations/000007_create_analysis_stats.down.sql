-- 007_create_analysis_stats.down.sql
-- 통합 분석 통계 테이블 삭제

DROP INDEX IF EXISTS idx_analysis_stats_draw_desc;
DROP INDEX IF EXISTS idx_analysis_stats_number;
DROP INDEX IF EXISTS idx_analysis_stats_draw_no;
DROP TABLE IF EXISTS lotto_analysis_stats;
