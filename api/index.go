package api

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	DeviceID  string  `json:"device_id"`
	PumpOn    bool    `json:"pump_on"`
	Moisture  float64 `json:"moisture"`
	Mode      string  `json:"mode"`
	Threshold float64 `json:"threshold"`
	Timestamp string  `json:"timestamp"`
}

type HistoryEntry struct {
	ID        int     `json:"id"`
	DeviceID  string  `json:"device_id"`
	Timestamp string  `json:"timestamp"`
	Action    string  `json:"action"`
	Mode      string  `json:"mode"`
	Moisture  float64 `json:"moisture"`
	Duration  string  `json:"duration"`
}

type DeviceItem struct {
	DeviceID    string `json:"device_id"`
	Name        string `json:"name"`
	IsOwner     bool   `json:"is_owner"`
	HasOwner    bool   `json:"has_owner"`
	PairingCode string `json:"pairing_code,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type UserClaims struct {
	UserID int    `json:"uid"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Exp    int64  `json:"exp"`
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
	return err
}

// ─── JWT & Auth Utilities ─────────────────────────────────────────────────────

func getJWTSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "plantcare-secret-key-2026-auth-token-tidb"
	}
	return []byte(s)
}

func hashPassword(pass string) string {
	salt := "plb_salt_2026"
	h := sha256.New()
	h.Write([]byte(salt + pass))
	return hex.EncodeToString(h.Sum(nil))
}

func createJWT(uid int, email, name string) (string, error) {
	exp := time.Now().Add(30 * 24 * time.Hour).Unix()
	claims := UserClaims{
		UserID: uid,
		Email:  email,
		Name:   name,
		Exp:    exp,
	}
	headerJSON := `{"alg":"HS256","typ":"JWT"}`
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerEnc := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))
	claimsEnc := base64.RawURLEncoding.EncodeToString(claimsJSON)
	payload := headerEnc + "." + claimsEnc

	mac := hmac.New(sha256.New, getJWTSecret())
	mac.Write([]byte(payload))
	sigEnc := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return payload + "." + sigEnc, nil
}

func verifyJWT(tokenStr string) (*UserClaims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	payload := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, getJWTSecret())
	mac.Write([]byte(payload))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if parts[2] != expectedSig {
		return nil, errors.New("invalid token signature")
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims UserClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, err
	}

	if time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expired")
	}
	return &claims, nil
}

func getUserFromRequest(r *http.Request) *UserClaims {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := verifyJWT(token)
		if err == nil {
			return claims
		}
	}
	return nil
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

func dbError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database unavailable", "detail": err.Error()})
}

func getThreshold() float64 {
	if envT := os.Getenv("MOISTURE_THRESHOLD"); envT != "" {
		if t, err := strconv.ParseFloat(envT, 64); err == nil {
			return t
		}
	}
	return 50.0
}

func resolveDeviceID(r *http.Request) string {
	dev := r.URL.Query().Get("device_id")
	if dev == "" {
		dev = "DEV-PLB-001"
	}
	return strings.TrimSpace(dev)
}

// ─── DB Operations ────────────────────────────────────────────────────────────

func readStatus(deviceID string) (SystemStatus, error) {
	conn, err := getDB()
	if err != nil {
		return SystemStatus{}, err
	}
	var s SystemStatus
	var pumpOn int
	var updatedAt time.Time
	err = conn.QueryRow(`
		SELECT device_id, pump_on, moisture, mode, threshold, updated_at
		FROM system_status WHERE device_id = ?
		LIMIT 1
	`, deviceID).Scan(&s.DeviceID, &pumpOn, &s.Moisture, &s.Mode, &s.Threshold, &updatedAt)
	if err == sql.ErrNoRows {
		// Initialize if missing
		_, _ = conn.Exec(`
			INSERT IGNORE INTO system_status (device_id, pump_on, moisture, mode, threshold)
			VALUES (?, 0, 0, 'auto', 50)
		`, deviceID)
		return SystemStatus{DeviceID: deviceID, PumpOn: false, Moisture: 0, Mode: "auto", Threshold: 50, Timestamp: time.Now().UTC().Format(time.RFC3339)}, nil
	}
	if err != nil {
		return s, err
	}
	s.PumpOn = pumpOn == 1
	s.Timestamp = updatedAt.UTC().Format(time.RFC3339)
	return s, nil
}

