package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rcon-web/internal/api"
	ashcon_client "rcon-web/internal/clients"
	"rcon-web/internal/rcon"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func getEnv(s string) *string {
	value := os.Getenv(s)
	if value == "" {
		return nil
	}
	return &value
}

func loadDotEnvFile() {
	if stat, err := os.Stat(".env"); err != nil || stat.IsDir() {
		log.Println(".env file not found, skipping load")
		return
	}

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func setupMinecraftRcon() *rcon.MinecraftRconClient {
	rconPassword := getEnv("RCON_PASSWORD")
	if rconPassword == nil {
		log.Fatal("RCON_PASSWORD environment variable is not set")
	}
	rconHost := getEnv("RCON_HOST")
	if rconHost == nil {
		rconHost = new(string)
		*rconHost = "localhost"
	}
	rconPort := getEnv("RCON_PORT")
	if rconPort == nil {
		rconPort = new(string)
		*rconPort = "25575"
	}
	mcRcon := rcon.NewMinecraftRconClient(*rconHost, *rconPort, *rconPassword)
	return mcRcon
}

func main() {
	loadDotEnvFile()
	ashconClient := ashcon_client.NewMojangUserNameChecker()
	rconClient := setupMinecraftRcon()
	defer rconClient.Close() // Ensure RCON connection is closed on exit

	r := api.InitializeWebServer(api.WebServerOptions{
		MinecraftRconClient: rconClient,
		AshconClient:        ashconClient,
	})

	// Create HTTP server with custom settings for graceful shutdown
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Println("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run server: %v", err)
		}
	}()

	// Block until we receive a signal
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
