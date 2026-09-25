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
	repo := repository.New(sqlite.DB)
	handlers := handler.New(repo)
	server.RegisterRoutes(mux, handlers)
	server.RegisterMessageRoutes(mux, repo, handlers)
	err := http.ListenAndServe(":8080", middleware.RateLimit(mux))
	if err != nil {
		panic(err)
	}

}
