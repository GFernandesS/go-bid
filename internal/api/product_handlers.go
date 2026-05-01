package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/GFernandesS/go-bid/internal/jsonutils"
	"github.com/GFernandesS/go-bid/internal/services"
	"github.com/GFernandesS/go-bid/internal/usecase/products"
	"github.com/google/uuid"
)

func (api *Api) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	content, problems, err := jsonutils.DecodeValidJson[products.CreateProductRequest](r)

	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusBadRequest, problems)

		return
	}

	userId, ok := api.Sessions.Get(r.Context(), AuthenticatedUserSessionKey).(uuid.UUID)

	if !ok {
		jsonutils.EncodeInternalError(w, r)
		return
	}

	id, err := api.ProductService.CreateProduct(r.Context(), content, userId)

	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, map[string]any{
			"error": "failed to create product, try again later",
		})

		return
	}

	createAuctionRoom(content, id, api)

	_ = jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"message": "auction has started with success",
		"id":      id,
	})
}

func (api *Api) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := api.ProductService.ListProducts(r.Context())

	if err != nil {
		slog.Error("failed to list products", err)

		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "failed to list products",
		})

		return
	}

	_ = jsonutils.EncodeJson(w, r, http.StatusOK, rows)
}

func createAuctionRoom(req products.CreateProductRequest, productId uuid.UUID, api *Api) {
	auctionContext, _ := context.WithDeadline(context.Background(), req.AuctionEnd)

	auctionRoom := services.NewAuctionRoom(auctionContext, productId, &api.BidsService)

	api.AuctionLobby.AddRoom(auctionRoom)

	go auctionRoom.Run()
}
