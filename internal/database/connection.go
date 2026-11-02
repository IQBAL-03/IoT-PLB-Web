// Package database handles connection pooling and automated schema migration.
package database

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"plantcare-iot/internal/config"
)

var (
	db   *sql.DB
	dbMu sync.Mutex
)

// Get returns the singleton database connection pool, reconnecting if unhealthy.
func Get(cfg config.Config) (*sql.DB, error) {
	dbMu.Lock()
	defer dbMu.Unlock()

	if db != nil {
		if err := db.Ping(); err == nil {
			return db, nil
		}
		_ = db.Close()
		db = nil
	}

	if cfg.TiDBHost == "" || cfg.TiDBUser == "" {
		return nil, fmt.Errorf("variabel lingkungan database belum disetel di Vercel (TIDB_HOST=%q, TIDB_USER=%q)", cfg.TiDBHost, cfg.TiDBUser)
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?tls=true&charset=utf8mb4&parseTime=true&loc=UTC&timeout=10s&readTimeout=10s&writeTimeout=10s",
		cfg.TiDBUser, cfg.TiDBPassword, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase,
	)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}

	conn.SetMaxOpenConns(5)
	conn.SetMaxIdleConns(2)
	conn.SetConnMaxLifetime(4 * time.Minute)

	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	if err := Migrate(conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("db migrate: %w", err)
	}

	db = conn
	return db, nil
}

