package main

import (
	"log"

	"github.com/oyinetare/url-shortener/config"
	"github.com/oyinetare/url-shortener/repository"
	"github.com/oyinetare/url-shortener/server"
)

func main() {
	// load config
	cfg := config.LoadConfig()

	// connect to db
	repo, err := repository.Connect(
		cfg.DB.Host,
		cfg.DB.Database,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Port,
	)

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := repo.Disconnect(); err != nil {
			log.Printf("Error disconnecting from database: %v", err)
		}
	}()

	log.Println("Connected. Starting server...")

	// create and start server
	srv := server.New(repo, cfg)
	log.Printf("Server started successfully, running on port %d", cfg.Port)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
