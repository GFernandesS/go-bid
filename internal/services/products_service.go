package services

import (
	"context"
	"errors"

	"github.com/GFernandesS/go-bid/internal/store/pgstore"
	"github.com/GFernandesS/go-bid/internal/usecase/products"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductsService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

var ErrProductNotFound = errors.New("product not found")

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

func (ps *ProductsService) GetProductById(ctx context.Context, id uuid.UUID) (pgstore.Product, error) {
	product, err := ps.queries.GetProductById(ctx, id)

	if err != nil {
		return pgstore.Product{}, ErrProductNotFound
	}

	return product, nil
}
