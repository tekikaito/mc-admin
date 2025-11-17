package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorcon/rcon"
)


type ServerPlayerInfo struct {
	PlayerNames []string `json:"player_names"`
	OnlineCount int      `json:"online_count"`
	MaxCount    int      `json:"max_count"`
}

func getPlayersOnline() ServerPlayerInfo {
	conn, err := rcon.Dial("127.0.0.1:25575", "")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	response, err := conn.Execute("list")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)

	// Example response: "There are 3 of a max 20 players online: Player1, Player2, Player3"
	var info ServerPlayerInfo
	var commaSeparatedNames string
	fmt.Sscanf(response, "There are %d of a max of %d players online: %s", &info.OnlineCount, &info.MaxCount, &commaSeparatedNames)
	
	if info.OnlineCount == 0 {
		return info
	}

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