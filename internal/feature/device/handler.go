package device

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"plantcare-iot/internal/config"
	"plantcare-iot/internal/middleware"
	"plantcare-iot/internal/model"
	"plantcare-iot/internal/pkg/response"
)

// Handler manages HTTP endpoints for devices.
type Handler struct {
	cfg  config.Config
	repo *Repository
}

// NewHandler instantiates a Device Handler.
func NewHandler(cfg config.Config, repo *Repository) *Handler {
	return &Handler{cfg: cfg, repo: repo}
}

// List returns available devices for current session.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromRequest(r, h.cfg.JWTSecret)
	var uid *int
	if claims != nil {
		uid = &claims.UserID
	}

	list, err := h.repo.List(uid)
	if err != nil {
		response.DBError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, list)
}

// Claim pairs a device with an authenticated user's account using pairing code.
func (h *Handler) Claim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims := middleware.GetUserFromRequest(r, h.cfg.JWTSecret)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Silakan login terlebih dahulu untuk mendaftarkan unit.")
		return
	}

	var req model.ClaimReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DeviceID == "" || req.PairingCode == "" {
		response.Error(w, http.StatusBadRequest, "Device ID dan Pairing Code wajib diisi.")
		return
	}

	req.DeviceID = strings.TrimSpace(req.DeviceID)
	req.PairingCode = strings.TrimSpace(req.PairingCode)
	if req.Name == "" {
		req.Name = "Unit Tanaman " + req.DeviceID
	}

	currentUserID, pairingCode, err := h.repo.GetDeviceInfo(req.DeviceID)
	if err == sql.ErrNoRows {
		// New unlisted device - allow claim with provided pairing code
		if err := h.repo.CreateWithClaim(req.DeviceID, req.PairingCode, claims.UserID, req.Name); err != nil {
			response.DBError(w, err)
			return
		}
	} else if err == nil {
		if pairingCode != req.PairingCode {
			response.Error(w, http.StatusBadRequest, "Kode pairing salah!")
			return
		}
		if currentUserID.Valid && int(currentUserID.Int64) != claims.UserID {
			response.Error(w, http.StatusConflict, "Unit ini sudah didaftarkan oleh akun lain!")
			return
		}
		if err := h.repo.UpdateOwner(req.DeviceID, claims.UserID, req.Name); err != nil {
			response.DBError(w, err)
			return
		}
	} else {
		response.DBError(w, err)
		return
	}

	h.repo.EnsureStatusRow(req.DeviceID)

	response.JSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"message":   "Perangkat berhasil dihubungkan ke akun Anda!",
		"device_id": req.DeviceID,
		"name":      req.Name,
	})
}

