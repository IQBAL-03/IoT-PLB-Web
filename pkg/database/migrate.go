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

	// Default device for immediate plug-and-play & simulation units
	_, _ = conn.Exec(`
		INSERT IGNORE INTO devices (device_id, pairing_code, name, blynk_token)
		VALUES 
			('DEV-PLB-001', '8821', 'Unit Tanaman Utama (Fisik)', 'dKpUYSW9CoruTAYJMQa82CBf-VyH4h42'),
			('DEV-PLB-002', '9932', 'Unit Simulasi 2 (Bonsai)', ''),
			('DEV-PLB-003', '7714', 'Unit Simulasi 3 (Cabai)', '')
	`)

	// 3. System Status Table
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS system_status (
			id          INT          NOT NULL DEFAULT 1,
			device_id   VARCHAR(50)  NOT NULL DEFAULT 'DEV-PLB-001',
			pump_on     TINYINT(1)   NOT NULL DEFAULT 0,
			moisture    DOUBLE       DEFAULT NULL,
			has_data    TINYINT(1)   NOT NULL DEFAULT 0,
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
	addColumnIfNotExists(conn, "system_status", "has_data", "TINYINT(1) NOT NULL DEFAULT 0")
	addUniqueIndexIfNotExists(conn, "system_status", "uq_dev", "device_id")

	// Allow NULL for moisture column
	_, _ = conn.Exec("ALTER TABLE system_status MODIFY COLUMN moisture DOUBLE DEFAULT NULL")

	// Insert initial row with zero/null defaults if missing
	_, _ = conn.Exec(`
		INSERT IGNORE INTO system_status (id, device_id, pump_on, moisture, has_data, mode, threshold)
		VALUES (1, 'DEV-PLB-001', 0, NULL, 0, 'auto', 50)
	`)

	// Reset default unit to standby state awaiting live IoT telemetry
	_, _ = conn.Exec(`
		UPDATE system_status
		SET pump_on = 0, moisture = NULL, has_data = 0
		WHERE has_data = 0
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

	// 5. OTP Codes Table for Email Authentication
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS otp_codes (
			id          INT AUTO_INCREMENT PRIMARY KEY,
			email       VARCHAR(150) NOT NULL,
			code        VARCHAR(10)  NOT NULL,
			expires_at  DATETIME(3)  NOT NULL,
			is_used     TINYINT(1)   NOT NULL DEFAULT 0,
			created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			INDEX idx_email_verify (email, code, is_used)
		)
	`)
	if err != nil {
		return err
	}

	// Ensure simulation units have rows in system_status
	_, _ = conn.Exec(`
		INSERT IGNORE INTO system_status (device_id, pump_on, moisture, has_data, mode, threshold)
		VALUES 
			('DEV-PLB-002', 0, NULL, 0, 'auto', 50),
			('DEV-PLB-003', 0, NULL, 0, 'auto', 50)
	`)

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
