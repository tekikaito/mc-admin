package main

import (
	"context"
	"log"
	"mc-admin/internal/api"
	ashcon_client "mc-admin/internal/clients"
	"mc-admin/internal/rcon"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func getEnv(s string) *string {
	value := os.Getenv(s)
	if value == "" {
		return nil
	}
	return &value111111
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

func buildAuthConfig() api.AuthConfig {
	allowedUsersEnv := os.Getenv("DISCORD_ALLOWED_USER_IDS")
	var allowedUsers []string
	if allowedUsersEnv != "" {
		candidates := strings.Split(allowedUsersEnv, ",")
		for _, candidate := range candidates {
			trimmed := strings.TrimSpace(candidate)
			if trimmed != "" {
				allowedUsers = append(allowedUsers, trimmed)
			}
		}
	}

	return api.AuthConfig{
		ClientID:       os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret:   os.Getenv("DISCORD_CLIENT_SECRET"),
		RedirectURI:    os.Getenv("DISCORD_REDIRECT_URI"),
		SessionSecret:  os.Getenv("SESSION_SECRET"),
		AllowedUserIDs: allowedUsers,
	}
}

func main() {
	loadDotEnvFile()
	ashconClient := ashcon_client.NewMojangUserNameChecker()
	rconClient := setupMinecraftRcon()
	defer rconClient.Close() // Ensure RCON connection is closed on exit

	r, err := api.InitializeWebServer(api.WebServerOptions{
		MinecraftRconClient: rconClient,
		AshconClient:        ashconClient,
		AuthConfig:          buildAuthConfig(),
	})
	if err != nil {
		log.Fatalf("failed to initialize web server: %v", err)
	}

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
