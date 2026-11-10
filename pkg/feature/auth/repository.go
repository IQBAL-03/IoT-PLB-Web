// Package auth handles user registration, credential verification, and sessions.
package auth

import (
	"database/sql"
	"strings"
)

// Repository manages user authentication records in the database.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new Auth Repository instance.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser inserts a new registered user.
func (r *Repository) CreateUser(name, email, passHash string) (int64, error) {
	res, err := r.db.Exec(`
		INSERT INTO users (name, email, password_hash)
		VALUES (?, ?, ?)
	`, name, strings.ToLower(strings.TrimSpace(email)), passHash)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetUserByEmail finds a user record by email address.
func (r *Repository) GetUserByEmail(email string) (int, string, string, error) {
	var uid int
	var name, storedHash string
	err := r.db.QueryRow(`
		SELECT id, name, password_hash FROM users WHERE email = ? LIMIT 1
	`, strings.ToLower(strings.TrimSpace(email))).Scan(&uid, &name, &storedHash)
	return uid, name, storedHash, err
}

// UpsertGoogleUser retrieves existing user or provisions a new one from Google token.
func (r *Repository) UpsertGoogleUser(name, email, defaultHash string) (int, error) {
	var uid int
	err := r.db.QueryRow("SELECT id FROM users WHERE email = ? LIMIT 1", email).Scan(&uid)
	if err == sql.ErrNoRows {
		res, err := r.db.Exec(`
			INSERT INTO users (name, email, password_hash)
			VALUES (?, ?, ?)
		`, name, email, defaultHash)
		if err != nil {
			return 0, err
		}
		newID, _ := res.LastInsertId()
		return int(newID), nil
	} else if err != nil {
		return 0, err
	}
	return uid, nil
}

// SaveOTP records an OTP code for an email address with expiration time.
func (r *Repository) SaveOTP(email, code string, expiresAt string) error {
	_, err := r.db.Exec(`
		INSERT INTO otp_codes (email, code, expires_at, is_used)
		VALUES (?, ?, ?, 0)
	`, strings.ToLower(strings.TrimSpace(email)), strings.TrimSpace(code), expiresAt)
	return err
}

// VerifyAndConsumeOTP checks if the OTP is valid, unexpired, and unused, then marks it as used.
func (r *Repository) VerifyAndConsumeOTP(email, code string) (bool, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	cleanCode := strings.TrimSpace(code)

	var otpID int
	err := r.db.QueryRow(`
		SELECT id FROM otp_codes 
		WHERE email = ? AND code = ? AND is_used = 0 AND expires_at > NOW(3)
		ORDER BY id DESC LIMIT 1
	`, cleanEmail, cleanCode).Scan(&otpID)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Mark as used
	_, _ = r.db.Exec("UPDATE otp_codes SET is_used = 1 WHERE id = ?", otpID)
	return true, nil
}

// UpsertEmailUser retrieves existing user or provisions a new user record.
func (r *Repository) UpsertEmailUser(name, email, defaultHash string) (int, string, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	var uid int
	var existingName string

	err := r.db.QueryRow("SELECT id, name FROM users WHERE email = ? LIMIT 1", cleanEmail).Scan(&uid, &existingName)
	if err == sql.ErrNoRows {
		if strings.TrimSpace(name) == "" {
			parts := strings.Split(cleanEmail, "@")
			name = parts[0]
		}
		res, err := r.db.Exec(`
			INSERT INTO users (name, email, password_hash)
			VALUES (?, ?, ?)
		`, name, cleanEmail, defaultHash)
		if err != nil {
			return 0, "", err
		}
		newID, _ := res.LastInsertId()
		return int(newID), name, nil
	} else if err != nil {
		return 0, "", err
	}
	return uid, existingName, nil
}
