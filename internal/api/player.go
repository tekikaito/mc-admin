package api

import (
	"net/http"
	"rcon-web/internal/services"
	"strings"

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

func handleGetKickPlayerDialog() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := strings.TrimSpace(c.Param("name"))
		if name == "" {
			c.HTML(http.StatusOK, "error.html", nil)
			return
		}
		c.HTML(http.StatusOK, "kick_player.html", gin.H{
			"PlayerName": name,
		})
	}
}

func handleKickPlayer(serverService *services.ServerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := strings.TrimSpace(c.Param("name"))
		if name == "" {
			c.HTML(http.StatusOK, "error.html", nil)
			return
		}
		reason := strings.TrimSpace(c.PostForm("reason"))
		if err := serverService.KickPlayerByName(name, reason); err != nil {
			c.HTML(http.StatusOK, "kick_player.html", gin.H{
				"PlayerName": name,
				"Reason":     reason,
				"Error":      err.Error(),
			})
			return
		}
		c.HTML(http.StatusOK, "kick_player_success.html", gin.H{
			"PlayerName": name,
		})
	}
}
