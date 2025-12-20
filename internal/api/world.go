package api

import (
	"mc-admin/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleGetWorldStats(worldService *services.WorldService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := worldService.GetWorldStats()
		if err != nil {
			c.String(http.StatusInternalServerError, "Error getting world stats: %v", err)
			return
		}

		c.HTML(http.StatusOK, "world_stats.html", gin.H{
			"Stats": stats,
		})
	}
}
