package api

import (
	"net/http"

	"github.com/GFernandesS/go-bid/internal/jsonutils"
	"github.com/gorilla/csrf"
)

func (api *Api) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !api.Sessions.Exists(r.Context(), "AuthenticatedUserId") {
			_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{})
		}

		next.ServeHTTP(w, r)
	})
}

func (api *Api) HandleGetCSRFToken(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)

	_ = jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"token": token,
	})
}
