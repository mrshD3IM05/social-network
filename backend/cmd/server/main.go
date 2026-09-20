package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"sn-backend/internal/db/sqlite"
	"sn-backend/internal/handler"
	"sn-backend/internal/middleware"
	"sn-backend/internal/repository"
	"sn-backend/internal/server"
)

func main() {
	databasePath := os.Getenv("DB_PATH")
	if databasePath == "" {
		databasePath = "sn.db"
	}
	if err := sqlite.InitDB(databasePath); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	repo := repository.New(sqlite.DB)
	handlers := handler.New(repo)
	server.RegisterRoutes(mux, handlers)
	server.RegisterRoutesExtra(mux, repo, handlers)
	server.StartSessionCleanup(repo, time.Hour)
	err := http.ListenAndServe(":8080", middleware.RateLimit(mux))
	if err != nil {
		panic(err)
	}

}
