package main

import (
	"log"
	"time"

	"github.com/longln/go-social-media/internal/db"
	"github.com/longln/go-social-media/internal/env"
	"github.com/longln/go-social-media/internal/store"
)

func main() {
	cfg := config{
		address:      env.GetString("ADDRESS", ":4000"),
		writeTimeout: 10 * time.Second,
		readTimeout:  5 * time.Second,
		idleTimeout:  60 * time.Second,
		db: dbConfig{
			address: env.GetString("DB_ADDRESS",
				"postgres://admin:adminpassword@localhost:5432/socialnetwork?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 25),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 25),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
	}
	db, err := db.New(cfg.db.address, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Printf("Connected to database %s", cfg.db.address)
	log.Println("database connection pool established")
	store := store.NewStorage(db)

	app := application{
		config: cfg,
		store:  store,
	}

	log.Fatal(app.serve(app.mount()))

}
