package main

import (
	"log"
	"net/http"
	"os"

	"github.com/VIctor-teles-Dev/write-better-codes/apps/backend-api/internal/db"
	"github.com/VIctor-teles-Dev/write-better-codes/apps/backend-api/internal/handler"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var pinger handler.Pinger
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		pool, err := db.Open(databaseURL)
		if err != nil {
			log.Fatalf("invalid DATABASE_URL: %v", err)
		}
		defer pool.Close()
		pinger = pool
	} else {
		log.Print("DATABASE_URL not set; /readyz will not check the database")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.Health)
	mux.Handle("GET /readyz", handler.Ready(pinger))

	log.Printf("backend-api listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
