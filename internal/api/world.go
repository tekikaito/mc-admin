package api

import (
	"mc-admin/internal/services"
	"net/http"
	"strconv"

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

// getClockData returns the current clock data for templates
func getClockData(worldService *services.WorldService) (gin.H, error) {
	ticks, err := worldService.GetDaytime()
	if err != nil {
		return nil, err
	}
	phase := worldService.GetPhaseFromTicks(ticks)
	rotation := ((float64(ticks) - 6000) / 24000) * 360
	return gin.H{
		"Ticks":    ticks,
		"Phase":    phase,
		"Rotation": rotation,
	}, nil
}

func handleGetClock(worldService *services.WorldService) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := getClockData(worldService)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error getting time: %v", err)
			return
		}
		c.HTML(http.StatusOK, "clock_view.html", data)
	}
}

func handleGetClockEdit(worldService *services.WorldService) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := getClockData(worldService)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error getting time: %v", err)
			return
		}
		c.HTML(http.StatusOK, "clock_edit.html", data)
	}
}

func handleSetTime(worldService *services.WorldService) gin.HandlerFunc {
	return func(c *gin.Context) {
		timeValue := c.PostForm("time")
		if timeValue == "" {
			c.String(http.StatusBadRequest, "Time value is required")
			return
		}

		// Validate that time is a valid number or preset
		validPresets := map[string]bool{
			"day": true, "noon": true, "night": true, "midnight": true,
		}
		if !validPresets[timeValue] {
			// Must be a number
			if _, err := strconv.Atoi(timeValue); err != nil {
				c.String(http.StatusBadRequest, "Invalid time value: must be a number or preset (day, noon, night, midnight)")
				return
			}
		}

		_, err := worldService.SetTime(timeValue)
		if err != nil {
			c.Header("HX-Trigger", `{"showToast": {"message": "Failed to set time: `+err.Error()+`", "type": "error"}}`)
			c.String(http.StatusInternalServerError, "Error setting time: %v", err)
			return
		}

		// Return the clock view with updated data
		data, err := getClockData(worldService)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error getting time: %v", err)
			return
		}

		c.Header("HX-Trigger", `{"showToast": {"message": "Time set to `+timeValue+`", "type": "success"}}`)
		c.HTML(http.StatusOK, "clock_view.html", data)
	}
}
