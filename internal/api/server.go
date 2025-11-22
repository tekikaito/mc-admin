package api

import (
	"rcon-web/internal/rcon"
	"rcon-web/internal/services"

	"github.com/gin-gonic/gin"
)

func initializeWebServer() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	return r
}

func initializeWebServerRoutes(r *gin.Engine, serverService *services.ServerService) {
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{})
	})
	r.GET("/player", handleGetServerInfo(serverService))
}

func InitializeWebServer(mcRcon *rcon.MinecraftRconClient) *gin.Engine {
	r := initializeWebServer()
	serverService := services.NewServerServiceFromRconClient(mcRcon)
	initializeWebServerRoutes(r, serverService)
	return r
}
