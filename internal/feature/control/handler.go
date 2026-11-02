package control

import (
	"encoding/json"
	"net/http"
	"strings"

	"plantcare-iot/internal/config"
	"plantcare-iot/internal/feature/history"
	"plantcare-iot/internal/model"
	"plantcare-iot/internal/pkg/response"
)

// Handler manages HTTP endpoints for actuators and status queries.
type Handler struct {
	cfg         config.Config
	repo        *Repository
	historyRepo *history.Repository
}

// NewHandler instantiates a Control Handler.
func NewHandler(cfg config.Config, repo *Repository, historyRepo *history.Repository) *Handler {
	return &Handler{
		cfg:         cfg,
		repo:        repo,
		historyRepo: historyRepo,
	}
}

// Status returns current sensor readings and pump status.
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Cache-Control", "no-store, must-revalidate")
	devID := response.ResolveDeviceID(r, h.cfg.DefaultDeviceID)

	s, err := h.repo.ReadStatus(devID)
	if err != nil {
		response.DBError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, s)
}

// PumpOn turns the pump on for the specified device and logs the action.
func (h *Handler) PumpOn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Cache-Control", "no-store")

	var req model.PumpReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	devID := strings.TrimSpace(req.DeviceID)
	if devID == "" {
		devID = response.ResolveDeviceID(r, h.cfg.DefaultDeviceID)
	}

	if err := h.repo.WritePump(devID, true); err != nil {
		response.DBError(w, err)
		return
	}

	s, _ := h.repo.ReadStatus(devID)
	_ = h.historyRepo.Insert(devID, "ON", s.Mode, s.Moisture)

	response.JSON(w, http.StatusOK, map[string]string{
		"ok":        "true",
		"message":   "Pompa dinyalakan",
		"device_id": devID,
	})
}

// PumpOff turns the pump off for the specified device and logs the action.
func (h *Handler) PumpOff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Cache-Control", "no-store")

	var req model.PumpReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	devID := strings.TrimSpace(req.DeviceID)
	if devID == "" {
		devID = response.ResolveDeviceID(r, h.cfg.DefaultDeviceID)
	}

	if err := h.repo.WritePump(devID, false); err != nil {
		response.DBError(w, err)
		return
	}

	s, _ := h.repo.ReadStatus(devID)
	_ = h.historyRepo.Insert(devID, "OFF", s.Mode, s.Moisture)

	response.JSON(w, http.StatusOK, map[string]string{
		"ok":        "true",
		"message":   "Pompa dimatikan",
		"device_id": devID,
	})
}

// Mode switches operation mode between 'auto' and 'manual'.
func (h *Handler) Mode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Cache-Control", "no-store")

	var req model.ModeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.Mode != "auto" && req.Mode != "manual") {
		response.Error(w, http.StatusBadRequest, "invalid mode, use 'auto' or 'manual'")
		return
	}

	devID := strings.TrimSpace(req.DeviceID)
	if devID == "" {
		devID = response.ResolveDeviceID(r, h.cfg.DefaultDeviceID)
	}

	if err := h.repo.WriteMode(devID, req.Mode); err != nil {
		response.DBError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"ok":        "true",
		"mode":      req.Mode,
		"device_id": devID,
	})
}

