package main

import (
	"log"

	"github.com/longln/go-social-media/internal/db"
	"github.com/longln/go-social-media/internal/env"
	"github.com/longln/go-social-media/internal/store"
)

func main() {
	address := env.GetString("DB_ADDRESS", "postgres://admin:adminpassword@localhost:5432/socialnetwork?sslmode=disable")
	log.Println("database address: ", address)
	dbConn, err := db.New(address, 3, 3, "15m")
	if err != nil {
		panic(err)
	}
	defer dbConn.Close()

	store := store.NewStorage(dbConn)

	db.Seed(store)
}