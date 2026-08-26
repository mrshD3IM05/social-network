package main

import (
	"log"
	"net/http"
	"sn-backend/internal/db/sqlite"
	"sn-backend/internal/handler"
	"sn-backend/internal/middleware"
	"sn-backend/internal/repository"
	"sn-backend/internal/server"
)

func main() {
	if err := sqlite.InitDB("sn.db"); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	server.RegisterRoutes(mux, handler.New(repository.New(sqlite.DB)))
	err := http.ListenAndServe(":8080", middleware.RateLimit(mux))
	if err != nil {
		panic(err)
	}

}
