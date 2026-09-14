package main

import (
	"net/http"
	"time"

	"github.com/JonasCappe/book-management-api/docs"
	"github.com/JonasCappe/book-management-api/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

type application struct {
	config config
	store  store.Storage
	logger *slog.Logger
}

type dbConfig struct {
	dsn          string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  time.Duration
}

type config struct {
	addr   string
	apiURL string
	db     dbConfig
	env    string
}

func (app *application) mount() http.Handler { // *chi.Mux
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(app.requestLogger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", app.healthCheckHandler)

	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	r.Route("/v1", func(r chi.Router) {

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
	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdownSignal, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	serverError := make(chan error, 1)

	go func() {
		app.logger.Info(
			"server started",
			"address", app.config.addr,
		)
		serverError <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err

	case <-shutdownSignal.Done():
		app.logger.Info("shutdown signal received")
	}

	shutDownCtx, cancel := context.WithTimeout(
		context.Background(),
		10^time.Second,
	)

	defer cancel()

	if err := srv.Shutdown(shutDownCtx); err != nil {
		return err
	}

	err := <-serverError
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	app.logger.Info("server shutdown complete")

	return nil
}
