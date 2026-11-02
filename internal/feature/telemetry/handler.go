// Package telemetry handles real-time IoT data ingestion via webhook and connectivity health checks.
package telemetry

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"plantcare-iot/internal/config"
	"plantcare-iot/internal/feature/control"
	"plantcare-iot/internal/feature/history"
	"plantcare-iot/internal/model"
	"plantcare-iot/internal/pkg/response"
)

// Handler manages webhook inputs from hardware devices and connectivity ping.
type Handler struct {
	db          *sql.DB
	cfg         config.Config
	controlRepo *control.Repository
	historyRepo *history.Repository
}

// NewHandler instantiates a Telemetry Handler.
func NewHandler(
	db *sql.DB,
	cfg config.Config,
	controlRepo *control.Repository,
	historyRepo *history.Repository,
) *Handler {
	return &Handler{
		db:          db,
		cfg:         cfg,
		controlRepo: controlRepo,
		historyRepo: historyRepo,
	}
}

// Ping verifies database connectivity and provides server timestamp.
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if err := h.db.Ping(); err != nil {
		response.DBError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"ts":     time.Now().UTC().Format(time.RFC3339),
	})
}

// Webhook processes sensor telemetry sent by ESP32 or Blynk Webhooks.
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Cache-Control", "no-store")

	var payload model.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	devID := strings.TrimSpace(payload.DeviceID)
	if devID == "" {
		devID = response.ResolveDeviceID(r, h.cfg.DefaultDeviceID)
	}

	threshold := h.cfg.MoistureThreshold

	prev, err := h.controlRepo.ReadStatus(devID)
	if err != nil {
		response.DBError(w, err)
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

	_, err = h.db.Exec(`
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
		response.DBError(w, err)
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
		_ = h.historyRepo.Insert(devID, action, newMode, payload.Moisture)
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"device_id": devID,
		"moisture":  payload.Moisture,
		"pump_on":   newPump,
		"mode":      newMode,
		"threshold": threshold,
	})
}

