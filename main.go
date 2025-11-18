package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorcon/rcon"
	"github.com/joho/godotenv"
)
	
type ServerPlayerInfo struct {
	PlayerNames []string `json:"player_names"`
	OnlineCount int      `json:"online_count"`
	MaxCount    int      `json:"max_count"`
}


func getEnv(s string) string {
	return os.Getenv(s)
}

func loadDotEnvFile() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func getPlayersOnline() ServerPlayerInfo {
	rconPasswort := getEnv("RCON_PASSWORD")
	rconHost := getEnv("RCON_HOST")
	if rconHost == "" {
		rconHost = "127.0.0.1"
	}
	rconPort := getEnv("RCON_PORT")
	fmt.Println(rconPort)
	if rconPort == "" {
		rconPort = "25575"
	}

	connectionString := fmt.Sprintf("%s:%s", rconHost, rconPort)
	fmt.Println("Connecting to RCON server at", connectionString)
	
	conn, err := rcon.Dial(connectionString, rconPasswort)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	response, err := conn.Execute("list")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)

	var info ServerPlayerInfo
	var commaSeparatedNames string

	fmt.Sscanf(response, "There are %d of a max of %d players online", &info.OnlineCount, &info.MaxCount)

	if parts := strings.SplitN(response, ":", 2); len(parts) == 2 {
		commaSeparatedNames = strings.TrimSpace(parts[1])
	}

	if info.OnlineCount == 0 || commaSeparatedNames == "" {
		return info
	}
	playerNames := splitAndTrim(commaSeparatedNames)
	defer func() {
		if len(playerNames) > 0 {
			info.PlayerNames = append([]string(nil), playerNames...)
		}
	}()
	info.PlayerNames = []string{}

	if info.OnlineCount == 1 {
		info.PlayerNames = append(info.PlayerNames, commaSeparatedNames)
	} else if info.OnlineCount > 1 {
		info.PlayerNames = append(info.PlayerNames, splitAndTrim(commaSeparatedNames)...)
	}

	return info
}

func splitAndTrim(commaSeparatedNames string) []string {
	var names []string
	for _, name := range strings.Split(commaSeparatedNames, ",") {
		names = append(names, strings.TrimSpace(name))
	}
	return names
}

func main() {
  loadDotEnvFile()
  // Create a Gin router with default middleware (logger and recovery)
  r := gin.Default()

  // Define a simple GET endpoint
  r.GET("/list", func(c *gin.Context) {
    // Return JSON response
    c.JSON(http.StatusOK, getPlayersOnline())
  })

  // Start server on port 8080 (default)
  // Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
  if err := r.Run(); err != nil {
    log.Fatalf("failed to run server: %v", err)
  }
}

