CREATE TABLE IF NOT EXISTS lotto_draws (
    id SERIAL PRIMARY KEY,
    draw_no INTEGER NOT NULL UNIQUE,
    draw_date DATE NOT NULL,
    num1 INTEGER NOT NULL,
    num2 INTEGER NOT NULL,
    num3 INTEGER NOT NULL,
    num4 INTEGER NOT NULL,
    num5 INTEGER NOT NULL,
    num6 INTEGER NOT NULL,
    bonus_num INTEGER NOT NULL,
    first_prize BIGINT NOT NULL,
    first_winners INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS lotto_number_stats (
    id SERIAL PRIMARY KEY,
    number INTEGER NOT NULL UNIQUE,
    total_count INTEGER DEFAULT 0,
    bonus_count INTEGER DEFAULT 0,
    last_draw_no INTEGER DEFAULT 0,
    calculated_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS lotto_reappear_stats (
    number INTEGER PRIMARY KEY,
    total_appear INTEGER DEFAULT 0,
    reappear_count INTEGER DEFAULT 0,
    probability DOUBLE PRECISION DEFAULT 0,
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Auth Tables (migrations/001_create_auth_tables.sql 내용 통합)

CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    device_id       VARCHAR(255) UNIQUE,
    email           VARCHAR(255) UNIQUE,
    password_hash   VARCHAR(255),
    is_member       BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT REFERENCES users(id) ON DELETE CASCADE,
    token           VARCHAR(512) UNIQUE NOT NULL,
    expires_at      TIMESTAMP NOT NULL,
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS email_verifications (
    id              BIGSERIAL PRIMARY KEY,
    email           VARCHAR(255) NOT NULL,
    code            VARCHAR(6) NOT NULL,
    expires_at      TIMESTAMP NOT NULL,
    verified        BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_device_id ON users(device_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_email_verifications_email ON email_verifications(email);