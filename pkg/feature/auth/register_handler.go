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

// Register initiates account creation and requires OTP verification before activation.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.RegisterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid.")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)
	if name == "" || email == "" || len(req.Password) < 6 {
		response.Error(w, http.StatusBadRequest, "Nama lengkap, email valid, dan password minimal 6 karakter dibutuhkan.")
		return
	}

	// Check if already registered and verified
	existing, err := h.repo.GetUserByEmail(email)
	if err != nil {
		response.DBError(w, err)
		return
	}
	if existing != nil && existing.IsVerified {
		response.Error(w, http.StatusConflict, "Email sudah terdaftar. Silakan masuk menggunakan kata sandi Anda.")
		return
	}

	passHash := crypto.HashPassword(req.Password)
	h.dispatchRegistrationOTP(w, email, name, passHash, "local", req.DeviceToken, req.DeviceName)
}

// dispatchRegistrationOTP generates, stores, and emails an OTP for account registration.
func (h *Handler) dispatchRegistrationOTP(w http.ResponseWriter, email, name, passHash, provider, devToken, devName string) {
	allowed, waitSec, err := h.repo.CheckOTPRateLimit(email)
	if err != nil {
		response.DBError(w, err)
		return
	}
	if !allowed {
		response.Error(w, http.StatusTooManyRequests, fmt.Sprintf("Mohon tunggu %d detik sebelum meminta kode verifikasi baru.", waitSec))
		return
	}

	otpCode := mailer.GenerateOTP()
	otpHash := crypto.HashOTP(otpCode)
	metaJSON, _ := json.Marshal(map[string]any{
		"name":          name,
		"password_hash": passHash,
		"provider":      provider,
		"device_token":  devToken,
		"device_name":   devName,
	})

	expiresAt := time.Now().Add(5 * time.Minute).UTC()
	if err := h.repo.SaveOTP(email, otpHash, "registration", string(metaJSON), expiresAt); err != nil {
		response.DBError(w, err)
		return
	}

	_ = mailer.SendOTP(h.cfg, email, otpCode)

	response.JSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"require_otp": true,
		"purpose":     "registration",
		"email":       email,
		"message":     "Kode verifikasi telah dikirim ke email Anda. Silakan masukkan kode OTP untuk menyelesaikan pendaftaran.",
	})
}

// RegisterRequestOTP is an alias for Register to maintain backwards compatibility.
func (h *Handler) RegisterRequestOTP(w http.ResponseWriter, r *http.Request) {
	h.Register(w, r)
}

// RegisterVerifyOTP is an alias for VerifyOTP to maintain backwards compatibility.
func (h *Handler) RegisterVerifyOTP(w http.ResponseWriter, r *http.Request) {
	h.VerifyOTP(w, r)
}

