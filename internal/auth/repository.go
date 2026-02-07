package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token expired")
	ErrTierNotFound  = errors.New("tier not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// User methods

func (r *Repository) CreateGuestUser(ctx context.Context, deviceID string) (*User, error) {
	query := `
		INSERT INTO users (device_id, lotto_tier, created_at, updated_at)
		VALUES ($1, 1, NOW(), NOW())
		RETURNING id, device_id, email, password_hash, lotto_tier, created_at, updated_at
	`
	user := &User{}
	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(
		&user.ID, &user.DeviceID, &user.Email, &user.PasswordHash,
		&user.LottoTier, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UserProfileInput 회원가입 시 프로필 정보
type UserProfileInput struct {
	Gender            Gender
	BirthDate         time.Time
	Region            *string
	Nickname          *string
	PurchaseFrequency *PurchaseFrequency
}

func (r *Repository) CreateMemberUser(ctx context.Context, email, passwordHash string, profile *UserProfileInput) (*User, error) {
	query := `
		INSERT INTO users (email, password_hash, lotto_tier, gender, birth_date, region, nickname, purchase_frequency, created_at, updated_at)
		VALUES ($1, $2, 2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, device_id, email, password_hash, lotto_tier, gender, birth_date, region, nickname, purchase_frequency, created_at, updated_at
	`
	user := &User{}
	err := r.db.QueryRowContext(ctx, query, email, passwordHash, profile.Gender, profile.BirthDate, profile.Region, profile.Nickname, profile.PurchaseFrequency).Scan(
		&user.ID, &user.DeviceID, &user.Email, &user.PasswordHash,
		&user.LottoTier, &user.Gender, &user.BirthDate, &user.Region, &user.Nickname, &user.PurchaseFrequency,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT u.id, u.device_id, u.email, u.password_hash, u.lotto_tier,
		       u.gender, u.birth_date, u.region, u.nickname, u.purchase_frequency,
		       u.created_at, u.updated_at,
		       t.id, t.code, t.name, t.level, t.description, t.is_active
		FROM users u
		JOIN membership_tiers t ON u.lotto_tier = t.id
		WHERE u.id = $1
	`
	user := &User{}
	tier := &MembershipTier{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.DeviceID, &user.Email, &user.PasswordHash,
		&user.LottoTier, &user.Gender, &user.BirthDate, &user.Region, &user.Nickname, &user.PurchaseFrequency,
		&user.CreatedAt, &user.UpdatedAt,
		&tier.ID, &tier.Code, &tier.Name, &tier.Level, &tier.Description, &tier.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	user.Tier = tier
	return user, nil
}

func (r *Repository) GetUserByDeviceID(ctx context.Context, deviceID string) (*User, error) {
	query := `
		SELECT u.id, u.device_id, u.email, u.password_hash, u.lotto_tier,
		       u.gender, u.birth_date, u.region, u.nickname, u.purchase_frequency,
		       u.created_at, u.updated_at,
		       t.id, t.code, t.name, t.level, t.description, t.is_active
		FROM users u
		JOIN membership_tiers t ON u.lotto_tier = t.id
		WHERE u.device_id = $1
	`
	user := &User{}
	tier := &MembershipTier{}
	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(
		&user.ID, &user.DeviceID, &user.Email, &user.PasswordHash,
		&user.LottoTier, &user.Gender, &user.BirthDate, &user.Region, &user.Nickname, &user.PurchaseFrequency,
		&user.CreatedAt, &user.UpdatedAt,
		&tier.ID, &tier.Code, &tier.Name, &tier.Level, &tier.Description, &tier.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	user.Tier = tier
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT u.id, u.device_id, u.email, u.password_hash, u.lotto_tier,
		       u.gender, u.birth_date, u.region, u.nickname, u.purchase_frequency,
		       u.created_at, u.updated_at,
		       t.id, t.code, t.name, t.level, t.description, t.is_active
		FROM users u
		JOIN membership_tiers t ON u.lotto_tier = t.id
		WHERE u.email = $1
	`
	user := &User{}
	tier := &MembershipTier{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.DeviceID, &user.Email, &user.PasswordHash,
		&user.LottoTier, &user.Gender, &user.BirthDate, &user.Region, &user.Nickname, &user.PurchaseFrequency,
		&user.CreatedAt, &user.UpdatedAt,
		&tier.ID, &tier.Code, &tier.Name, &tier.Level, &tier.Description, &tier.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	user.Tier = tier
	return user, nil
}

func (r *Repository) LinkEmail(ctx context.Context, userID int64, email, passwordHash string) error {
	query := `
		UPDATE users SET email = $1, password_hash = $2, lotto_tier = 2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, email, passwordHash, userID)
	return err
}

func (r *Repository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, passwordHash, userID)
	return err
}

// Refresh Token methods

func (r *Repository) CreateRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) (*RefreshToken, error) {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, user_id, token, expires_at, created_at
	`
	rt := &RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, userID, token, expiresAt).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return rt, nil
}

func (r *Repository) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens WHERE token = $1
	`
	rt := &RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	if rt.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}
	return rt, nil
}

func (r *Repository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *Repository) DeleteUserRefreshTokens(ctx context.Context, userID int64) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// Email Verification methods

func (r *Repository) CreateEmailVerification(ctx context.Context, email, code string, expiresAt time.Time) (*EmailVerification, error) {
	query := `
		INSERT INTO email_verifications (email, code, expires_at, verified, created_at)
		VALUES ($1, $2, $3, FALSE, NOW())
		RETURNING id, email, code, expires_at, verified, created_at
	`
	ev := &EmailVerification{}
	err := r.db.QueryRowContext(ctx, query, email, code, expiresAt).Scan(
		&ev.ID, &ev.Email, &ev.Code, &ev.ExpiresAt, &ev.Verified, &ev.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ev, nil
}

func (r *Repository) VerifyEmail(ctx context.Context, email, code string) error {
	query := `
		UPDATE email_verifications
		SET verified = TRUE
		WHERE email = $1 AND code = $2 AND expires_at > NOW() AND verified = FALSE
	`
	result, err := r.db.ExecContext(ctx, query, email, code)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("invalid or expired verification code")
	}
	return nil
}

func (r *Repository) IsEmailVerified(ctx context.Context, email string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM email_verifications
			WHERE email = $1 AND verified = TRUE
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

// MembershipTier methods

func (r *Repository) GetTierByID(ctx context.Context, id int) (*MembershipTier, error) {
	query := `
		SELECT id, code, name, level, description, is_active, created_at, updated_at
		FROM membership_tiers WHERE id = $1
	`
	tier := &MembershipTier{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tier.ID, &tier.Code, &tier.Name, &tier.Level, &tier.Description,
		&tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTierNotFound
	}
	if err != nil {
		return nil, err
	}
	return tier, nil
}

func (r *Repository) GetTierByCode(ctx context.Context, code string) (*MembershipTier, error) {
	query := `
		SELECT id, code, name, level, description, is_active, created_at, updated_at
		FROM membership_tiers WHERE code = $1
	`
	tier := &MembershipTier{}
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&tier.ID, &tier.Code, &tier.Name, &tier.Level, &tier.Description,
		&tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTierNotFound
	}
	if err != nil {
		return nil, err
	}
	return tier, nil
}

func (r *Repository) GetAllTiers(ctx context.Context) ([]MembershipTier, error) {
	query := `
		SELECT id, code, name, level, description, is_active, created_at, updated_at
		FROM membership_tiers
		WHERE is_active = TRUE
		ORDER BY level ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiers []MembershipTier
	for rows.Next() {
		var tier MembershipTier
		err := rows.Scan(
			&tier.ID, &tier.Code, &tier.Name, &tier.Level, &tier.Description,
			&tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tiers = append(tiers, tier)
	}
	return tiers, rows.Err()
}

func (r *Repository) UpdateUserTier(ctx context.Context, userID int64, tierID int) error {
	query := `UPDATE users SET lotto_tier = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, tierID, userID)
	return err
}
