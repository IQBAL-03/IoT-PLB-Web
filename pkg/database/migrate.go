package database

import (
	"database/sql"
	"fmt"
)

// Migrate checks and initializes tables and columns required by PlantCare IoT.
func Migrate(conn *sql.DB) error {
	// 1. Users Table
	_, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id            INT AUTO_INCREMENT PRIMARY KEY,
			name          VARCHAR(100) NOT NULL,
			email         VARCHAR(150) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
		)
	`)
	if err != nil {
		return err
	}

	// 2. Devices Table
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS devices (
			device_id     VARCHAR(50)  PRIMARY KEY,
			pairing_code  VARCHAR(20)  NOT NULL DEFAULT '8821',
			user_id       INT          DEFAULT NULL,
			name          VARCHAR(100) NOT NULL DEFAULT 'Unit Tanaman',
			blynk_token   VARCHAR(64)  NOT NULL DEFAULT '',
			created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			INDEX idx_user (user_id)
		)
	`)
	if err != nil {
		return err
	}

	// Default device for immediate plug-and-play
	_, _ = conn.Exec(`
		INSERT IGNORE INTO devices (device_id, pairing_code, name, blynk_token)
		VALUES ('DEV-PLB-001', '8821', 'Unit Tanaman Utama', 'dKpUYSW9CoruTAYJMQa82CBf-VyH4h42')
	`)

	// 3. System Status Table
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS system_status (
			id          INT          NOT NULL DEFAULT 1,
			device_id   VARCHAR(50)  NOT NULL DEFAULT 'DEV-PLB-001',
			pump_on     TINYINT(1)   NOT NULL DEFAULT 0,
			moisture    DOUBLE       NOT NULL DEFAULT 0,
			mode        VARCHAR(10)  NOT NULL DEFAULT 'auto',
			threshold   DOUBLE       NOT NULL DEFAULT 50,
			updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
			PRIMARY KEY (id),
			UNIQUE KEY uq_dev (device_id)
		)
	`)
	if err != nil {
		return err
	}

	// Schema upgrades for existing installations
	addColumnIfNotExists(conn, "system_status", "device_id", "VARCHAR(50) NOT NULL DEFAULT 'DEV-PLB-001'")
	addUniqueIndexIfNotExists(conn, "system_status", "uq_dev", "device_id")

	_, _ = conn.Exec(`
		INSERT IGNORE INTO system_status (id, device_id, pump_on, moisture, mode, threshold)
		VALUES (1, 'DEV-PLB-001', 0, 0, 'auto', 50)
	`)

	// 4. Pump History Table
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS pump_history (
			id         INT AUTO_INCREMENT PRIMARY KEY,
			device_id  VARCHAR(50)  NOT NULL DEFAULT 'DEV-PLB-001',
			timestamp  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			action     VARCHAR(10)  NOT NULL,
			mode       VARCHAR(10)  NOT NULL,
			moisture   DOUBLE       NOT NULL DEFAULT 0,
			duration   VARCHAR(20)  NOT NULL DEFAULT '—',
			INDEX idx_dev (device_id, id DESC)
		)
	`)
	if err != nil {
		return err
	}

	addColumnIfNotExists(conn, "pump_history", "device_id", "VARCHAR(50) NOT NULL DEFAULT 'DEV-PLB-001'")
	return nil
}

func addColumnIfNotExists(conn *sql.DB, table, column, colDef string) {
	var count int
	err := conn.QueryRow(`
		SELECT COUNT(*) 
		FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`, table, column).Scan(&count)
	if err == nil && count == 0 {
		_, _ = conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, colDef))
	}
}

func addUniqueIndexIfNotExists(conn *sql.DB, table, indexName, colName string) {
	var count int
	err := conn.QueryRow(`
		SELECT COUNT(*) 
		FROM information_schema.STATISTICS 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?
	`, table, indexName).Scan(&count)
	if err == nil && count == 0 {
		_, _ = conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD UNIQUE KEY %s (%s)", table, indexName, colName))
	}
}

