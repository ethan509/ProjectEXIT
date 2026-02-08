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

-- 외래키 설정
ALTER TABLE users ADD CONSTRAINT fk_users_lotto_tier
    FOREIGN KEY (lotto_tier) REFERENCES membership_tiers(id);

-- 인덱스
CREATE INDEX IF NOT EXISTS idx_users_lotto_tier ON users(lotto_tier);
CREATE INDEX IF NOT EXISTS idx_membership_tiers_code ON membership_tiers(code);
CREATE INDEX IF NOT EXISTS idx_membership_tiers_level ON membership_tiers(level);
