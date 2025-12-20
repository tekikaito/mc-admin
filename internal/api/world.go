package api

import (
	"mc-admin/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleGetWorld() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusOK, "world.html", gin.H{})
			return
		}

		data := getCommonPageData(c)
		data["ActiveModule"] = "world"
		c.HTML(http.StatusOK, "index.html", data)
	}
}

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
