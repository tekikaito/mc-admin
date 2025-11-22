package main

import (
	"log"
	"os"
	"rcon-web/internal/rcon"

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
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func initRconClient() *rcon.MinecraftRconClient {
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
	mcRcon := initRconClient()

	// Start server on port 8080 (default)
	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
