-- Remove zam transactions table
DROP TABLE IF EXISTS zam_transactions;

-- Remove zam columns from users
ALTER TABLE users DROP COLUMN IF EXISTS last_daily_reward_at;
ALTER TABLE users DROP COLUMN IF EXISTS zam_balance;
