package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmed-wassim/wassimo-cart/db/gen"
	"github.com/ahmed-wassim/wassimo-cart/internal/cart"
	"github.com/ahmed-wassim/wassimo-cart/internal/catalog"
	"github.com/ahmed-wassim/wassimo-cart/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db pool: %v", err)
	}
	defer pool.Close()

	store := cart.NewStore(gen.New(pool), catalog.New(cfg.CatalogURL))
	h := cart.NewHandlers(store, pool)

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger)
	r.Get("/health", h.Health)
	r.Get("/ready", h.Ready)
	r.Mount("/cart", h.Routes())

	log.Printf("cart listening on :%s catalog=%s", cfg.Port, cfg.CatalogURL)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
