package bids

import "github.com/google/uuid"

type PlaceBidRequest struct {
	ProductId uuid.UUID `json:"product_id"`
	Amount    float64   `json:"amount"`
}
