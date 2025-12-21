package api

import (
	ashcon_client "mc-admin/internal/clients"
	"mc-admin/internal/files"
	"mc-admin/internal/rcon"
	"mc-admin/internal/services"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
		// assetVersion returns a cache-busting version for a /static/* web path.
		// Falls back to "0" if the file can't be stat'ed.
		"assetVersion": func(webPath string) string {
			rel := strings.TrimPrefix(webPath, "/")
			rel = strings.TrimPrefix(rel, "static/")
			fi, err := os.Stat(filepath.Join("static", rel))
			if err != nil {
				return "0"
			}
			return strconv.FormatInt(fi.ModTime().Unix(), 10)
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
		"ActiveModule":      "world",
	}
}

func getIndexPageHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "index.html", getCommonPageData(c))
	}
}

type WebServerParts struct {
	AuthController   *discordAuthController
	ServerService    *services.ServerService
	WhitelistService *services.WhitelistService
	CommandService   *services.CommandService
	FileService      *services.FileService
	WorldService     *services.WorldService
}

func initializeWebServerRoutes(r *gin.Engine, parts WebServerParts) {
	protected := r.Group("/")
	protected.Use(parts.AuthController.RequireAuth())
	protected.GET("/", getIndexPageHandler())
	protected.GET("/server-info", handleGetServerInfo(parts.ServerService))
	protected.GET("/whitelist", handleGetWhitelist(parts.WhitelistService))
	protected.POST("/whitelist/toggle", handleToggleWhitelist(parts.WhitelistService))
	protected.POST("/whitelist/player", handleAddNameToWhitelist(parts.WhitelistService))
	protected.DELETE("/whitelist/player/:name", handleRemoveNameFromWhitelist(parts.WhitelistService))
	protected.GET("/world/stats", handleGetWorldStats(parts.WorldService))
	protected.GET("/world/clock", handleGetClock(parts.WorldService))
	protected.GET("/world/clock/edit", handleGetClockEdit(parts.WorldService))
	protected.POST("/world/time", handleSetTime(parts.WorldService))
	protected.GET("/players/:name/kick", handleGetKickPlayerDialog())
	protected.POST("/players/:name/kick", handleKickPlayer(parts.ServerService))
	protected.GET("/rcon", handleGetCommandConsole())
	protected.POST("/commands/execute", handleExecuteRawCommand(parts.CommandService))
	protected.GET("/files", handleGetFiles(parts.FileService))
	protected.GET("/files/content", handleGetFileContent(parts.FileService))
	protected.GET("/files/download", handleDownloadFile(parts.FileService))
	protected.POST("/files/create", handleCreateFile(parts.FileService))
	protected.POST("/files/upload", handleUploadFile(parts.FileService))
	protected.DELETE("/files/delete", handleDeleteFile(parts.FileService))
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
	worldService := services.NewWorldService(options.MinecraftRconClient)

	parts := WebServerParts{
		AuthController:   authController,
		ServerService:    serverService,
		WhitelistService: whitelistService,
		CommandService:   commandService,
		FileService:      fileService,
		WorldService:     worldService,
	}

	initializeWebServerRoutes(r, parts)
	return r, nil
}
