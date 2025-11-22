package main

import (
	"log"
	"net/http"
	"os"
	"rcon-web/internal/rcon"

	"github.com/gin-gonic/gin"
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


func main() {
  loadDotEnvFile()
  // Create a Gin router with default middleware (logger and recovery)
  r := gin.Default()
  r.LoadHTMLGlob("templates/*")

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

  handleGetIndex := func(c *gin.Context) {
	c.HTML(200, "index.html", gin.H{})
  }
  r.GET("/", handleGetIndex)

  mcRcon := rcon.NewMinecraftRconClient(*rconHost, *rconPort, *rconPassword)

  handleGetPlayerList := func(c *gin.Context) {
	info := mcRcon.GetServerPlayerInfo()
	c.HTML(http.StatusOK, "player_list.html", gin.H{
		"Players":     info.PlayerNames,
		"OnlineCount": info.OnlineCount,
		"MaxCount":    info.MaxCount,
	})
  }

  r.GET("/player", handleGetPlayerList)

  // Start server on port 8080 (default)
  // Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
  if err := r.Run(); err != nil {
    log.Fatalf("failed to run server: %v", err)
  }
}