func writePump(deviceID string, on bool) error {
	conn, err := getDB()
	if err != nil {
		return err
	}
	pumpVal := 0
	if on {
		pumpVal = 1
	}
	_, err = conn.Exec(`
		INSERT INTO system_status (device_id, pump_on, updated_at)
		VALUES (?, ?, NOW(3))
		ON DUPLICATE KEY UPDATE pump_on = VALUES(pump_on), updated_at = NOW(3)
	`, deviceID, pumpVal)
	return err
}

func writeMode(deviceID, mode string) error {
	conn, err := getDB()
	if err != nil {
		return err
	}
	_, err = conn.Exec(`
		INSERT INTO system_status (device_id, mode, updated_at)
		VALUES (?, ?, NOW(3))
		ON DUPLICATE KEY UPDATE mode = VALUES(mode), updated_at = NOW(3)
	`, deviceID, mode)
	return err
}

func insertHistory(deviceID, action, mode string, moisture float64) error {
	conn, err := getDB()
	if err != nil {
		return err
	}
	_, err = conn.Exec(`
		INSERT INTO pump_history (device_id, action, mode, moisture, duration)
		VALUES (?, ?, ?, ?, '—')
	`, deviceID, action, mode, moisture)
	return err
}

func readHistory(deviceID string) ([]HistoryEntry, error) {
	conn, err := getDB()
	if err != nil {
		return nil, err
	}
	rows, err := conn.Query(`
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

	var list []HistoryEntry
	for rows.Next() {
		var e HistoryEntry
		var ts time.Time
		if err := rows.Scan(&e.ID, &e.DeviceID, &ts, &e.Action, &e.Mode, &e.Moisture, &e.Duration); err != nil {
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

	// Auth Endpoints
	mux.HandleFunc("/api/auth/register", chain(handleAuthRegister))
	mux.HandleFunc("/api/auth/login",    chain(handleAuthLogin))
	mux.HandleFunc("/api/auth/google",   chain(handleAuthGoogle))
	mux.HandleFunc("/api/auth/me",       chain(handleAuthMe))

	// Device Endpoints
	mux.HandleFunc("/api/devices",       chain(handleDevicesList))
	mux.HandleFunc("/api/devices/claim", chain(handleDeviceClaim))

	// Control & Status Endpoints
	mux.HandleFunc("/api/status",        chain(handleStatus))
	mux.HandleFunc("/api/pump/on",       chain(handlePumpOn))
	mux.HandleFunc("/api/pump/off",      chain(handlePumpOff))
	mux.HandleFunc("/api/mode",          chain(handleMode))
	mux.HandleFunc("/api/history",       chain(handleHistory))
	mux.HandleFunc("/api/history/clear", chain(handleClearHistory))
	mux.HandleFunc("/api/webhook",       chain(handleWebhook))
	mux.HandleFunc("/api/ping",          chain(handlePing))

	mux.ServeHTTP(w, r)
}

// ─── Auth Handlers ────────────────────────────────────────────────────────────

func handleAuthRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	type RegisterReq struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req RegisterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || len(req.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Nama, email valid, dan password minimal 6 karakter dibutuhkan."})
		return
	}

	conn, err := getDB()
	if err != nil {
		dbError(w, err)
		return
	}

	passHash := hashPassword(req.Password)
	res, err := conn.Exec(`
		INSERT INTO users (name, email, password_hash)
		VALUES (?, ?, ?)
	`, req.Name, strings.ToLower(strings.TrimSpace(req.Email)), passHash)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Email sudah terdaftar. Silakan login."})
		return
	}

	uid, _ := res.LastInsertId()
	token, _ := createJWT(int(uid), req.Email, req.Name)

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"token": token,
		"user": map[string]any{
			"id":    uid,
			"name":  req.Name,
			"email": req.Email,
		},
	})
}

func handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
		return
	}

	conn, err := getDB()
	if err != nil {
		dbError(w, err)
		return
	}

	var uid int
	var name, storedHash string
	err = conn.QueryRow(`
		SELECT id, name, password_hash FROM users WHERE email = ? LIMIT 1
	`, strings.ToLower(strings.TrimSpace(req.Email))).Scan(&uid, &name, &storedHash)
	if err != nil || hashPassword(req.Password) != storedHash {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Email atau password salah."})
		return
	}

	token, _ := createJWT(uid, req.Email, name)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"token": token,
		"user": map[string]any{
			"id":    uid,
			"name":  name,
			"email": req.Email,
		},
	})
}

func handleAuthGoogle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	type GoogleAuthReq struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	var req GoogleAuthReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Email akun Google tidak valid."})
		return
	}

	conn, err := getDB()
	if err != nil {
		dbError(w, err)
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)
	if name == "" {
		parts := strings.Split(email, "@")
		name = parts[0]
	}

	var uid int
	err = conn.QueryRow("SELECT id FROM users WHERE email = ? LIMIT 1", email).Scan(&uid)
	if err == sql.ErrNoRows {
		res, err := conn.Exec(`
			INSERT INTO users (name, email, password_hash)
			VALUES (?, ?, ?)
		`, name, email, hashPassword("google_oauth_authorized"))
		if err != nil {
			dbError(w, err)
			return
		}
		newID, _ := res.LastInsertId()
		uid = int(newID)
	} else if err != nil {
		dbError(w, err)
		return
	}

	token, _ := createJWT(uid, email, name)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"token": token,
		"user": map[string]any{
			"id":    uid,
			"name":  name,
			"email": email,
		},
	})
}

func handleAuthMe(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromRequest(r)
	if claims == nil {
		writeJSON(w, http.StatusOK, map[string]any{"authenticated": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"user": map[string]any{
			"id":    claims.UserID,
			"name":  claims.Name,
			"email": claims.Email,
		},
	})
}

// ─── Device Handlers ──────────────────────────────────────────────────────────

func handleDevicesList(w http.ResponseWriter, r *http.Request) {
	conn, err := getDB()
	if err != nil {
		dbError(w, err)
		return
	}

	claims := getUserFromRequest(r)
	var rows *sql.Rows
	if claims != nil {
		rows, err = conn.Query(`
			SELECT device_id, name, (user_id IS NOT NULL) as has_owner, (user_id = ?) as is_owner, created_at
			FROM devices
			WHERE user_id = ? OR user_id IS NULL
			ORDER BY is_owner DESC, device_id ASC
		`, claims.UserID, claims.UserID)
	} else {
		rows, err = conn.Query(`
			SELECT device_id, name, (user_id IS NOT NULL) as has_owner, 0 as is_owner, created_at
			FROM devices
			ORDER BY device_id ASC
		`)
	}
	if err != nil {
		dbError(w, err)
		return
	}
	defer rows.Close()

	var list []DeviceItem
	for rows.Next() {
		var d DeviceItem
		var hasOwner, isOwner int
		var ts time.Time
		if err := rows.Scan(&d.DeviceID, &d.Name, &hasOwner, &isOwner, &ts); err == nil {
			d.HasOwner = hasOwner == 1
			d.IsOwner = isOwner == 1
			d.CreatedAt = ts.UTC().Format(time.RFC3339)
			list = append(list, d)
		}
	}
	if len(list) == 0 {
		list = []DeviceItem{
			{DeviceID: "DEV-PLB-001", Name: "Unit Tanaman Utama", IsOwner: true, HasOwner: false},
		}
	}
	writeJSON(w, http.StatusOK, list)
}

func handleDeviceClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	claims := getUserFromRequest(r)
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Silakan login terlebih dahulu untuk mendaftarkan unit."})
		return
	}

	type ClaimReq struct {
		DeviceID    string `json:"device_id"`
		PairingCode string `json:"pairing_code"`
		Name        string `json:"name"`
	}
	var req ClaimReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DeviceID == "" || req.PairingCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Device ID dan Pairing Code wajib diisi."})
		return
	}

	req.DeviceID = strings.TrimSpace(req.DeviceID)
	req.PairingCode = strings.TrimSpace(req.PairingCode)
	if req.Name == "" {
		req.Name = "Unit Tanaman " + req.DeviceID
	}

	conn, err := getDB()
	if err != nil {
		dbError(w, err)
		return
	}

	// Upsert or Claim
	var currentUserID sql.NullInt64
	var pairingCode string
	err = conn.QueryRow(`
		SELECT user_id, pairing_code FROM devices WHERE device_id = ?
	`, req.DeviceID).Scan(&currentUserID, &pairingCode)

	if err == sql.ErrNoRows {
		// New unlisted device - allow claim if pairing code provided
		_, err = conn.Exec(`
			INSERT INTO devices (device_id, pairing_code, user_id, name)
			VALUES (?, ?, ?, ?)
		`, req.DeviceID, req.PairingCode, claims.UserID, req.Name)
	} else if err == nil {
		if pairingCode != req.PairingCode {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Kode pairing salah!"})
			return
		}
		if currentUserID.Valid && int(currentUserID.Int64) != claims.UserID {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "Unit ini sudah didaftarkan oleh akun lain!"})
			return
		}
		_, err = conn.Exec(`
			UPDATE devices SET user_id = ?, name = ? WHERE device_id = ?
		`, claims.UserID, req.Name, req.DeviceID)
	}

	if err != nil {
		dbError(w, err)
		return
	}

	// Initialize status row
	_, _ = conn.Exec(`
		INSERT IGNORE INTO system_status (device_id, pump_on, moisture, mode, threshold)
		VALUES (?, 0, 0, 'auto', 50)
	`, req.DeviceID)

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"message":   "Perangkat berhasil dihubungkan ke akun Anda!",
		"device_id": req.DeviceID,
		"name":      req.Name,
	})
}

// ─── Status & Control Handlers ────────────────────────────────────────────────

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
	devID := resolveDeviceID(r)
	s, err := readStatus(devID)
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
	type PumpReq struct {
		DeviceID string `json:"device_id"`
	}
	var req PumpReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	devID := strings.TrimSpace(req.DeviceID)
	if devID == "" {
		devID = resolveDeviceID(r)
	}

	if err := writePump(devID, true); err != nil {
		dbError(w, err)
		return
	}
	s, _ := readStatus(devID)
	_ = insertHistory(devID, "ON", s.Mode, s.Moisture)
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true", "message": "Pompa dinyalakan", "device_id": devID})
}

func handlePumpOff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	type PumpReq struct {
		DeviceID string `json:"device_id"`
	}
	var req PumpReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	devID := strings.TrimSpace(req.DeviceID)
	if devID == "" {
		devID = resolveDeviceID(r)
	}

	if err := writePump(devID, false); err != nil {
		dbError(w, err)
		return
	}
	s, _ := readStatus(devID)
	_ = insertHistory(devID, "OFF", s.Mode, s.Moisture)
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true", "message": "Pompa dimatikan", "device_id": devID})
}

func handleMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	type ModeReq struct {
		DeviceID string `json:"device_id"`
		Mode     string `json:"mode"`
	}
	var req ModeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.Mode != "auto" && req.Mode != "manual") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid mode, use 'auto' or 'manual'"})
		return
	}
	devID := strings.TrimSpace(req.DeviceID)
	if devID == "" {
		devID = resolveDeviceID(r)
	}

	if err := writeMode(devID, req.Mode); err != nil {
		dbError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true", "mode": req.Mode, "device_id": devID})
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	devID := resolveDeviceID(r)
	list, err := readHistory(devID)
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
	devID := resolveDeviceID(r)
	conn, err := getDB()
	if err != nil {
		dbError(w, err)
		return
	}
	_, err = conn.Exec("DELETE FROM pump_history WHERE device_id = ?", devID)
	if err != nil {
		dbError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true", "message": "Riwayat berhasil dibersihkan", "device_id": devID})
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")

	type WebhookPayload struct {
		DeviceID string  `json:"device_id"`
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

	devID := strings.TrimSpace(payload.DeviceID)
	if devID == "" {
		devID = resolveDeviceID(r)
	}

	conn, err := getDB()
	if err != nil {
		dbError(w, err)
		return
	}

	threshold := getThreshold()

	prev, err := readStatus(devID)
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
		INSERT INTO system_status (device_id, moisture, threshold, pump_on, mode, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(3))
		ON DUPLICATE KEY UPDATE
			moisture = VALUES(moisture),
			threshold = VALUES(threshold),
			pump_on = VALUES(pump_on),
			mode = VALUES(mode),
			updated_at = NOW(3)
	`, devID, payload.Moisture, threshold, pumpVal, newMode)
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
		_ = insertHistory(devID, action, newMode, payload.Moisture)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"device_id": devID,
		"moisture":  payload.Moisture,
		"pump_on":   newPump,
		"mode":      newMode,
		"threshold": threshold,
	})
}
