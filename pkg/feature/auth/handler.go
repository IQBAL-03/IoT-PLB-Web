// Package auth provides HTTP handlers and repository logic for user authentication.
package auth

import (
	"plantcare-iot/pkg/config"
)

// Handler coordinates authentication requests across login, registration, and OTP verification.
type Handler struct {
	cfg  config.Config
	repo *Repository
}

// NewHandler instantiates an Auth Handler with configuration and repository dependencies.
func NewHandler(cfg config.Config, repo *Repository) *Handler {
	return &Handler{cfg: cfg, repo: repo}
}
