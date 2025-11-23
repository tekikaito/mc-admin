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

// get name from path parameter and remove from whitelist
func handleRemoveNameFromWhitelist(whitelistService *services.WhitelistService) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		err := whitelistService.RemoveNameFromWhitelist(name)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", nil)
			return
		}
		c.Redirect(http.StatusSeeOther, "/whitelist")
	}
}
