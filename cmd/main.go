package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/useopenward/openward/internal/db"
	"github.com/useopenward/openward/internal/proxy"
)

func main() {
	dbPath := env("OPENWARD_DB", "openward.db")
	addr := env("OPENWARD_ADDR", ":8080")

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	handler := proxy.NewHandler(database.Reader, database.Writer)

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("openward listening on %s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
