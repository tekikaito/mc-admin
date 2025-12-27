package api

import (
	"mc-admin/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleGetCommandConsole() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusOK, "command_console.html", gin.H{})
			return
		}

		data := getCommonPageData(c)
		data["ActiveModule"] = "rcon"
		c.HTML(http.StatusOK, "index.html", data)
	}
}

func handleExecuteRawCommand(commandService *services.CommandService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawCommand := c.PostForm("command")
		response, err := commandService.ExecuteRawCommand(rawCommand)
		if err != nil {
			c.HTML(http.StatusOK, "command_result.html", gin.H{
				"HasError": true,
				"Message":  err.Error(),
				"Command":  rawCommand,
			})
			return
		}

		c.HTML(http.StatusOK, "command_result.html", gin.H{
			"HasError": false,
			"Message":  response,
			"Command":  rawCommand,
		})
	}
}
