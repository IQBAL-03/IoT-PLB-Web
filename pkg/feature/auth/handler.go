package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"plantcare-iot/pkg/config"
	"plantcare-iot/pkg/middleware"
	"plantcare-iot/pkg/model"
	"plantcare-iot/pkg/crypto"
	"plantcare-iot/pkg/response"
)

// Handler exposes HTTP handlers for authentication operations.
type Handler struct {
	cfg  config.Config
	repo *Repository
}

// NewHandler instantiates an Auth Handler.
func NewHandler(cfg config.Config, repo *Repository) *Handler {
	return &Handler{cfg: cfg, repo: repo}
}

// Register processes new user account creation.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.RegisterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || len(req.Password) < 6 {
		response.Error(w, http.StatusBadRequest, "Nama, email valid, dan password minimal 6 karakter dibutuhkan.")
		return
	}

	passHash := crypto.HashPassword(req.Password)
	uid, err := h.repo.CreateUser(req.Name, req.Email, passHash)
	if err != nil {
		response.Error(w, http.StatusConflict, "Email sudah terdaftar. Silakan login.")
		return
	}

	token, _ := crypto.CreateJWT(int(uid), req.Email, req.Name, h.cfg.JWTSecret)
	response.JSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"token": token,
		"user": map[string]any{
			"id":    uid,
			"name":  req.Name,
			"email": req.Email,
		},
	})
}

// Login verifies user credentials and issues a signed JWT.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid payload")
		return
	}

	uid, name, storedHash, err := h.repo.GetUserByEmail(req.Email)
	if err != nil || crypto.HashPassword(req.Password) != storedHash {
		response.Error(w, http.StatusUnauthorized, "Email atau password salah.")
		return
	}

	token, _ := crypto.CreateJWT(uid, req.Email, name, h.cfg.JWTSecret)
	response.JSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"token": token,
		"user": map[string]any{
			"id":    uid,
			"name":  name,
			"email": req.Email,
		},
	})
}

// Google handles Google OAuth2 verified profile synchronization.
func (h *Handler) Google(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.GoogleAuthReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		response.Error(w, http.StatusBadRequest, "Email akun Google tidak valid.")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)
	if name == "" {
		parts := strings.Split(email, "@")
		name = parts[0]
	}

	uid, err := h.repo.UpsertGoogleUser(name, email, crypto.HashPassword("google_oauth_authorized"))
	if err != nil {
		response.DBError(w, err)
		return
	}

	token, _ := crypto.CreateJWT(uid, email, name, h.cfg.JWTSecret)
	response.JSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"token": token,
		"user": map[string]any{
			"id":    uid,
			"name":  name,
			"email": email,
		},
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

