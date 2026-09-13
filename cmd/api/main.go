package main

import (
	"log"

	"github.com/JonasCappe/book-management-api/internal/db"
	"github.com/JonasCappe/book-management-api/internal/env"
	"github.com/JonasCappe/book-management-api/internal/store"
)

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":3000"),
		db: dbConfig{
			dsn:          env.GetString("DB_DSN", "postgres://postgres:postgres@localhost/book-management?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 50),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 50),
			maxIdleTime:  env.GetTimeDuration("DB_MAX_IDLE_TIME"),
		},
	}


	db, err := db.New(
		cfg.db.dsn,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	log.Println("database connection pool sucessfully established")

	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))

}
