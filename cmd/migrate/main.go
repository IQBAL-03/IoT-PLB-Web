// Command line tool to manage PlantCare IoT database migrations.
// Usage:
//   go run cmd/migrate/main.go         # Run standard migration
//   go run cmd/migrate/main.go -fresh  # Drop all tables, re-create schema and seed data
package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"plantcare-iot/pkg/config"
	"plantcare-iot/pkg/database"
)

// loadEnvFile reads key-value pairs from .env into process environment variables.
func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
	}
}

func main() {
	fresh := flag.Bool("fresh", false, "Drop all existing tables and re-create schema with initial seeds")
	flag.Parse()

	loadEnvFile(".env")
	cfg := config.Load()

	if cfg.TiDBHost == "" || cfg.TiDBUser == "" {
		log.Fatalf("Error: TiDB configuration is missing. Ensure .env file exists or set environment variables.")
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?tls=true&charset=utf8mb4&parseTime=true&loc=UTC&timeout=15s",
		cfg.TiDBUser, cfg.TiDBPassword, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase,
	)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatalf("Error connecting to TiDB Cloud: %v", err)
	}

	if *fresh {
		log.Println("[MIGRATE] Menghapus seluruh tabel dan membangun ulang schema (migrate:fresh)...")
		if err := database.Reset(conn); err != nil {
			log.Fatalf("Error resetting database: %v", err)
		}
		log.Println("[MIGRATE] Berhasil: Seluruh tabel telah direset dan data bawaan (seed) berhasil dimasukkan.")
	} else {
		log.Println("[MIGRATE] Menjalankan migrasi standar...")
		if err := database.Migrate(conn); err != nil {
			log.Fatalf("Error migrating database: %v", err)
		}
		log.Println("[MIGRATE] Berhasil: Skema database telah diperbarui.")
	}
}

