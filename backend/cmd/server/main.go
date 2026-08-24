package main

import (
	"log"
	"net/http"
	"sn-backend/internals/db/sqlite"
	"sn-backend/internals/handler"
	"sn-backend/internals/middleware"
	"sn-backend/internals/repository"
	"sn-backend/internals/server"
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
