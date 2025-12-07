package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetEnv(s string) *string {
	value := os.Getenv(s)
	if value == "" {
		return nil
	}
	return &value
}

func LoadDotEnvFile(validator *Validator) {
	if stat, err := os.Stat(".env"); err != nil || stat.IsDir() {
		log.Println(".env file not found, skipping load")
		return
	}

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	if validator == nil {
		return
	}

	// Validate environment variables
	if err := validator.Validate(); err != nil {
		log.Fatalf("Environment validation failed: %v", err)
	}
}
