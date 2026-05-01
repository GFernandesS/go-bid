package api

import (
	"os"

	"github.com/GFernandesS/go-bid/internal/configuration"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
)

func (api *Api) BindRoutes() {
	api.Router.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger, api.Sessions.LoadAndSave)

	if configuration.ShouldUseCSRFToken() {
		csrfMiddleware := csrf.Protect([]byte(os.Getenv("GOBID_CSTF_KEY")), csrf.Secure(false))

		api.Router.Use(csrfMiddleware)
	}

	api.Router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			if configuration.ShouldUseCSRFToken() {
				r.Get("/csrftoken", api.handleGetCSRFToken)
			}

			r.Route("/users", func(r chi.Router) {
				r.Post("/signup", api.handleSignupUser)
				r.Post("/login", api.handleLoginUser)
				r.Group(func(r chi.Router) {
					r.Use(api.AuthMiddleware)
					r.Post("/logout", api.handleLogoutUser)
				})
			})

			r.Route("/products", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(api.AuthMiddleware)

					r.Get("/subscribe/{product_id}", api.handleSubscribeUserToAuction)

					r.Post("/", api.handleCreateProduct)
					r.Get("/", api.handleGetProducts)
				})
			})
		})

	})
}
