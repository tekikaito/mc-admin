package api

import (
	"mc-admin/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleGetWhitelist(whitelistService *services.WhitelistService) gin.HandlerFunc {
	return func(c *gin.Context) {
		whitelist, err := whitelistService.GetWhitelist()
		if err != nil {
			c.HTML(http.StatusOK, "error.html", nil)
			return
		}

		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusOK, "whitelist.html", gin.H{
				"Players": whitelist,
				"Count":   len(whitelist),
			})
			return
		}

		data := getCommonPageData(c)
		data["ActiveModule"] = "whitelist"
		data["Players"] = whitelist
		data["Count"] = len(whitelist)
		c.HTML(http.StatusOK, "index.html", data)
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

func handleAddNameToWhitelist(whitelistService *services.WhitelistService) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.PostForm("playerName")
		err := whitelistService.AddNameToWhitelist(name)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", nil)
			return
		}
		c.Redirect(http.StatusSeeOther, "/whitelist")
	}
}
