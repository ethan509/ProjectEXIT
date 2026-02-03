-- 011_add_user_profile.sql
-- 회원 프로필 정보 컬럼 추가

-- 성별 타입
-- M: 남성, F: 여성, O: 기타
ALTER TABLE users ADD COLUMN IF NOT EXISTS gender VARCHAR(1);

-- 생년월일
ALTER TABLE users ADD COLUMN IF NOT EXISTS birth_date DATE;

-- 거주지역 (시/도)
ALTER TABLE users ADD COLUMN IF NOT EXISTS region VARCHAR(50);

-- 닉네임
ALTER TABLE users ADD COLUMN IF NOT EXISTS nickname VARCHAR(20);

-- 로또 구매 빈도
-- WEEKLY: 주1회, MONTHLY: 월1회, BIMONTHLY: 월2~3회, IRREGULAR: 비정기
ALTER TABLE users ADD COLUMN IF NOT EXISTS purchase_frequency VARCHAR(20);

-- 인덱스
CREATE INDEX IF NOT EXISTS idx_users_gender ON users(gender);
CREATE INDEX IF NOT EXISTS idx_users_region ON users(region);
CREATE INDEX IF NOT EXISTS idx_users_nickname ON users(nickname);
