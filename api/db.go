package api

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"os"
	"sync"

	"github.com/go-sql-driver/mysql"
)

var (
	db     *sql.DB
	dbOnce sync.Once
	dbErr  error
)

// getDB returns a singleton *sql.DB connected to TiDB Cloud (MySQL-compatible).
// Connection params are read from environment variables:
//
//	TIDB_HOST     – your TiDB Cloud host
//	TIDB_PORT     – e.g. 4000 (default)
//	TIDB_USER     – your TiDB Cloud username
//	TIDB_PASSWORD – your TiDB Cloud password
//	TIDB_DB       – database name, e.g. plantcare
func getDB() (*sql.DB, error) {
	dbOnce.Do(func() {
		host := os.Getenv("TIDB_HOST")
		port := os.Getenv("TIDB_PORT")
		user := os.Getenv("TIDB_USER")
		pass := os.Getenv("TIDB_PASSWORD")
		dbName := os.Getenv("TIDB_DB")

		if port == "" {
			port = "4000"
		}

		// Register TLS config required by TiDB Cloud public endpoint.
		tlsKey := "tidb"
		if err := mysql.RegisterTLSConfig(tlsKey, &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: host,
		}); err != nil {
			dbErr = fmt.Errorf("db: register TLS: %w", err)
			return
		}

		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?tls=%s&parseTime=true&loc=UTC&timeout=10s&readTimeout=10s&writeTimeout=10s",
			user, pass, host, port, dbName, tlsKey,
		)

		conn, err := sql.Open("mysql", dsn)
		if err != nil {
			dbErr = fmt.Errorf("db: open: %w", err)
			return
		}

		conn.SetMaxOpenConns(5)
		conn.SetMaxIdleConns(2)

		if err := conn.Ping(); err != nil {
			dbErr = fmt.Errorf("db: ping: %w", err)
			return
		}

		if err := migrate(conn); err != nil {
			dbErr = fmt.Errorf("db: migrate: %w", err)
			return
		}

		db = conn
	})
	return db, dbErr
}

// migrate creates tables if they don't exist yet.
func migrate(conn *sql.DB) error {
	_, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS system_status (
			id         INT          NOT NULL DEFAULT 1,
			pump_on    TINYINT(1)   NOT NULL DEFAULT 0,
			moisture   DOUBLE       NOT NULL DEFAULT 0,
			mode       VARCHAR(10)  NOT NULL DEFAULT 'auto',
			threshold  DOUBLE       NOT NULL DEFAULT 50,
			updated_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
				ON UPDATE CURRENT_TIMESTAMP(3),
			PRIMARY KEY (id)
		)
	`)
	if err != nil {
		return err
	}

	// Ensure the single status row exists.
	_, err = conn.Exec(`
		INSERT IGNORE INTO system_status (id, pump_on, moisture, mode, threshold)
		VALUES (1, 0, 0, 'auto', 50)
	`)
	if err != nil {
		return err
	}

	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS pump_history (
			id         INT AUTO_INCREMENT PRIMARY KEY,
			timestamp  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			action     VARCHAR(3)   NOT NULL,
			mode       VARCHAR(10)  NOT NULL,
			moisture   DOUBLE       NOT NULL DEFAULT 0,
			duration   VARCHAR(20)  NOT NULL DEFAULT '—'
		)
	`)
	return err
}

