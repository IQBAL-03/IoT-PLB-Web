// Package config provides application configuration and environment variable loading.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strconv"
)

// generateEphemeralSecret creates a cryptographically secure 256-bit random string.
func generateEphemeralSecret(numBytes int) string {
	bytes := make([]byte, numBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "emergency-random-seed-2026"
	}
	return hex.EncodeToString(bytes)
}

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
	SMTPHost          string
	SMTPPort          string
	SMTPUser          string
	SMTPPass          string
	SMTPFrom          string
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
		log.Println("[SECURITY WARNING] JWT_SECRET environment variable is missing. Generated ephemeral runtime secret.")
		jwtSec = generateEphemeralSecret(32)
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

	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "587"
	}

	smtpFrom := os.Getenv("SMTP_FROM")
	if smtpFrom == "" {
		smtpFrom = os.Getenv("SMTP_USER")
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
		SMTPHost:          os.Getenv("SMTP_HOST"),
		SMTPPort:          smtpPort,
		SMTPUser:          os.Getenv("SMTP_USER"),
		SMTPPass:          os.Getenv("SMTP_PASS"),
		SMTPFrom:          smtpFrom,
	}
}


