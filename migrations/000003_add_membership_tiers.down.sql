-- 000003_add_membership_tiers.down.sql
-- Rollback: 회원 등급 시스템 롤백

-- 외래키 제거
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_tier_id;

-- is_member 컬럼 복원
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_member BOOLEAN DEFAULT FALSE;

-- tier_id 기반으로 is_member 복원
UPDATE users SET is_member = CASE
    WHEN tier_id >= 2 THEN TRUE
    ELSE FALSE
END;

-- tier_id 컬럼 삭제
ALTER TABLE users DROP COLUMN IF EXISTS tier_id;

-- 인덱스 삭제
DROP INDEX IF EXISTS idx_membership_tiers_level;
DROP INDEX IF EXISTS idx_membership_tiers_code;
DROP INDEX IF EXISTS idx_users_tier_id;

-- membership_tiers 테이블 삭제
DROP TABLE IF EXISTS membership_tiers;
