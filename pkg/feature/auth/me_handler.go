package auth

import (
	"net/http"

	"plantcare-iot/pkg/middleware"
	"plantcare-iot/pkg/response"
)

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
