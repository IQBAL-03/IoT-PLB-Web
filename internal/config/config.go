// Package config provides application configuration and environment variable loading.
package config

import (
	"os"
	"strconv"
)

// Config holds runtime configuration settings loaded from environment variables.
type Config struct {
	TiDBHost          string
	TiDBPort          string
	TiDBUser          string
	TiDBPassword      string
	TiDBDatabase      string
	JWTSecret         string
	AllowedOrigin     string
	MoistureThreshold float64
	DefaultDeviceID   string
}

// Load reads and parses configuration values from system environment variables.
func Load() Config {
	port := os.Getenv("TIDB_PORT")
	if port == "" {
		port = "4000"
	}

	dbName := os.Getenv("TIDB_DB")
	if dbName == "" {
		dbName = "test"
	}

	jwtSec := os.Getenv("JWT_SECRET")
	if jwtSec == "" {
		jwtSec = "plantcare-secret-key-2026-auth-token-tidb"
	}

	origin := os.Getenv("ALLOWED_ORIGIN")
	if origin == "" {
		origin = "*"
	}

	threshold := 50.0
	if envT := os.Getenv("MOISTURE_THRESHOLD"); envT != "" {
		if t, err := strconv.ParseFloat(envT, 64); err == nil {
			threshold = t
		}
	}

	return Config{
		TiDBHost:          os.Getenv("TIDB_HOST"),
		TiDBPort:          port,
		TiDBUser:          os.Getenv("TIDB_USER"),
		TiDBPassword:      os.Getenv("TIDB_PASSWORD"),
		TiDBDatabase:      dbName,
		JWTSecret:         jwtSec,
		AllowedOrigin:     origin,
		MoistureThreshold: threshold,
		DefaultDeviceID:   "DEV-PLB-001",
	}
}

