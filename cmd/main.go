package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/useopenward/openward/internal/api"
	"github.com/useopenward/openward/internal/db"
	"github.com/useopenward/openward/internal/proxy"
)

func main() {
	loadDotEnv(".env")

	dbPath := env("OPENWARD_DB", "openward.db")
	proxyAddr := env("OPENWARD_ADDR", ":8080")
	adminAddr := env("OPENWARD_ADMIN_ADDR", ":9090")

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	// proxy server
	proxyServer := &http.Server{
		Addr:         proxyAddr,
		Handler:      proxy.NewHandler(database.Reader, database.Writer),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// admin API server
	adminServer := api.NewServer(database)
	adminServer.Addr = adminAddr

	log.Printf("openward proxy listening on %s", proxyAddr)
	log.Printf("openward admin API listening on %s", adminAddr)

	errc := make(chan error, 2)
	go func() { errc <- proxyServer.ListenAndServe() }()
	go func() { errc <- adminServer.ListenAndServe() }()

	if err := <-errc; err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)

		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
}
