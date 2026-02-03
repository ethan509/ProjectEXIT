-- 011_add_user_profile.down.sql
-- 회원 프로필 정보 컬럼 삭제

DROP INDEX IF EXISTS idx_users_nickname;
DROP INDEX IF EXISTS idx_users_region;
DROP INDEX IF EXISTS idx_users_gender;

ALTER TABLE users DROP COLUMN IF EXISTS purchase_frequency;
ALTER TABLE users DROP COLUMN IF EXISTS nickname;
ALTER TABLE users DROP COLUMN IF EXISTS region;
ALTER TABLE users DROP COLUMN IF EXISTS birth_date;
ALTER TABLE users DROP COLUMN IF EXISTS gender;
