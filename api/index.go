// Package api exposes the Vercel Serverless Function entry point.
package api

import (
	"net/http"

	"plantcare-iot/pkg/config"
	"plantcare-iot/pkg/database"
	"plantcare-iot/pkg/feature/auth"
	"plantcare-iot/pkg/feature/control"
	"plantcare-iot/pkg/feature/device"
	"plantcare-iot/pkg/feature/history"
	"plantcare-iot/pkg/feature/telemetry"
	"plantcare-iot/pkg/middleware"
	"plantcare-iot/pkg/response"
)

// Handler serves incoming HTTP requests for Vercel Serverless runtime.
func Handler(w http.ResponseWriter, r *http.Request) {
	cfg := config.Load()

	db, err := database.Get(cfg)
	if err != nil {
		response.DBError(w, err)
		return
	}

	// Repositories
	authRepo := auth.NewRepository(db)
	deviceRepo := device.NewRepository(db)
	historyRepo := history.NewRepository(db)
	controlRepo := control.NewRepository(db)

	// Handlers
	authH := auth.NewHandler(cfg, authRepo)
	deviceH := device.NewHandler(cfg, deviceRepo)
	historyH := history.NewHandler(cfg, historyRepo)
	controlH := control.NewHandler(cfg, controlRepo, historyRepo)
	telemetryH := telemetry.NewHandler(db, cfg, controlRepo, historyRepo)

	mux := http.NewServeMux()

	wrap := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.Chain(cfg, h)
	}

	// Auth Endpoints
	mux.HandleFunc("/api/auth/register", wrap(authH.Register))
	mux.HandleFunc("/api/auth/register/request-otp", wrap(authH.RegisterRequestOTP))
	mux.HandleFunc("/api/auth/register/verify-otp", wrap(authH.RegisterVerifyOTP))
	mux.HandleFunc("/api/auth/login", wrap(authH.Login))
	mux.HandleFunc("/api/auth/google", wrap(authH.Google))
	mux.HandleFunc("/api/auth/me", wrap(authH.Me))

	// Device Endpoints
	mux.HandleFunc("/api/devices", wrap(deviceH.List))
	mux.HandleFunc("/api/devices/claim", wrap(deviceH.Claim))
	mux.HandleFunc("/api/devices/unpair", wrap(deviceH.Unpair))

	// Control & Status Endpoints
	mux.HandleFunc("/api/status", wrap(controlH.Status))
	mux.HandleFunc("/api/pump/on", wrap(controlH.PumpOn))
	mux.HandleFunc("/api/pump/off", wrap(controlH.PumpOff))
	mux.HandleFunc("/api/mode", wrap(controlH.Mode))

	// History Endpoints
	mux.HandleFunc("/api/history", wrap(historyH.List))
	mux.HandleFunc("/api/history/clear", wrap(historyH.Clear))

	// Telemetry & Health Endpoints
	mux.HandleFunc("/api/webhook", wrap(telemetryH.Webhook))
	mux.HandleFunc("/api/ping", wrap(telemetryH.Ping))

	mux.ServeHTTP(w, r)
}
