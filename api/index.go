package api

import (
	"compress/gzip"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

// ─── Types ───────────────────────────────────────────────────────────────────

type SystemStatus struct {
	PumpOn    bool    `json:"pump_on"`
	Moisture  float64 `json:"moisture"`
	Mode      string  `json:"mode"`
	Threshold float64 `json:"threshold"`
	Timestamp string  `json:"timestamp"`
}

type HistoryEntry struct {
	ID        int     `json:"id"`
	Timestamp string  `json:"timestamp"`
	Action    string  `json:"action"`
	Mode      string  `json:"mode"`
	Moisture  float64 `json:"moisture"`
	Duration  string  `json:"duration"`
}

type ModeRequest struct {
	Mode string `json:"mode"`
}

// ─── Database Singleton & Migration ──────────────────────────────────────────

var (
	db     *sql.DB
	dbOnce sync.Once
	dbErr  error
)

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
		if dbName == "" {
			dbName = "test"
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
			action     VARCHAR(10)  NOT NULL,
			mode       VARCHAR(10)  NOT NULL,
			moisture   DOUBLE       NOT NULL DEFAULT 0,
			duration   VARCHAR(20)  NOT NULL DEFAULT '—'
		)
	`)
	return err
}

// ─── Middleware ───────────────────────────────────────────────────────────────

type gzipWriter struct {
	io.Writer
	http.ResponseWriter
}

func (g gzipWriter) Write(b []byte) (int, error) { return g.Writer.Write(b) }

func withGzip(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		next(gzipWriter{gz, w}, r)
	}
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := os.Getenv("ALLOWED_ORIGIN")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func chain(h http.HandlerFunc) http.HandlerFunc { return cors(withGzip(h)) }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func getThreshold() float64 {
	if envT := os.Getenv("MOISTURE_THRESHOLD"); envT != "" {
		if t, err := strconv.ParseFloat(envT, 64); err == nil {
			return t
		}
	}
	return 50.0
}

func dbError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database unavailable", "detail": err.Error()})
}

// ─── DB Operations ────────────────────────────────────────────────────────────

func readStatus() (SystemStatus, error) {
	conn, err := getDB()
	if err != nil {
		return SystemStatus{}, err
	}
	var s SystemStatus
	var pumpOn int
	var updatedAt time.Time
	err = conn.QueryRow(`
		SELECT pump_on, moisture, mode, threshold, updated_at
		FROM system_status WHERE id = 1
	`).Scan(&pumpOn, &s.Moisture, &s.Mode, &s.Threshold, &updatedAt)
	if err != nil {
		return s, err
	}
	s.PumpOn = pumpOn == 1
	s.Timestamp = updatedAt.UTC().Format(time.RFC3339)
	return s, nil
}

func writePump(on bool) error {
	conn, err := getDB()
	if err != nil {
		return err
	}
	pumpVal := 0
	if on {
		pumpVal = 1
	}
	_, err = conn.Exec(`UPDATE system_status SET pump_on = ?, updated_at = NOW(3) WHERE id = 1`, pumpVal)
	return err
}

func writeMode(mode string) error {
	conn, err := getDB()
	if err != nil {
		return err
	}
	_, err = conn.Exec(`UPDATE system_status SET mode = ?, updated_at = NOW(3) WHERE id = 1`, mode)
	return err
}

func insertHistory(action, mode string, moisture float64) error {
	conn, err := getDB()
	if err != nil {
		return err
	}
	_, err = conn.Exec(`
		INSERT INTO pump_history (action, mode, moisture, duration)
		VALUES (?, ?, ?, '—')
	`, action, mode, moisture)
	return err
}

func readHistory() ([]HistoryEntry, error) {
	conn, err := getDB()
	if err != nil {
		return nil, err
	}
	rows, err := conn.Query(`
		SELECT id, timestamp, action, mode, moisture, duration
		FROM pump_history
		ORDER BY id DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []HistoryEntry
	for rows.Next() {
		var e HistoryEntry
		var ts time.Time
		if err := rows.Scan(&e.ID, &ts, &e.Action, &e.Mode, &e.Moisture, &e.Duration); err != nil {
			return nil, err
		}
		e.Timestamp = ts.UTC().Format(time.RFC3339)
		list = append(list, e)
	}
	if list == nil {
		list = []HistoryEntry{}
	}
	return list, rows.Err()
}

