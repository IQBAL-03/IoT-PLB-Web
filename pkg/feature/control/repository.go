// Package control manages hardware actuators (pump) and operational modes.
package control

import (
	"database/sql"
	"time"

	"plantcare-iot/pkg/model"
)

// Repository manages pump state and operational mode data.
type Repository struct {
	db *sql.DB
}

// NewRepository initializes a Control Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ReadStatus fetches current system status for a device.
// Enforces default standby state (pump OFF, moisture nil) if device is unlinked or IoT is offline.
func (r *Repository) ReadStatus(deviceID string, userID *int) (model.SystemStatus, error) {
	var s model.SystemStatus
	var pumpOn int
	var moistureVal sql.NullFloat64
	var hasDataVal sql.NullInt64
	var updatedAt time.Time

	err := r.db.QueryRow(`
		SELECT device_id, pump_on, moisture, COALESCE(has_data, 0), mode, threshold, updated_at
		FROM system_status WHERE device_id = ?
		LIMIT 1
	`, deviceID).Scan(&s.DeviceID, &pumpOn, &moistureVal, &hasDataVal, &s.Mode, &s.Threshold, &updatedAt)

	if err == sql.ErrNoRows {
		_, _ = r.db.Exec(`
			INSERT IGNORE INTO system_status (device_id, pump_on, moisture, has_data, mode, threshold)
			VALUES (?, 0, NULL, 0, 'auto', 50)
		`, deviceID)
		return model.SystemStatus{
			DeviceID:  deviceID,
			PumpOn:    false,
			Moisture:  nil,
			HasData:   false,
			IsOnline:  false,
			IsLinked:  false,
			HasOwner:  false,
			Mode:      "auto",
			Threshold: 50,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	if err != nil {
		return s, err
	}

	// Verify device ownership linking
	var ownerID sql.NullInt64
	devErr := r.db.QueryRow(`
		SELECT user_id FROM devices WHERE device_id = ? LIMIT 1
	`, deviceID).Scan(&ownerID)

	hasOwner := devErr == nil && ownerID.Valid
	isLinked := false
	if hasOwner && userID != nil && int(ownerID.Int64) == *userID {
		isLinked = true
	}
	s.HasOwner = hasOwner
	s.IsLinked = isLinked

	// Verify IoT connectivity: considered online if telemetry received within last 60 seconds
	isRecent := time.Since(updatedAt) < 60*time.Second
	isOnline := hasDataVal.Valid && hasDataVal.Int64 == 1 && isRecent
	s.IsOnline = isOnline

	// If account is not linked to device OR IoT failed to be detected, apply default standby state
	if !isLinked || !isOnline {
		s.Moisture = nil
		s.PumpOn = false
		s.HasData = false
	} else if moistureVal.Valid {
		m := moistureVal.Float64
		s.Moisture = &m
		s.HasData = true
		s.PumpOn = pumpOn == 1
	}

	s.Timestamp = updatedAt.UTC().Format(time.RFC3339)
	return s, nil
}

// WritePump updates the physical pump state flag in the database.
func (r *Repository) WritePump(deviceID string, on bool) error {
	pumpVal := 0
	if on {
		pumpVal = 1
	}
	_, err := r.db.Exec(`
		INSERT INTO system_status (device_id, pump_on, updated_at)
		VALUES (?, ?, NOW(3))
		ON DUPLICATE KEY UPDATE pump_on = VALUES(pump_on), updated_at = NOW(3)
	`, deviceID, pumpVal)
	return err
}

// WriteMode toggles the device between 'auto' and 'manual' operation modes.
func (r *Repository) WriteMode(deviceID, mode string) error {
	_, err := r.db.Exec(`
		INSERT INTO system_status (device_id, mode, updated_at)
		VALUES (?, ?, NOW(3))
		ON DUPLICATE KEY UPDATE mode = VALUES(mode), updated_at = NOW(3)
	`, deviceID, mode)
	return err
}

// CheckOwnership checks if a user is the authorized owner of a device.
// Returns (isOwner bool, hasOwner bool, err error)
func (r *Repository) CheckOwnership(deviceID string, userID *int) (bool, bool, error) {
	var ownerID sql.NullInt64
	err := r.db.QueryRow("SELECT user_id FROM devices WHERE device_id = ? LIMIT 1", deviceID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}

	hasOwner := ownerID.Valid
	isOwner := hasOwner && userID != nil && int(ownerID.Int64) == *userID
	return isOwner, hasOwner, nil
}

