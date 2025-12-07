package api

import (
	ashcon_client "mc-admin/internal/clients"
	"mc-admin/internal/rcon"
	"mc-admin/internal/services"
	"os"

	"github.com/gin-gonic/gin"
)

func initializeWebServer() *gin.Engine {
	r := gin.Default()
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")
	return r
}

func getCommonPageData(c *gin.Context) gin.H {
	serverName := os.Getenv("SERVER_NAME")
	if serverName == "" {
		serverName = "Minecraft Server"
	}
	serverHost := os.Getenv("SERVER_HOST")
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

	user := CurrentUser(c)
	return gin.H{
		"ServerName":        serverName,
		"ServerHost":        serverHost,
		"ServerPort":        serverPort,
		"ServerVersion":     serverVersion,
		"ServerDescription": serverDescription,
		"User":              user,
	}
}

func getIndexPageHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "index.html", getCommonPageData(c))
	}
}

func initializeWebServerRoutes(r *gin.Engine, authController *discordAuthController, serverService *services.ServerService, whitelistService *services.WhitelistService, commandService *services.CommandService) {
	protected := r.Group("/")
	protected.Use(authController.RequireAuth())
	protected.GET("/", getIndexPageHandler())
	protected.GET("/server-info", handleGetServerInfo(serverService))
	protected.GET("/whitelist", handleGetWhitelist(whitelistService))
	protected.POST("/whitelist/player", handleAddNameToWhitelist(whitelistService))
	protected.DELETE("/whitelist/player/:name", handleRemoveNameFromWhitelist(whitelistService))
	protected.GET("/players/:name/kick", handleGetKickPlayerDialog())
	protected.POST("/players/:name/kick", handleKickPlayer(serverService))
	protected.GET("/rcon", handleGetCommandConsole(commandService))
	protected.POST("/commands/execute", handleExecuteRawCommand(commandService))
}

type WebServerOptions struct {
	MinecraftRconClient rcon.CommandExecutor
	AshconClient        ashcon_client.MojangUserNameChecker
	AuthConfig          AuthConfig
}

func InitializeWebServer(options WebServerOptions) (*gin.Engine, error) {
	r := initializeWebServer()
	authController, err := ConfigureDiscordAuth(r, options.AuthConfig)
	if err != nil {
		return nil, err
	}
	serverService := services.NewServerServiceFromRconClient(options.MinecraftRconClient)
	whitelistService := services.NewWhitelistService(options.MinecraftRconClient, options.AshconClient)
	commandService := services.NewCommandServiceFromRconClient(options.MinecraftRconClient)
	initializeWebServerRoutes(r, authController, serverService, whitelistService, commandService)
	return r, nil
}
