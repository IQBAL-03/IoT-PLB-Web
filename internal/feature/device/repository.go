// Package device handles multi-unit IoT device registration, querying, and claiming.
package device

import (
	"database/sql"
	"time"

	"plantcare-iot/internal/model"
)

// Repository handles database operations for devices.
type Repository struct {
	db *sql.DB
}

// NewRepository initializes a Device Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List returns devices accessible to the current user.
func (r *Repository) List(userID *int) ([]model.DeviceItem, error) {
	var rows *sql.Rows
	var err error

	if userID != nil {
		rows, err = r.db.Query(`
			SELECT device_id, name, (user_id IS NOT NULL) as has_owner, (user_id = ?) as is_owner, created_at
			FROM devices
			WHERE user_id = ? OR user_id IS NULL
			ORDER BY is_owner DESC, device_id ASC
		`, *userID, *userID)
	} else {
		rows, err = r.db.Query(`
			SELECT device_id, name, (user_id IS NOT NULL) as has_owner, 0 as is_owner, created_at
			FROM devices
			ORDER BY device_id ASC
		`)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.DeviceItem
	for rows.Next() {
		var d model.DeviceItem
		var hasOwner, isOwner int
		var ts time.Time
		if err := rows.Scan(&d.DeviceID, &d.Name, &hasOwner, &isOwner, &ts); err == nil {
			d.HasOwner = hasOwner == 1
			d.IsOwner = isOwner == 1
			d.CreatedAt = ts.UTC().Format(time.RFC3339)
			list = append(list, d)
		}
	}

	if len(list) == 0 {
		list = []model.DeviceItem{
			{DeviceID: "DEV-PLB-001", Name: "Unit Tanaman Utama", IsOwner: true, HasOwner: false},
		}
	}
	return list, nil
}

// GetDeviceInfo fetches existing user ownership and pairing code for a device.
func (r *Repository) GetDeviceInfo(deviceID string) (sql.NullInt64, string, error) {
	var currentUserID sql.NullInt64
	var pairingCode string
	err := r.db.QueryRow(`
		SELECT user_id, pairing_code FROM devices WHERE device_id = ?
	`, deviceID).Scan(&currentUserID, &pairingCode)
	return currentUserID, pairingCode, err
}

// CreateWithClaim inserts a new device claimed by user.
func (r *Repository) CreateWithClaim(deviceID, pairingCode string, uid int, name string) error {
	_, err := r.db.Exec(`
		INSERT INTO devices (device_id, pairing_code, user_id, name)
		VALUES (?, ?, ?, ?)
	`, deviceID, pairingCode, uid, name)
	return err
}

// UpdateOwner transfers ownership of an existing device to user.
func (r *Repository) UpdateOwner(deviceID string, uid int, name string) error {
	_, err := r.db.Exec(`
		UPDATE devices SET user_id = ?, name = ? WHERE device_id = ?
	`, uid, name, deviceID)
	return err
}

// EnsureStatusRow makes sure a status row exists in system_status for the device.
func (r *Repository) EnsureStatusRow(deviceID string) {
	_, _ = r.db.Exec(`
		INSERT IGNORE INTO system_status (device_id, pump_on, moisture, mode, threshold)
		VALUES (?, 0, 0, 'auto', 50)
	`, deviceID)
}

