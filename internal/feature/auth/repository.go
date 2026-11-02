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
