// Package middleware provides HTTP middlewares for CORS, compression, and authentication.
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"plantcare-iot/pkg/config"
	"plantcare-iot/pkg/model"
	"plantcare-iot/pkg/crypto"
)

type gzipWriter struct {
	io.Writer
	http.ResponseWriter
}

func (g gzipWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

// Gzip compresses responses when client provides Accept-Encoding: gzip header.
func Gzip(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		next(gzipWriter{Writer: gz, ResponseWriter: w}, r)
	}
}

// CORS attaches cross-origin resource sharing headers to allow web clients.
func CORS(allowedOrigin string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := allowedOrigin
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// Chain wraps an HTTP handler with standard CORS and Gzip middlewares.
func Chain(cfg config.Config, h http.HandlerFunc) http.HandlerFunc {
	return CORS(cfg.AllowedOrigin, Gzip(h))
}

// GetUserFromRequest extracts and verifies bearer JWT token from request header.
func GetUserFromRequest(r *http.Request, secret string) *model.UserClaims {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := crypto.VerifyJWT(token, secret)
		if err == nil {
			return claims
		}
	}
	return nil
}

