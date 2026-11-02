package history

import (
	"net/http"

	"plantcare-iot/internal/config"
	"plantcare-iot/internal/pkg/response"
)

// Handler manages HTTP endpoints for history.
type Handler struct {
	cfg  config.Config
	repo *Repository
}

// NewHandler instantiates a History Handler.
func NewHandler(cfg config.Config, repo *Repository) *Handler {
	return &Handler{cfg: cfg, repo: repo}
}

// List returns historical log records for the selected device.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	devID := response.ResolveDeviceID(r, h.cfg.DefaultDeviceID)

	list, err := h.repo.List(devID)
	if err != nil {
		response.DBError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, list)
}

// Clear clears the history entries for the selected device.
func (h *Handler) Clear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	devID := response.ResolveDeviceID(r, h.cfg.DefaultDeviceID)

	if err := h.repo.Clear(devID); err != nil {
		response.DBError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{
		"ok":        "true",
		"message":   "Riwayat berhasil dibersihkan",
		"device_id": devID,
	})
}

