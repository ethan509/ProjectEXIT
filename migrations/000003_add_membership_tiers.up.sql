-- 003_add_membership_tiers.sql
-- 회원 등급 시스템 추가

-- membership_tiers 메타 테이블
CREATE TABLE IF NOT EXISTS membership_tiers (
    id              SERIAL PRIMARY KEY,
    code            VARCHAR(20) UNIQUE NOT NULL,  -- GUEST, MEMBER, GOLD, VIP
    name            VARCHAR(50) NOT NULL,          -- 표시 이름
    level           INT NOT NULL DEFAULT 0,        -- 등급 레벨 (낮을수록 낮은 등급)
    description     TEXT,                          -- 등급 설명
    is_active       BOOLEAN DEFAULT TRUE,          -- 활성화 여부
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

-- 기본 등급 데이터 삽입
INSERT INTO membership_tiers (code, name, level, description) VALUES
    ('GUEST', '게스트', 0, '회원가입하지 않은 사용자'),
    ('MEMBER', '정회원', 1, '회원가입한 사용자'),
    ('GOLD', '골드', 2, '월정액 구독 사용자'),
    ('VIP', 'VIP', 3, '특별 등급 사용자')
ON CONFLICT (code) DO NOTHING;

-- users 테이블에 tier_id 컬럼 추가
ALTER TABLE users ADD COLUMN IF NOT EXISTS tier_id INT;

-- 기존 데이터 마이그레이션: is_member 기반으로 tier_id 설정
UPDATE users SET tier_id = CASE
    WHEN is_member = TRUE THEN 2  -- MEMBER
    ELSE 1                         -- GUEST
END
WHERE tier_id IS NULL;

-- tier_id 기본값 설정 (GUEST)
ALTER TABLE users ALTER COLUMN tier_id SET DEFAULT 1;

-- NOT NULL 제약조건 추가
ALTER TABLE users ALTER COLUMN tier_id SET NOT NULL;

-- 외래키 설정
ALTER TABLE users ADD CONSTRAINT fk_users_tier_id
    FOREIGN KEY (tier_id) REFERENCES membership_tiers(id);

-- is_member 컬럼 삭제
ALTER TABLE users DROP COLUMN IF EXISTS is_member;

-- 인덱스
CREATE INDEX IF NOT EXISTS idx_users_tier_id ON users(tier_id);
CREATE INDEX IF NOT EXISTS idx_membership_tiers_code ON membership_tiers(code);
CREATE INDEX IF NOT EXISTS idx_membership_tiers_level ON membership_tiers(level);