// ─── Handler Entry-Point (Vercel Export) ───────────────────────────────────────

func Handler(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", chain(handleStatus))
	mux.HandleFunc("/api/pump/on", chain(handlePumpOn))
	mux.HandleFunc("/api/pump/off", chain(handlePumpOff))
	mux.HandleFunc("/api/mode", chain(handleMode))
	mux.HandleFunc("/api/history", chain(handleHistory))
	mux.HandleFunc("/api/history/clear", chain(handleClearHistory))
	mux.HandleFunc("/api/webhook", chain(handleWebhook))
	mux.HandleFunc("/api/ping", chain(handlePing))
	mux.ServeHTTP(w, r)
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	conn, err := getDB()
	if err != nil {
		dbError(w, err)
		return
	}
	if err := conn.Ping(); err != nil {
		dbError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "ts": time.Now().UTC().Format(time.RFC3339)})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store, must-revalidate")
	s, err := readStatus()
	if err != nil {
		dbError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func handlePumpOn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if err := writePump(true); err != nil {
		dbError(w, err)
		return
	}
	s, _ := readStatus()
	_ = insertHistory("ON", s.Mode, s.Moisture)
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true", "message": "Pompa dinyalakan"})
}

func handlePumpOff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if err := writePump(false); err != nil {
		dbError(w, err)
		return
	}
	s, _ := readStatus()
	_ = insertHistory("OFF", s.Mode, s.Moisture)
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true", "message": "Pompa dimatikan"})
}

func handleMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	var req ModeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.Mode != "auto" && req.Mode != "manual") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid mode, use 'auto' or 'manual'"})
		return
	}
	if err := writeMode(req.Mode); err != nil {
		dbError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true", "mode": req.Mode})
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	list, err := readHistory()
	if err != nil {
		dbError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func handleClearHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	conn, err := getDB()
	if err != nil {
		dbError(w, err)
		return
	}
	_, err = conn.Exec("DELETE FROM pump_history")
	if err != nil {
		dbError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true", "message": "Riwayat berhasil dibersihkan"})
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")

	type WebhookPayload struct {
		Moisture float64 `json:"moisture"`
		PumpOn   *bool   `json:"pump_on,omitempty"`
		Action   string  `json:"action,omitempty"`
		Mode     string  `json:"mode,omitempty"`
		ForceLog bool    `json:"force_log,omitempty"`
	}
	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
		return
	}

	conn, err := getDB()
	if err != nil {
		dbError(w, err)
		return
	}

	threshold := getThreshold()

	prev, err := readStatus()
	if err != nil {
		dbError(w, err)
		return
	}

	newMode := prev.Mode
	if payload.Mode == "auto" || payload.Mode == "manual" {
		newMode = payload.Mode
	}

	newPump := prev.PumpOn
	if newMode == "auto" && payload.PumpOn == nil {
		newPump = payload.Moisture < threshold
	} else if payload.PumpOn != nil {
		newPump = *payload.PumpOn
	}

	pumpVal := 0
	if newPump {
		pumpVal = 1
	}
	_, err = conn.Exec(`
		UPDATE system_status
		SET moisture = ?, threshold = ?, pump_on = ?, mode = ?, updated_at = NOW(3)
		WHERE id = 1
	`, payload.Moisture, threshold, pumpVal, newMode)
	if err != nil {
		dbError(w, err)
		return
	}

	action := strings.ToUpper(strings.TrimSpace(payload.Action))
	if action != "ON" && action != "OFF" {
		if newPump {
			action = "ON"
		} else {
			action = "OFF"
		}
	}

	if prev.PumpOn != newPump || payload.ForceLog {
		_ = insertHistory(action, newMode, payload.Moisture)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"moisture":  payload.Moisture,
		"pump_on":   newPump,
		"mode":      newMode,
		"threshold": threshold,
	})
}
