package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"plantcare-iot/pkg/config"
	"plantcare-iot/pkg/crypto"
	"plantcare-iot/pkg/mailer"
	"plantcare-iot/pkg/middleware"
	"plantcare-iot/pkg/model"
	"plantcare-iot/pkg/response"
)

// Handler exposes HTTP endpoints for authentication and session management.
type Handler struct {
	cfg  config.Config
	repo *Repository
}

// NewHandler instantiates an Auth Handler.
func NewHandler(cfg config.Config, repo *Repository) *Handler {
	return &Handler{cfg: cfg, repo: repo}
}

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

// Me returns the active authenticated user info extracted from Bearer token.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromRequest(r, h.cfg.JWTSecret)
	if claims == nil {
		response.JSON(w, http.StatusOK, map[string]any{"authenticated": false})
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"user": map[string]any{
			"id":    claims.UserID,
			"name":  claims.Name,
			"email": claims.Email,
		},
	})
}
