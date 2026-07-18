package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/alamjad/marketplace/internal/delivery/http/handler"
	"github.com/alamjad/marketplace/internal/delivery/http/middleware"
)

// RouterDeps holds the handlers and middleware wired into the router.
type RouterDeps struct {
	AllowedOrigins []string
	UploadDir      string
	Auth           *middleware.Authenticator
	Health         *handler.HealthHandler
	AuthH          *handler.AuthHandler
	Catalog        *handler.CatalogHandler
	Listing        *handler.ListingHandler
}

// NewRouter builds the chi router with the /api/v1 tree and static uploads.
func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   deps.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", deps.Health.Health)

	// Serve uploaded files.
	fs := http.StripPrefix("/uploads/", http.FileServer(http.Dir(deps.UploadDir)))
	r.Handle("/uploads/*", fs)

	auth := deps.Auth

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", deps.Health.Health)

		// --- auth ---
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", deps.AuthH.Register)
			r.Post("/login", deps.AuthH.Login)
			r.Post("/refresh", deps.AuthH.Refresh)
			r.With(auth.Require).Get("/me", deps.AuthH.Me)
		})

		// --- catalog (public reads) ---
		r.Get("/categories", deps.Catalog.ListCategories)
		r.Get("/categories/{slug}", deps.Catalog.GetCategoryBySlug)
		r.Get("/categories/{id}/fields", deps.Catalog.GetCategoryFields)
		r.Get("/locations", deps.Catalog.ListLocations)

		// --- listings (public reads) ---
		r.Get("/listings", deps.Listing.Search)
		r.Get("/listings/{id}", deps.Listing.Get)

		// --- authenticated actions ---
		r.Group(func(r chi.Router) {
			r.Use(auth.Require)

			r.Post("/listings", deps.Listing.Create)
			r.Put("/listings/{id}", deps.Listing.Update)
			r.Delete("/listings/{id}", deps.Listing.Delete)
			r.Post("/listings/{id}/images", deps.Listing.UploadImage)
			r.Post("/listings/{id}/favorite", deps.Listing.AddFavorite)
			r.Delete("/listings/{id}/favorite", deps.Listing.RemoveFavorite)

			r.Get("/me/listings", deps.Listing.MyListings)
			r.Get("/me/favorites", deps.Listing.MyFavorites)
		})

		// --- admin ---
		r.Group(func(r chi.Router) {
			r.Use(auth.Require, auth.RequireAdmin)

			r.Post("/admin/categories", deps.Catalog.CreateCategory)
			r.Put("/admin/categories/{id}", deps.Catalog.UpdateCategory)
			r.Delete("/admin/categories/{id}", deps.Catalog.DeleteCategory)
			r.Post("/admin/categories/{id}/fields", deps.Catalog.CreateField)
			r.Get("/admin/listings", deps.Listing.AdminList)
			r.Patch("/admin/listings/{id}/status", deps.Listing.SetStatus)
		})
	})

	return r
}
