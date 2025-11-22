package api

import (
	"net/http"
	"rcon-web/internal/rcon"

	"github.com/gin-gonic/gin"
)

func handleGetPlayerList(mcRcon *rcon.MinecraftRconClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := mcRcon.GetServerPlayerInfo()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get player info"})
			return
		}
		c.HTML(http.StatusOK, "player_list.html", gin.H{
			"Players":     info.PlayerNames,
			"OnlineCount": info.OnlineCount,
			"MaxCount":    info.MaxCount,
		})
	}
}
