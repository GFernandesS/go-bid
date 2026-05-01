package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/GFernandesS/go-bid/internal/jsonutils"
	"github.com/GFernandesS/go-bid/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (api *Api) handleSubscribeUserToAuction(w http.ResponseWriter, r *http.Request) {
	slog.Info("subscribe user to auction")
	rawProductId := chi.URLParam(r, "product_id")

	productId, err := uuid.Parse(rawProductId)

	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "invalid product id. Must be an valid uuid",
		})

		return
	}

	_, err = api.ProductService.GetProductById(r.Context(), productId)

	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			_ = jsonutils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
				"message": fmt.Sprintf("%s", services.ErrProductNotFound),
			})

			return
		}

		slog.Error("error on subscribe user to auction", err)

		jsonutils.EncodeInternalError(w, r)

		return
	}

	userId, ok := api.Sessions.Get(r.Context(), AuthenticatedUserSessionKey).(uuid.UUID)

	if !ok {
		slog.Error("error to get user to subscribe to auction")

		jsonutils.EncodeInternalError(w, r)

		return
	}

	api.AuctionLobby.Lock()

	room, ok := api.AuctionLobby.Rooms[productId]

	api.AuctionLobby.Unlock()

	if !ok {
		_ = jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "the auction has ended",
		})

		return
	}

	conn, err := api.wsUpgrader.Upgrade(w, r, nil)

	if err != nil {
		slog.Error("could not upgrade connection to websocket", err)
		jsonutils.EncodeInternalError(w, r, "could not upgrade connection to websocket")

		return
	}

	client := services.NewClient(room, conn, userId)

	room.Register <- client

	go client.WriteEventLoop()
	go client.ReadEventLoop()
}
