package main

import (
	"log/slog"
	"os"

	"github.com/JonasCappe/book-management-api/internal/db"
	"github.com/JonasCappe/book-management-api/internal/env"
	"github.com/JonasCappe/book-management-api/internal/store"
)

const version = "0.0.1"

//	@title			Go Book management - Backend
//	@description	A small backend service that manages books and keeps history of the changes made to them.

//	@contact.name	Jonas Cap
//	@contact.url	https://www.jonascap.com
//	@contact.email	jonas.cap@outlook.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/
// @schemes					http
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
func main() {
	cfg := config{
		addr:   env.GetString("ADDR", ":8080"),
		apiURL: env.GetString("EXTERNALURL", "localhost:8080"),
		db: dbConfig{
			dsn:          env.GetString("DB_DSN", "postgres://postgres:postgres@localhost/book-management?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 50),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 50),
			maxIdleTime:  env.GetTimeDuration("DB_MAX_IDLE_TIME"),
		},
		env: env.GetString("ENVIRONMENT", "development"),
	}

	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	database, err := db.New(
		cfg.db.dsn,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	defer database.Close()
	logger.Info("database connection pool established")

	storage := store.NewStorage(database)

	app := &application{
		config: cfg,
		store:  storage,
		logger: logger,
	}

	mux := app.mount()

	if err := app.run(mux); err != nil {
		logger.Error("application stopped with error", "error", err)
		os.Exit(1)
	}

}
