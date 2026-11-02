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

// ReadStatus fetches current system status for a device, initializing it if absent.
func (r *Repository) ReadStatus(deviceID string) (model.SystemStatus, error) {
	var s model.SystemStatus
	var pumpOn int
	var updatedAt time.Time

	err := r.db.QueryRow(`
		SELECT device_id, pump_on, moisture, mode, threshold, updated_at
		FROM system_status WHERE device_id = ?
		LIMIT 1
	`, deviceID).Scan(&s.DeviceID, &pumpOn, &s.Moisture, &s.Mode, &s.Threshold, &updatedAt)

	if err == sql.ErrNoRows {
		_, _ = r.db.Exec(`
			INSERT IGNORE INTO system_status (device_id, pump_on, moisture, mode, threshold)
			VALUES (?, 0, 0, 'auto', 50)
		`, deviceID)
		return model.SystemStatus{
			DeviceID:  deviceID,
			PumpOn:    false,
			Moisture:  0,
			Mode:      "auto",
			Threshold: 50,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	if err != nil {
		return s, err
	}

	s.PumpOn = pumpOn == 1
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

