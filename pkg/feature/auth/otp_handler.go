package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"plantcare-iot/pkg/crypto"
	"plantcare-iot/pkg/mailer"
	"plantcare-iot/pkg/model"
	"plantcare-iot/pkg/response"
)

// VerifyOTP validates the OTP code, consumes it, and finalizes registration or trusts the device.
func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.VerifyOTPReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Permintaan verifikasi tidak valid.")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	code := strings.TrimSpace(req.Code)
	purpose := strings.TrimSpace(req.Purpose)

	if email == "" || code == "" || purpose == "" {
		response.Error(w, http.StatusBadRequest, "Email, kode OTP, dan jenis verifikasi wajib diisi.")
		return
	}

	rec, err := h.repo.GetActiveOTP(email, purpose)
	if err != nil {
		response.DBError(w, err)
		return
	}
	if rec == nil {
		response.Error(w, http.StatusBadRequest, "Kode OTP salah, kedaluwarsa, atau sudah digunakan. Silakan minta kode baru.")
		return
	}

	// Attempt limiting: Max 5 tries
	if rec.Attempts >= 5 {
		_ = h.repo.MarkOTPUsed(rec.ID)
		response.Error(w, http.StatusBadRequest, "Terlalu banyak percobaan salah. Kode OTP dibatalkan demi keamanan. Silakan minta kode baru.")
		return
	}

	// Verify hashed OTP
	if crypto.HashOTP(code) != rec.CodeHash {
		attempts, _ := h.repo.IncrementOTPAttempts(rec.ID)
		rem := 5 - attempts
		if rem <= 0 {
			_ = h.repo.MarkOTPUsed(rec.ID)
			response.Error(w, http.StatusBadRequest, "Batas percobaan tercapai. Kode OTP dibatalkan.")
			return
		}
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("Kode OTP salah. Sisa percobaan: %d kali.", rem))
		return
	}

	// Mark OTP as consumed
	_ = h.repo.MarkOTPUsed(rec.ID)

	// Decode metadata stored with the OTP
	var meta map[string]any
	_ = json.Unmarshal([]byte(rec.Metadata), &meta)

	devToken := req.DeviceToken
	devName := req.DeviceName
	if devToken == "" && meta != nil {
		if t, ok := meta["device_token"].(string); ok {
			devToken = t
		}
		if n, ok := meta["device_name"].(string); ok {
			devName = n
		}
	}

	switch purpose {
	case "registration", "google_registration":
		name := ""
		passHash := ""
		provider := "local"
		if meta != nil {
			if n, ok := meta["name"].(string); ok {
				name = n
			}
			if p, ok := meta["password_hash"].(string); ok {
				passHash = p
			}
			if pr, ok := meta["provider"].(string); ok {
				provider = pr
			}
		}

		if passHash == "" && provider == "google" {
			passHash = crypto.HashPassword("google_verified_auth")
		}

		// Check if user already exists
		existing, err := h.repo.GetUserByEmail(email)
		if err != nil {
			response.DBError(w, err)
			return
		}

		var uid int
		if existing != nil {
			uid = existing.ID
			_ = h.repo.VerifyUser(uid)
		} else {
			newID, err := h.repo.CreateUser(name, email, passHash, provider, true)
			if err != nil {
				response.DBError(w, err)
				return
			}
			uid = newID
		}

		// Trust current device
		_ = h.repo.TrustDevice(uid, devToken, devName)

		token, _ := crypto.CreateJWT(uid, email, name, h.cfg.JWTSecret)
		response.JSON(w, http.StatusOK, map[string]any{
			"ok":      true,
			"message": "Pendaftaran berhasil diverifikasi! Selamat datang di PlantCare IoT.",
			"token":   token,
			"user": map[string]any{
				"id":    uid,
				"name":  name,
				"email": email,
			},
		})

	case "login_new_device":
		user, err := h.repo.GetUserByEmail(email)
		if err != nil || user == nil {
			response.Error(w, http.StatusNotFound, "Pengguna tidak ditemukan.")
			return
		}

		// Mark new device as trusted
		_ = h.repo.TrustDevice(user.ID, devToken, devName)

		token, _ := crypto.CreateJWT(user.ID, user.Email, user.Name, h.cfg.JWTSecret)
		response.JSON(w, http.StatusOK, map[string]any{
			"ok":      true,
			"message": "Perangkat berhasil diverifikasi dan ditandai sebagai perangkat terpercaya.",
			"token":   token,
			"user": map[string]any{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
			},
		})

	default:
		response.Error(w, http.StatusBadRequest, "Tujuan verifikasi tidak dikenal.")
	}
}

// ResendOTP delivers a new OTP code if the cooldown period has elapsed.
func (h *Handler) ResendOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.ResendOTPReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Email) == "" {
		response.Error(w, http.StatusBadRequest, "Email wajib diisi.")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	purpose := strings.TrimSpace(req.Purpose)
	if purpose == "" {
		purpose = "registration"
	}

	allowed, waitSec, err := h.repo.CheckOTPRateLimit(email)
	if err != nil {
		response.DBError(w, err)
		return
	}
	if !allowed {
		response.Error(w, http.StatusTooManyRequests, fmt.Sprintf("Mohon tunggu %d detik sebelum meminta kode baru.", waitSec))
		return
	}

	// Carry over existing metadata if available
	metaStr := ""
	oldRec, _ := h.repo.GetActiveOTP(email, purpose)
	if oldRec != nil {
		metaStr = oldRec.Metadata
	}

	otpCode := mailer.GenerateOTP()
	otpHash := crypto.HashOTP(otpCode)
	expiresAt := time.Now().Add(5 * time.Minute).UTC()

	if err := h.repo.SaveOTP(email, otpHash, purpose, metaStr, expiresAt); err != nil {
		response.DBError(w, err)
		return
	}

	_ = mailer.SendOTP(h.cfg, email, otpCode)

	response.JSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "Kode OTP baru telah berhasil dikirim ke email Anda.",
		"email":   email,
	})
}

