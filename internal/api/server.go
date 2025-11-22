package api

import (
	"os"
	"rcon-web/internal/rcon"
	"rcon-web/internal/services"

	"github.com/gin-gonic/gin"
)

func initializeWebServer() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	return r
}

func getIndexPageHandler() gin.HandlerFunc {
	serverName := os.Getenv("SERVER_NAME")
	if serverName == "" {
		serverName = "Minecraft Server"
	}
	serverHost := os.Getenv("RCON_HOST")
	if serverHost == "" {
		serverHost = "localhost"
	}
	serverPort := os.Getenv("GAME_PORT")
	if serverPort == "" {
		serverPort = "25565"
	}
	serverVersion := os.Getenv("SERVER_VERSION")
	if serverVersion == "" {
		serverVersion = "Unknown Version"
	}
	serverDescription := os.Getenv("SERVER_DESCRIPTION")
	if serverDescription == "" {
		serverDescription = "Live status for your community"
	}

	return func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"ServerName":        serverName,
			"ServerHost":        serverHost,
			"ServerPort":        serverPort,
			"ServerVersion":     serverVersion,
			"ServerDescription": serverDescription,
		})
	}
}

func initializeWebServerRoutes(r *gin.Engine, serverService *services.ServerService, whitelistService *services.WhitelistService) {
	r.GET("/", getIndexPageHandler())
	r.GET("/server-info", handleGetServerInfo(serverService))
	r.GET("/whitelist", handleGetWhitelist(whitelistService))
}

func InitializeWebServer(mcRcon *rcon.MinecraftRconClient) *gin.Engine {
	r := initializeWebServer()
	serverService := services.NewServerServiceFromRconClient(mcRcon)
	whitelistService := services.NewWhitelistServiceFromRconClient(mcRcon)
	initializeWebServerRoutes(r, serverService, whitelistService)
	return r
}
