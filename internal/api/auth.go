package api

import (
	"log/slog"
	"net/http"

	"github.com/GFernandesS/go-bid/internal/jsonutils"
	"github.com/gorilla/csrf"
)

const AuthenticatedUserSessionKey = "AuthenticatedUserId"

func (api *Api) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !api.Sessions.Exists(r.Context(), AuthenticatedUserSessionKey) {
			slog.Info("Test")
			_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{})
		}

		next.ServeHTTP(w, r)
	})
}

func (api *Api) handleGetCSRFToken(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)

	_ = jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"token": token,
	})
}
