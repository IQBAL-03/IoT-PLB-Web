package api

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

var (
	mu     sync.RWMutex
	status = SystemStatus{
		PumpOn:    false,
		Moisture:  0,
		Mode:      "auto",
		Threshold: 50,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	history []HistoryEntry
	nextID  = 1
)

type gzipWriter struct {
	io.Writer
	http.ResponseWriter
}

func (g gzipWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func chain(h http.HandlerFunc) http.HandlerFunc {
	return cors(withGzip(h))
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func getThreshold() float64 {
	if envT := os.Getenv("MOISTURE_THRESHOLD"); envT != "" {
		if t, err := strconv.ParseFloat(envT, 64); err == nil {
			return t
		}
	}
	return 50.0
}

func Handler(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/status",   chain(handleStatus))
	mux.HandleFunc("/api/pump/on",  chain(handlePumpOn))
	mux.HandleFunc("/api/pump/off", chain(handlePumpOff))
	mux.HandleFunc("/api/mode",     chain(handleMode))
	mux.HandleFunc("/api/history",  chain(handleHistory))
	mux.HandleFunc("/api/webhook",  chain(handleWebhook))
	mux.HandleFunc("/api/ping",     chain(handlePing))
	mux.ServeHTTP(w, r)
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "ts": time.Now().Format(time.RFC3339)})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store, must-revalidate")
	mu.RLock()
	s := status
	mu.RUnlock()
	writeJSON(w, http.StatusOK, s)
}

func handlePumpOn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	mu.Lock()
	status.PumpOn = true
	status.Timestamp = time.Now().Format(time.RFC3339)
	entry := HistoryEntry{
		ID:        nextID,
		Timestamp: status.Timestamp,
		Action:    "ON",
		Mode:      status.Mode,
		Moisture:  status.Moisture,
		Duration:  "—",
	}
	nextID++
	history = append([]HistoryEntry{entry}, history...)
	if len(history) > 100 {
		history = history[:100]
	}
	mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true", "message": "Pompa dinyalakan"})
}

func handlePumpOff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	mu.Lock()
	status.PumpOn = false
	status.Timestamp = time.Now().Format(time.RFC3339)
	entry := HistoryEntry{
		ID:        nextID,
		Timestamp: status.Timestamp,
		Action:    "OFF",
		Mode:      status.Mode,
		Moisture:  status.Moisture,
		Duration:  "—",
	}
	nextID++
	history = append([]HistoryEntry{entry}, history...)
	if len(history) > 100 {
		history = history[:100]
	}
	mu.Unlock()
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
	mu.Lock()
	status.Mode = req.Mode
	status.Timestamp = time.Now().Format(time.RFC3339)
	mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true", "mode": req.Mode})
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	mu.RLock()
	h := make([]HistoryEntry, len(history))
	copy(h, history)
	mu.RUnlock()
	if h == nil {
		h = []HistoryEntry{}
	}
	writeJSON(w, http.StatusOK, h)
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
	}
	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
		return
	}

	threshold := getThreshold()

	mu.Lock()
	prevPump := status.PumpOn
	status.Moisture = payload.Moisture
	status.Threshold = threshold
	status.Timestamp = time.Now().Format(time.RFC3339)
	if status.Mode == "auto" {
		status.PumpOn = payload.Moisture < threshold
	} else if payload.PumpOn != nil {
		status.PumpOn = *payload.PumpOn
	}
	var newEntry *HistoryEntry
	if prevPump != status.PumpOn {
		action := "OFF"
		if status.PumpOn {
			action = "ON"
		}
		e := HistoryEntry{
			ID:        nextID,
			Timestamp: status.Timestamp,
			Action:    action,
			Mode:      status.Mode,
			Moisture:  status.Moisture,
			Duration:  "—",
		}
		nextID++
		history = append([]HistoryEntry{e}, history...)
		if len(history) > 100 {
			history = history[:100]
		}
		newEntry = &e
	}
	mu.Unlock()

	resp := map[string]any{
		"ok":        true,
		"moisture":  payload.Moisture,
		"pump_on":   status.PumpOn,
		"threshold": threshold,
	}
	if newEntry != nil {
		resp["history_entry"] = newEntry
	}
	writeJSON(w, http.StatusOK, resp)
}
