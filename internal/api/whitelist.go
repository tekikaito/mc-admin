package api

import (
	"mc-admin/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleGetWhitelist(whitelistService *services.WhitelistService) gin.HandlerFunc {
	return func(c *gin.Context) {
		whitelistInfo, err := whitelistService.GetWhitelistInfo()
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{
				"DetailedError": err.Error(),
			})
			return
		}

		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusOK, "whitelist.html", gin.H{
				"Players": whitelistInfo.PlayerNames,
				"Count":   len(whitelistInfo.PlayerNames),
				"Enabled": whitelistInfo.Enabled,
			})
			return
		}

		data := getCommonPageData(c)
		data["ActiveModule"] = "whitelist"
		data["Players"] = whitelistInfo.PlayerNames
		data["Count"] = len(whitelistInfo.PlayerNames)
		data["Enabled"] = whitelistInfo.Enabled
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

func handleToggleWhitelist(whitelistService *services.WhitelistService) gin.HandlerFunc {
	return func(c *gin.Context) {
		whitelistInfo, err := whitelistService.GetWhitelistInfo()
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{
				"DetailedError": err.Error(),
			})
			return
		}

		if whitelistInfo.Enabled {
			err = whitelistService.DisableWhitelist()
		} else {
			err = whitelistService.EnableWhitelist()
		}

		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{
				"DetailedError": err.Error(),
			})
			return
		}

		c.Redirect(http.StatusSeeOther, "/whitelist")
	}
}
