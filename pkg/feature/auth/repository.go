// Package auth handles user registration, credential verification, and sessions.
package auth

import (
	"database/sql"
	"strings"
	"time"

	"plantcare-iot/pkg/model"
)

// Repository manages user authentication and device trust records in the database.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new Auth Repository instance.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetUserByEmail finds a user record by email address.
func (r *Repository) GetUserByEmail(email string) (*model.User, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	var u model.User
	var isVerifiedInt int

	err := r.db.QueryRow(`
		SELECT id, name, email, password_hash, auth_provider, is_verified, created_at
		FROM users
		WHERE email = ?
		LIMIT 1
	`, cleanEmail).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.AuthProvider,
		&isVerifiedInt,
		&u.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	u.IsVerified = isVerifiedInt == 1
	return &u, nil
}

// CreateUser inserts a new user record.
func (r *Repository) CreateUser(name, email, passHash, provider string, isVerified bool) (int, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	verifiedInt := 0
	if isVerified {
		verifiedInt = 1
	}

	res, err := r.db.Exec(`
		INSERT INTO users (name, email, password_hash, auth_provider, is_verified)
		VALUES (?, ?, ?, ?, ?)
	`, strings.TrimSpace(name), cleanEmail, passHash, provider, verifiedInt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	return int(id), err
}

// VerifyUser marks an existing user account as verified.
func (r *Repository) VerifyUser(userID int) error {
	_, err := r.db.Exec(`UPDATE users SET is_verified = 1 WHERE id = ?`, userID)
	return err
}

// IsDeviceTrusted checks whether a device token is registered for the specified user.
func (r *Repository) IsDeviceTrusted(userID int, deviceToken string) (bool, error) {
	cleanToken := strings.TrimSpace(deviceToken)
	if cleanToken == "" {
		return false, nil
	}

	var id int
	err := r.db.QueryRow(`
		SELECT id FROM trusted_devices
		WHERE user_id = ? AND device_token = ?
		LIMIT 1
	`, userID, cleanToken).Scan(&id)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Update last used timestamp
	_, _ = r.db.Exec(`UPDATE trusted_devices SET last_used_at = NOW(3) WHERE id = ?`, id)
	return true, nil
}

// TrustDevice records or updates a trusted device for a user.
func (r *Repository) TrustDevice(userID int, deviceToken, deviceName string) error {
	cleanToken := strings.TrimSpace(deviceToken)
	if cleanToken == "" {
		return nil
	}

	cleanName := strings.TrimSpace(deviceName)
	if cleanName == "" {
		cleanName = "Browser"
	}

	_, err := r.db.Exec(`
		INSERT INTO trusted_devices (user_id, device_token, device_name, created_at, last_used_at)
		VALUES (?, ?, ?, NOW(3), NOW(3))
		ON DUPLICATE KEY UPDATE device_name = VALUES(device_name), last_used_at = NOW(3)
	`, userID, cleanToken, cleanName)
	return err
}

// CheckOTPRateLimit enforces a 60-second cooldown between OTP sends for an email.
func (r *Repository) CheckOTPRateLimit(email string) (bool, int, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	var lastCreated time.Time

	err := r.db.QueryRow(`
		SELECT created_at FROM otp_codes
		WHERE email = ?
		ORDER BY id DESC
		LIMIT 1
	`, cleanEmail).Scan(&lastCreated)

	if err == sql.ErrNoRows {
		return true, 0, nil
	}
	if err != nil {
		return false, 0, err
	}

	elapsed := time.Since(lastCreated)
	if elapsed < 60*time.Second {
		waitSec := int((60*time.Second - elapsed).Seconds()) + 1
		return false, waitSec, nil
	}

	return true, 0, nil
}

// SaveOTP stores a newly generated OTP with its purpose and metadata.
func (r *Repository) SaveOTP(email, codeHash, purpose, metadata string, expiresAt time.Time) error {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))

	// Invalidate previous unused OTPs with same email and purpose
	_, _ = r.db.Exec(`
		UPDATE otp_codes SET is_used = 1
		WHERE email = ? AND purpose = ? AND is_used = 0
	`, cleanEmail, purpose)

	_, err := r.db.Exec(`
		INSERT INTO otp_codes (email, code_hash, purpose, metadata, attempts, expires_at, is_used)
		VALUES (?, ?, ?, ?, 0, ?, 0)
	`, cleanEmail, codeHash, purpose, metadata, expiresAt.Format("2006-01-02 15:04:05"))
	return err
}

// GetActiveOTP finds the latest active, unexpired, and unused OTP for an email and purpose.
func (r *Repository) GetActiveOTP(email, purpose string) (*model.OTPRecord, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	var rec model.OTPRecord
	var isUsedInt int

	err := r.db.QueryRow(`
		SELECT id, email, code_hash, purpose, COALESCE(metadata, ''), attempts, expires_at, is_used, created_at
		FROM otp_codes
		WHERE email = ? AND purpose = ? AND is_used = 0 AND expires_at > NOW(3)
		ORDER BY id DESC
		LIMIT 1
	`, cleanEmail, purpose).Scan(
		&rec.ID,
		&rec.Email,
		&rec.CodeHash,
		&rec.Purpose,
		&rec.Metadata,
		&rec.Attempts,
		&rec.ExpiresAt,
		&isUsedInt,
		&rec.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	rec.IsUsed = isUsedInt == 1
	return &rec, nil
}

// IncrementOTPAttempts increases the failed verification attempt count for an OTP.
func (r *Repository) IncrementOTPAttempts(otpID int) (int, error) {
	_, err := r.db.Exec(`UPDATE otp_codes SET attempts = attempts + 1 WHERE id = ?`, otpID)
	if err != nil {
		return 0, err
	}

	var attempts int
	err = r.db.QueryRow(`SELECT attempts FROM otp_codes WHERE id = ?`, otpID).Scan(&attempts)
	return attempts, err
}

// MarkOTPUsed invalidates an OTP code after successful verification or exceeding maximum attempts.
func (r *Repository) MarkOTPUsed(otpID int) error {
	_, err := r.db.Exec(`UPDATE otp_codes SET is_used = 1 WHERE id = ?`, otpID)
	return err
}
