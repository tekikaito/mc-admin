package api

import (
	"encoding/json"
	"net/http"
	"os"

	"mc-admin/internal/services"

	"github.com/gin-gonic/gin"
)

// handleGetUserStats renders a user stats overview using the UserStatsService.
func handleGetUserStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := CurrentUser(c)

		minecraftDataDir := os.Getenv("MINECRAFT_DATA_DIR")
		var stats []*services.PlayerStats
		if minecraftDataDir != "" {
			svc, err := services.NewUserStatsService(minecraftDataDir)
			if err == nil {
				ps, _ := svc.GetAllPlayerStats()
				stats = ps
			}
		}

		// marshal stats to JSON for client-side rendering
		var statsJSON string
		if b, err := json.Marshal(stats); err == nil {
			statsJSON = string(b)
		}

		// If this is an HTMX request, return only the partial
		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusOK, "user_stats.html", gin.H{
				"User":      user,
				"Stats":     stats,
				"StatsJSON": statsJSON,
			})
			return
		}

		// Otherwise render the full page (so a reload on /users/stats returns a complete HTML document)
		data := getCommonPageData(c)
		data["User"] = user
		data["Stats"] = stats
		data["StatsJSON"] = statsJSON
		data["ActiveModule"] = "users"
		c.HTML(http.StatusOK, "index.html", data)
	}
}

// handleGetUserStatsByUUID returns a server-rendered partial for a single player's stats.
func handleGetUserStatsByUUID() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.Param("uuid")
		minecraftDataDir := os.Getenv("MINECRAFT_DATA_DIR")
		if minecraftDataDir == "" {
			c.String(http.StatusInternalServerError, "MINECRAFT_DATA_DIR not configured")
			return
		}
		svc, err := services.NewUserStatsService(minecraftDataDir)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error initializing stats service: %v", err)
			return
		}
		ps, err := svc.GetStatsForUUID(uuid)
		if err != nil {
			// If not found and this is an HTMX request, return 404 text; otherwise render full page
			if c.GetHeader("HX-Request") == "true" {
				c.String(http.StatusNotFound, "Stats not found for %s", uuid)
				return
			}
			data := getCommonPageData(c)
			data["ActiveModule"] = "users"
			data["InitialUUID"] = uuid
			c.HTML(http.StatusOK, "index.html", data)
			return
		}
		// If HTMX request, return the partial only
		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusOK, "user_stats_detail.html", gin.H{"Player": ps})
			return
		}

		// Otherwise render the full page and set InitialUUID so the detail is loaded after page load
		data := getCommonPageData(c)
		data["ActiveModule"] = "users"
		// include stats list so the user list shows
		if svcAll, err := services.NewUserStatsService(minecraftDataDir); err == nil {
			if alls, err2 := svcAll.GetAllPlayerStats(); err2 == nil {
				data["Stats"] = alls
			}
		}
		data["InitialUUID"] = uuid
		c.HTML(http.StatusOK, "index.html", data)
	}
}
