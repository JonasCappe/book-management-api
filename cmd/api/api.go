package main

import (
	"log"
	"net/http"
	"time"

	"github.com/JonasCappe/book-management-api/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type application struct {
	config config
	store  store.Storage
}

type dbConfig struct {
	dsn          string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  time.Duration
}

type config struct {
	addr string
	db   dbConfig
	env  string
}

func (app *application) mount() http.Handler { // *chi.Mux
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP) // TODO: REPLACE
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	// TODO Setup Endpoints
	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/books", func(r chi.Router) {
			r.Post("/", app.createBookHandler)
			r.Get("/", app.getBooksHandler)
			r.Route("/{bookID}", func(r chi.Router) {
				r.Get("/", app.getBookHandler)
				r.Get("/history", app.getBookHistoryHandler)
				r.Delete("/", app.deleteBookHandler)
				r.Patch("/", app.patchBookHandler)
			})
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {

	srv := http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("server started on address: %s", app.config.addr)

	return srv.ListenAndServe()
}
