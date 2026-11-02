// Package history manages audit trails for pump activity and sensor events.
package history

import (
	"database/sql"
	"time"

	"plantcare-iot/pkg/model"
)

// Repository handles database operations for pump history.
type Repository struct {
	db *sql.DB
}

// NewRepository initializes a History Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Insert adds a new pump activity record.
func (r *Repository) Insert(deviceID, action, mode string, moisture float64) error {
	_, err := r.db.Exec(`
		INSERT INTO pump_history (device_id, action, mode, moisture, duration)
		VALUES (?, ?, ?, ?, '—')
	`, deviceID, action, mode, moisture)
	return err
}

// List retrieves recent history entries for a given device ID.
func (r *Repository) List(deviceID string) ([]model.HistoryEntry, error) {
	rows, err := r.db.Query(`
		SELECT id, device_id, timestamp, action, mode, moisture, duration
		FROM pump_history
		WHERE device_id = ?
		ORDER BY id DESC
		LIMIT 100
	`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.HistoryEntry
	for rows.Next() {
		var e model.HistoryEntry
		var ts time.Time
		if err := rows.Scan(&e.ID, &e.DeviceID, &ts, &e.Action, &e.Mode, &e.Moisture, &e.Duration); err != nil {
			return nil, err
		}
		e.Timestamp = ts.UTC().Format(time.RFC3339)
		list = append(list, e)
	}
	if list == nil {
		list = []model.HistoryEntry{}
	}
	return list, rows.Err()
}

// Clear removes all history entries for a device.
func (r *Repository) Clear(deviceID string) error {
	_, err := r.db.Exec("DELETE FROM pump_history WHERE device_id = ?", deviceID)
	return err
}

