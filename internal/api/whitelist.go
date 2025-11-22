package api

import (
	"net/http"
	"rcon-web/internal/services"

	"github.com/gin-gonic/gin"
)

func handleGetWhitelist(whitelistService *services.WhitelistService) gin.HandlerFunc {
	return func(c *gin.Context) {
		whitelist, err := whitelistService.GetWhitelist()
		if err != nil {
			c.HTML(http.StatusOK, "error.html", nil)
			return
		}
		c.HTML(http.StatusOK, "whitelist.html", gin.H{
			"Players": whitelist,
			"Count":   len(whitelist),
		})
	}
}
