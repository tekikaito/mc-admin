package main

import (
	"context"
	"log"
	"mc-admin/internal/api"
	"mc-admin/internal/clients/ashcon"
	"mc-admin/internal/clients/rcon"
	"mc-admin/internal/config"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	config.LoadDotEnvFile(config.NewValidator([]config.EnvVarDefinition{
		{Name: "RCON_HOST", Required: false},
		{Name: "RCON_PORT", Required: false},
		{Name: "RCON_PASSWORD", Required: true},
		{Name: "ENABLE_MINECRAFT_USERNAME_CHECK", Required: false},
		{Name: "MINECRAFT_DATA_DIR", Required: false},
		{Name: "DISCORD_CLIENT_ID", Required: true, FeatureFlag: "ENABLE_DISCORD_OAUTH", ValidationFunc: config.IsInteger},
		{Name: "DISCORD_CLIENT_SECRET", Required: true, FeatureFlag: "ENABLE_DISCORD_OAUTH", ValidationFunc: config.IsNotEmpty},
		{Name: "DISCORD_REDIRECT_URI", Required: true, FeatureFlag: "ENABLE_DISCORD_OAUTH", ValidationFunc: config.IsNotEmpty},
		{Name: "SESSION_SECRET", Required: true, FeatureFlag: "ENABLE_DISCORD_OAUTH", ValidationFunc: config.IsNotEmpty},
	}))

	var ashconClient ashcon.MojangUserNameChecker
	enableUsernameCheck := strings.ToLower(os.Getenv("ENABLE_MINECRAFT_USERNAME_CHECK")) == "true"
	if enableUsernameCheck {
		ashconClient = ashcon.NewMojangUserNameChecker()
	}
	rconClient := rcon.BuildMinecraftRconClientFromEnv()
	defer rconClient.Close() // Ensure RCON connection is closed on exit

	r, err := api.InitializeWebServer(api.WebServerOptions{
		MinecraftRconClient: rconClient,
		AshconClient:        ashconClient,
		AuthConfig:          api.BuildAuthConfigFromEnv(),
	})
	if err != nil {
		log.Fatalf("failed to initialize web server: %v", err)
	}

	// Create HTTP server with custom settings for graceful shutdown
	srv := &http.Server{Addr: ":8080", Handler: r}
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
