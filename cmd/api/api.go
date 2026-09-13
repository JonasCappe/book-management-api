package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type application struct {
	config config
}

type config struct {
	addr string
	
}

func (app *application) mount() http.Handler { // *chi.Mux
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	// TODO Setup Endpoints
	r.Route("/v1", func(r chi.Router) {
		// General
		r.Get("/health", app.healthCheckHandler)
	})


	return r
}

func (app *application) run(mux http.Handler) error {
	

	srv := http.Server{
		Addr: app.config.addr,
		Handler: mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	log.Printf("server started om port: %s", app.config.addr)

	return srv.ListenAndServe()
}