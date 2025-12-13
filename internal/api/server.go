package api

import (
	ashcon_client "mc-admin/internal/clients"
	"mc-admin/internal/files"
	"mc-admin/internal/rcon"
	"mc-admin/internal/services"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/gin-gonic/gin"
)

func initializeWebServer() *gin.Engine {
	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"joinPath": func(base, child string) string {
			return filepath.Join(base, child)
		},
		"dir": func(path string) string {
			return filepath.Dir(path)
		},
	})
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
	minecraftDataDir := os.Getenv("MINECRAFT_DATA_DIR")

	return gin.H{
		"ServerName":        serverName,
		"ServerHost":        serverHost,
		"ServerPort":        serverPort,
		"ServerVersion":     serverVersion,
		"ServerDescription": serverDescription,
		"User":              user,
		"FilesEnabled":      minecraftDataDir != "",
	}
}

func getIndexPageHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "index.html", getCommonPageData(c))
	}
}

func initializeWebServerRoutes(r *gin.Engine, authController *discordAuthController, serverService *services.ServerService, whitelistService *services.WhitelistService, commandService *services.CommandService, fileService *services.FileService) {
	protected := r.Group("/")
	protected.Use(authController.RequireAuth())
	protected.GET("/", getIndexPageHandler())
	protected.GET("/server-info", handleGetServerInfo(serverService))
	protected.GET("/whitelist", handleGetWhitelist(whitelistService))
	protected.POST("/whitelist/toggle", handleToggleWhitelist(whitelistService))
	protected.POST("/whitelist/player", handleAddNameToWhitelist(whitelistService))
	protected.DELETE("/whitelist/player/:name", handleRemoveNameFromWhitelist(whitelistService))
	protected.GET("/players/:name/kick", handleGetKickPlayerDialog())
	protected.POST("/players/:name/kick", handleKickPlayer(serverService))
	protected.GET("/rcon", handleGetCommandConsole(commandService))
	protected.POST("/commands/execute", handleExecuteRawCommand(commandService))

	protected.GET("/files", handleGetFiles(fileService))
	protected.GET("/files/content", handleGetFileContent(fileService))
	protected.GET("/files/download", handleDownloadFile(fileService))
	protected.POST("/files/create", handleCreateFile(fileService))
	protected.POST("/files/upload", handleUploadFile(fileService))
	protected.DELETE("/files/delete", handleDeleteFile(fileService))
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
	commandService := services.NewCommandServiceFromRconClient(options.MinecraftRconClient)

	maxDisplaySizeStr := os.Getenv("MAX_FILE_DISPLAY_SIZE")
	var maxDisplaySize int64
	if maxDisplaySizeStr != "" {
		maxDisplaySize, _ = strconv.ParseInt(maxDisplaySizeStr, 10, 64)
	}
	minecraftDataDir := os.Getenv("MINECRAFT_DATA_DIR")
	var fileClient files.MinecraftFilesClient
	if minecraftDataDir != "" {
		fileClient = files.NewMinecraftFilesClient(minecraftDataDir, maxDisplaySize)
	}

	whitelistService := services.NewWhitelistService(options.MinecraftRconClient, options.AshconClient, &fileClient)
	fileService := services.NewFileService(&fileClient)
	initializeWebServerRoutes(r, authController, serverService, whitelistService, commandService, fileService)
	return r, nil
}
