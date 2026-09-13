package main

import (
	"log"

	"github.com/JonasCappe/book-management-api/internal/db"
	"github.com/JonasCappe/book-management-api/internal/env"
	"github.com/JonasCappe/book-management-api/internal/store"
	"github.com/go-playground/validator/v10"
)

const version = "0.0.1"

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":3000"),
		db: dbConfig{
			dsn:          env.GetString("DB_DSN", "postgres://postgres:postgres@localhost/book-management?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 50),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 50),
			maxIdleTime:  env.GetTimeDuration("DB_MAX_IDLE_TIME"),
		},
		env: env.GetString("ENVIRONMENT", "development"),
	}

	database, err := db.New(
		cfg.db.dsn,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer database.Close()
	log.Println("database connection pool successfully established")

	storage := store.NewStorage(database)
	validate := validator.New(validator.WithRequiredStructEnabled())
	configureValidator(validate)

	app := &application{
		config:    cfg,
		store:     storage,
		validator: validate,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))

}
