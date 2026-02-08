-- 001_create_auth_tables.sql
-- 인증 관련 테이블 생성

-- users 테이블
CREATE TABLE IF NOT EXISTS users (
    id                   BIGSERIAL PRIMARY KEY,
    device_id            VARCHAR(255) UNIQUE,
    email                VARCHAR(255) UNIQUE,
    password_hash        VARCHAR(255),
    lotto_tier           INT DEFAULT 1 NOT NULL,
    gender               VARCHAR(1),
    birth_date           DATE,
    region               VARCHAR(50),
    nickname             VARCHAR(20),
    purchase_frequency   VARCHAR(20),
    zam_balance          BIGINT NOT NULL DEFAULT 0,
    last_daily_reward_at TIMESTAMP,
    created_at           TIMESTAMP DEFAULT NOW(),
    updated_at           TIMESTAMP DEFAULT NOW()
);

-- 외래키는 000003에서 설정

-- refresh_tokens 테이블
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT REFERENCES users(id) ON DELETE CASCADE,
    token           VARCHAR(512) UNIQUE NOT NULL,
    expires_at      TIMESTAMP NOT NULL,
    created_at      TIMESTAMP DEFAULT NOW()
);

-- email_verifications 테이블
CREATE TABLE IF NOT EXISTS email_verifications (
    id              BIGSERIAL PRIMARY KEY,
    email           VARCHAR(255) NOT NULL,
    code            VARCHAR(6) NOT NULL,
    expires_at      TIMESTAMP NOT NULL,
    verified        BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP DEFAULT NOW()
);

-- 인덱스
CREATE INDEX IF NOT EXISTS idx_users_device_id ON users(device_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_email_verifications_email ON email_verifications(email);
CREATE INDEX IF NOT EXISTS idx_users_gender ON users(gender);
CREATE INDEX IF NOT EXISTS idx_users_region ON users(region);
CREATE INDEX IF NOT EXISTS idx_users_nickname ON users(nickname);
