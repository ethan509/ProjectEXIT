-- 000001_create_auth_tables.down.sql
-- Rollback: 인증 관련 테이블 삭제

DROP INDEX IF EXISTS idx_email_verifications_email;
DROP INDEX IF EXISTS idx_refresh_tokens_token;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_device_id;

DROP TABLE IF EXISTS email_verifications;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
