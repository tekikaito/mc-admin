package api

import (
	"mc-admin/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func handleGetCommandConsole(commandService *services.CommandService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "command_console.html", gin.H{})
	}
}

func handleExecuteRawCommand(commandService *services.CommandService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawCommand := c.PostForm("command")
		trimmed := strings.TrimSpace(rawCommand)
		response, err := commandService.ExecuteRawCommand(trimmed)
		if err != nil {
			c.HTML(http.StatusOK, "command_result.html", gin.H{
				"HasError": true,
				"Message":  err.Error(),
				"Command":  trimmed,
			})
			return
		}

		c.HTML(http.StatusOK, "command_result.html", gin.H{
			"HasError": false,
			"Message":  response,
			"Command":  trimmed,
		})
	}
}
