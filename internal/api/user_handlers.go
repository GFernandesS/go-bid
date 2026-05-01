package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/GFernandesS/go-bid/internal/jsonutils"
	"github.com/GFernandesS/go-bid/internal/services"
	"github.com/GFernandesS/go-bid/internal/usecase/users"
)

func (api *Api) handleSignupUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[users.CreateUserRequest](r)

	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)

		return
	}

	id, err := api.UserService.CreateUser(r.Context(), data)

	if err != nil {
		if errors.Is(err, services.ErrDuplicatedEmailOrPassword) {
			_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, map[string]any{
				"error": fmt.Sprintf("%s", err),
			})
		}

		return
	}

	_ = jsonutils.EncodeJson(w, r, http.StatusCreated, id)
}

func (api *Api) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[users.LoginUserRequest](r)

	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)

		return
	}

	id, err := api.UserService.AuthenticateUser(r.Context(), data)

	if err != nil {
		if errors.Is(err, services.ErrInvalidUserOrPassword) {
			_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, map[string]any{
				"error": fmt.Sprintf("%s", err),
			})

			return
		}

		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error",
		})
		return
	}

	err = api.Sessions.RenewToken(r.Context())

	if err != nil {
		slog.Error("Error renewing token on login", err)

		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error",
		})
	}

	api.Sessions.Put(r.Context(), AuthenticatedUserSessionKey, id)

	_ = jsonutils.EncodeJson(w, r, http.StatusNoContent, map[string]any{})
}

func (api *Api) handleLogoutUser(w http.ResponseWriter, r *http.Request) {

	err := api.Sessions.RenewToken(r.Context())

	if err != nil {
		slog.Error("Error renewing token on logout", err)

		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error",
		})
	}

	api.Sessions.Remove(r.Context(), AuthenticatedUserSessionKey)

	_ = jsonutils.EncodeJson(w, r, http.StatusNoContent, map[string]any{})
}
