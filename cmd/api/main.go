package main

import (
	"log"
	"os"

	"github.com/Firebreather-heart/ningen/internal/api"
	"github.com/Firebreather-heart/ningen/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	router := api.SetupRouter(cfg)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
		os.Exit(1)
	}
}
