package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"sn-backend/internal/db/sqlite"
	"sn-backend/internal/handler"
	"sn-backend/internal/middleware"
	"sn-backend/internal/repository"
	"sn-backend/internal/server"
)

// apiPrefix is stripped before a request reaches the API mux, so the routes in
// server.RegisterRoutes stay prefix-free ("/login", "/posts", ...).
const apiPrefix = "/api/v1"

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	frontendDir := flag.String("frontend", "frontend", "directory with the static frontend")
	flag.Parse()

	if _, err := os.Stat(*frontendDir); err != nil {
		log.Fatalf("frontend directory %s not found", *frontendDir)
	}

	if err := sqlite.InitDB("sn.db"); err != nil {
		log.Fatal(err)
	}

	api := http.NewServeMux()
	server.RegisterRoutes(api, handler.New(repository.New(sqlite.DB)))

	root := http.NewServeMux()
	root.Handle(apiPrefix+"/", http.StripPrefix(apiPrefix, middleware.RateLimit(api)))
	root.Handle("/", http.FileServer(http.Dir(*frontendDir)))

	log.Printf("api on %s%s, frontend %s on %s", *addr, apiPrefix, *frontendDir, *addr)
	if err := http.ListenAndServe(*addr, root); err != nil {
		log.Fatal(err)
	}
}
