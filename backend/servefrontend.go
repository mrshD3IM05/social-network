package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	addr := flag.String("addr", ":5500", "listen address")
	dir := flag.String("dir", "frontend", "directory to serve")
	flag.Parse()
	if _, err := os.Stat(*dir); err != nil {
		log.Fatalf("directory %s not found", *dir)
	}
	log.Printf("serving %s on %s", *dir, *addr)
	log.Fatal(http.ListenAndServe(*addr, http.FileServer(http.Dir(*dir))))
}
