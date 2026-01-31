-- 000004_create_unclaimed_prizes.down.sql
-- 미수령 당첨금 테이블 삭제 (롤백)

DROP INDEX IF EXISTS idx_unclaimed_prizes_status;
DROP INDEX IF EXISTS idx_unclaimed_prizes_expiration_date;
DROP INDEX IF EXISTS idx_unclaimed_prizes_prize_rank;
DROP INDEX IF EXISTS idx_unclaimed_prizes_draw_no;

DROP TABLE IF EXISTS unclaimed_prizes;
