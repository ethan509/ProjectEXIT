-- 000005 마이그레이션 롤백

ALTER TABLE lotto_draws
DROP COLUMN IF EXISTS first_per_game,
DROP COLUMN IF EXISTS second_prize,
DROP COLUMN IF EXISTS second_winners,
DROP COLUMN IF EXISTS second_per_game,
DROP COLUMN IF EXISTS third_prize,
DROP COLUMN IF EXISTS third_winners,
DROP COLUMN IF EXISTS third_per_game,
DROP COLUMN IF EXISTS fourth_prize,
DROP COLUMN IF EXISTS fourth_winners,
DROP COLUMN IF EXISTS fourth_per_game,
DROP COLUMN IF EXISTS fifth_prize,
DROP COLUMN IF EXISTS fifth_winners,
DROP COLUMN IF EXISTS fifth_per_game;
