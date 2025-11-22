package api

import (
	"rcon-web/internal/rcon"

	"github.com/gin-gonic/gin"
)

func initializeWebServer() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	return r
}

func initializeWebServerRoutes(mcRcon *rcon.MinecraftRconClient, r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{})
	})
	r.GET("/player", handleGetPlayerList(mcRcon))
}

func InitializeWebServer(mcRcon *rcon.MinecraftRconClient) *gin.Engine {
	r := initializeWebServer()
	initializeWebServerRoutes(mcRcon, r)
	return r
}
