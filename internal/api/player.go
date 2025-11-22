package api

import (
	"net/http"
	"rcon-web/internal/services"

	"github.com/gin-gonic/gin"
)

func handleGetServerInfo(serverService *services.ServerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := serverService.GetServerPlayerInfo()
		if err != nil {
			// Return HTML error for HTMX compatibility
			c.HTML(http.StatusOK, "error.html", nil)
			return
		}
		c.HTML(http.StatusOK, "player_list.html", gin.H{
			"Players":     info.PlayerNames,
			"OnlineCount": info.OnlineCount,
			"MaxCount":    info.MaxCount,
		})
	}
}
