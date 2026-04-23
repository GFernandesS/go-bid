package api

import (
	"context"
	"net/http"
	"time"

	"github.com/GFernandesS/go-bid/internal/services"
	"github.com/GFernandesS/go-bid/internal/store/pgstore"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Api struct {
	Router         *chi.Mux
	UserService    services.UserService
	ProductService services.ProductsService
	Sessions       *scs.SessionManager
	pool           *pgxpool.Pool
}

func GetApi(ctx context.Context) (*Api, error) {
	router := chi.NewRouter()

	pgPool, err := pgstore.BuildPool(ctx)

	if err != nil {
		return nil, err
	}

	s := scs.New()

	s.Store = pgxstore.New(pgPool)
	s.Lifetime = 24 * time.Hour
	s.Cookie.HttpOnly = true
	s.Cookie.SameSite = http.SameSiteLaxMode

	return &Api{
		Router:         router,
		UserService:    services.NewUserService(pgPool),
		ProductService: services.NewProductsService(pgPool),
		Sessions:       s,
	}, nil
}

func (api *Api) Close() {
	api.pool.Close()
}
