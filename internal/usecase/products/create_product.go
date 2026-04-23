package products

import (
	"context"
	"time"

	"github.com/GFernandesS/go-bid/internal/validator"
)

type CreateProductRequest struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	BasePrice   float64   `json:"base_price"`
	AuctionEnd  time.Time `json:"auction_end"`
}

const minAuctionDuration = time.Hour * 2

func (r CreateProductRequest) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(r.Name), "product_name", "this field cannot be blank")

	eval.CheckField(validator.NotBlank(r.Description), "product_description", "this field cannot be blank")

	eval.CheckField(validator.MinChars(r.Description, 10) && validator.MaxChars(r.Description, 255), "description", "this field must have a length between 10 and 255")

	eval.CheckField(r.BasePrice > 0, "base_price", "this field must be greater than zero")

	eval.CheckField(r.AuctionEnd.Sub(time.Now()) >= minAuctionDuration, "auction_end", "must be at least two hours duration")

	return eval
}
