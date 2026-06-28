package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"github.com/useopenward/openward/internal/core"
	"github.com/useopenward/openward/internal/db"
)

func main() {
	database, err := db.Open("openward.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	limit := 10000000000
	window := 10 * time.Second

	p := &core.Project{
		ID:        "proj_01",
		Name:      "Test Project",
		APIKey:    randomKey(),
		Enabled:   true,
		Upstream:  "http://localhost:9999", // great for testing, reflects requests back
		Algorithm: core.AlgoFixedWindow,
		FWLimit:   &limit,
		FWWindow:  &window,
	}

	if err := db.CreateProject(database.Writer, p); err != nil {
		log.Fatal(err)
	}

	log.Printf("created project, api key: %s", p.APIKey)
}

func randomKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "ow_" + hex.EncodeToString(b)
}
