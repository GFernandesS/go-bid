package services

import (
	"context"

	"github.com/GFernandesS/go-bid/internal/store/pgstore"
	"github.com/GFernandesS/go-bid/internal/usecase/products"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductsService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

func NewProductsService(pool *pgxpool.Pool) ProductsService {
	return ProductsService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (ps *ProductsService) CreateProduct(ctx context.Context, req products.CreateProductRequest, sellerId uuid.UUID) (uuid.UUID, error) {
	id, err := ps.queries.CreateProduct(ctx, pgstore.CreateProductParams{
		SellerID:    sellerId,
		ProductName: req.Name,
		Description: req.Description,
		BasePrice:   req.BasePrice,
		AuctionEnd:  req.AuctionEnd,
	})

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (ps *ProductsService) ListProducts(ctx context.Context) ([]pgstore.ListProductsRow, error) {
	rows, err := ps.queries.ListProducts(ctx)

	if err != nil {
		return nil, err
	}

	return rows, nil
}
