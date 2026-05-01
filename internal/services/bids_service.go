package services

import (
	"context"
	"errors"

	"github.com/GFernandesS/go-bid/internal/store/pgstore"
	"github.com/GFernandesS/go-bid/internal/usecase/bids"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var InsufficientBidError = errors.New("the bid value is too low")

type BidsService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

func NewBidsService(pool *pgxpool.Pool) BidsService {
	return BidsService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (bs *BidsService) PlaceBid(ctx context.Context, req bids.PlaceBidRequest, bidderId uuid.UUID) (pgstore.Bid, error) {
	product, err := bs.queries.GetProductById(ctx, req.ProductId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
	}

	highestBid, err := bs.queries.GetHighestBidByProductId(ctx, req.ProductId)

	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
	}

	if product.BasePrice >= req.Amount || highestBid.BidAmount >= req.Amount {
		return pgstore.Bid{}, InsufficientBidError
	}

	highestBid, err = bs.queries.CreateBid(ctx, pgstore.CreateBidParams{
		BidderID:  bidderId,
		ProductID: req.ProductId,
		BidAmount: req.Amount,
	})

	if err != nil {
		return pgstore.Bid{}, err
	}

	return highestBid, nil
}
