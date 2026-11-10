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

// Login handles password authentication and checks for unrecognized/new devices.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid.")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "Email dan password wajib diisi.")
		return
	}

	user, err := h.repo.GetUserByEmail(email)
	if err != nil {
		response.DBError(w, err)
		return
	}

	// 1. If email is not registered in database, reject and instruct user to register
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      false,
			"error":   "user_not_found",
			"message": "Email belum terdaftar di sistem. Silakan lakukan pendaftaran akun terlebih dahulu.",
			"email":   email,
		})
		return
	}

	// 2. Unverified accounts cannot login until OTP verification is completed
	if !user.IsVerified {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      false,
			"error":   "unverified_account",
			"message": "Akun Anda belum menyelesaikan verifikasi OTP. Silakan selesaikan pendaftaran Anda.",
			"email":   email,
		})
		return
	}

	// 3. Validate password
	if crypto.HashPassword(req.Password) != user.PasswordHash {
		response.Error(w, http.StatusUnauthorized, "Email atau password yang Anda masukkan salah.")
		return
	}

	// 4. Check whether this device is already trusted
	isTrusted, err := h.repo.IsDeviceTrusted(user.ID, req.DeviceToken)
	if err != nil {
		response.DBError(w, err)
		return
	}

	// If device is not recognized, require OTP verification before granting access
	if !isTrusted {
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
			"user_id":      user.ID,
			"device_token": req.DeviceToken,
			"device_name":  req.DeviceName,
		})

		expiresAt := time.Now().Add(5 * time.Minute).UTC()
		if err := h.repo.SaveOTP(email, otpHash, "login_new_device", string(metaJSON), expiresAt); err != nil {
			response.DBError(w, err)
			return
		}

		_ = mailer.SendOTP(h.cfg, email, otpCode)

		response.JSON(w, http.StatusOK, map[string]any{
			"ok":          true,
			"require_otp": true,
			"purpose":     "login_new_device",
			"email":       email,
			"message":     "Perangkat baru terdeteksi. Kode verifikasi OTP telah dikirimkan ke email Anda.",
		})
		return
	}

	// Trusted device: Login immediately and issue JWT
	token, _ := crypto.CreateJWT(user.ID, user.Email, user.Name, h.cfg.JWTSecret)
	response.JSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "Login berhasil.",
		"token":   token,
		"user": map[string]any{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

// Google handles Google OAuth login and registration dispatching.
func (h *Handler) Google(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.GoogleAuthReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Email) == "" {
		response.Error(w, http.StatusBadRequest, "Data profil Google tidak valid.")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = strings.Split(email, "@")[0]
	}

	user, err := h.repo.GetUserByEmail(email)
	if err != nil {
		response.DBError(w, err)
		return
	}

	// If action is explicit registration from Google
	if req.Action == "register" {
		if user != nil && user.IsVerified {
			response.Error(w, http.StatusConflict, "Akun dengan email ini sudah terdaftar. Silakan login.")
			return
		}

		// Dispatch registration OTP
		h.dispatchRegistrationOTP(w, email, name, "", "google", req.DeviceToken, req.DeviceName)
		return
	}

	// Normal Login via Google
	// Rule: If Google account is not registered in database, do NOT auto-create account. Direct user to register.
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":          false,
			"error":       "user_not_found",
			"message":     "Akun Google ini belum terdaftar di sistem. Silakan selesaikan proses registrasi.",
			"google_user": map[string]string{"email": email, "name": name},
		})
		return
	}

	if !user.IsVerified {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      false,
			"error":   "unverified_account",
			"message": "Akun Anda belum menyelesaikan verifikasi OTP. Silakan selesaikan pendaftaran Anda.",
			"email":   email,
		})
		return
	}

	// Check trusted device for Google login as well
	isTrusted, err := h.repo.IsDeviceTrusted(user.ID, req.DeviceToken)
	if err != nil {
		response.DBError(w, err)
		return
	}

	if !isTrusted {
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
			"user_id":      user.ID,
			"device_token": req.DeviceToken,
			"device_name":  req.DeviceName,
		})

		expiresAt := time.Now().Add(5 * time.Minute).UTC()
		if err := h.repo.SaveOTP(email, otpHash, "login_new_device", string(metaJSON), expiresAt); err != nil {
			response.DBError(w, err)
			return
		}

		_ = mailer.SendOTP(h.cfg, email, otpCode)

		response.JSON(w, http.StatusOK, map[string]any{
			"ok":          true,
			"require_otp": true,
			"purpose":     "login_new_device",
			"email":       email,
			"message":     "Perangkat baru terdeteksi. Kode verifikasi telah dikirim ke email Anda.",
		})
		return
	}

	// Trusted device Google login
	token, _ := crypto.CreateJWT(user.ID, user.Email, user.Name, h.cfg.JWTSecret)
	response.JSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "Login Google berhasil.",
		"token":   token,
		"user": map[string]any{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}
