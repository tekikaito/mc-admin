package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

const (
	sessionName          = "mc-admin-session"
	sessionKeyState      = "oauth_state"
	sessionKeyUserID     = "user_id"
	sessionKeyUserName   = "user_name"
	sessionKeyUserAvatar = "user_avatar"
	userContextKey       = "authenticatedUser"
)

// AuthConfig describes the Discord OAuth configuration required by the server.
type AuthConfig struct {
	Enabled        bool
	ClientID       string
	ClientSecret   string
	RedirectURI    string
	SessionSecret  string
	AllowedUserIDs []string
}

// AuthenticatedUser represents the information persisted in the session once Discord authentication succeeds.
type AuthenticatedUser struct {
	ID        string
	Username  string
	AvatarURL string
}

type discordAuthController struct {
	enabled        bool
	oauthConfig    *oauth2.Config
	allowedUserIDs map[string]struct{}
}

var discordOAuthEndpoint = oauth2.Endpoint{
	AuthURL:  "https://discord.com/api/oauth2/authorize",
	TokenURL: "https://discord.com/api/oauth2/token",
}

// ConfigureDiscordAuth wires the session middleware, registers auth routes, and returns a helper for downstream middleware.
func ConfigureDiscordAuth(r *gin.Engine, cfg AuthConfig) (*discordAuthController, error) {
	if err := validateAuthConfig(cfg); err != nil {
		return nil, err
	}

	if !cfg.Enabled {
		return &discordAuthController{enabled: false}, nil
	}

	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((24 * time.Hour).Seconds()),
		Secure:   strings.HasPrefix(strings.ToLower(cfg.RedirectURI), "https"),
	})
	r.Use(sessions.Sessions(sessionName, store))

	allowedSet := make(map[string]struct{})
	for _, id := range cfg.AllowedUserIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		allowedSet[trimmed] = struct{}{}
	}

	controller := &discordAuthController{
		enabled: true,
		oauthConfig: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURI,
			Scopes:       []string{"identify"},
			Endpoint:     discordOAuthEndpoint,
		},
		allowedUserIDs: allowedSet,
	}

	controller.registerRoutes(r)
	return controller, nil
}

func validateAuthConfig(cfg AuthConfig) error {
	if !cfg.Enabled {
		return nil
	}
	switch {
	case strings.TrimSpace(cfg.ClientID) == "":
		return errors.New("missing ClientID")
	case strings.TrimSpace(cfg.ClientSecret) == "":
		return errors.New("missing ClientSecret")
	case strings.TrimSpace(cfg.RedirectURI) == "":
		return errors.New("missing RedirectURI")
	case strings.TrimSpace(cfg.SessionSecret) == "":
		return errors.New("missing SessionSecret")
	}
	return nil
}

func (a *discordAuthController) registerRoutes(r *gin.Engine) {
	if !a.enabled {
		return
	}
	authGroup := r.Group("/auth")
	authGroup.GET("/login", a.renderLoginPage())
	authGroup.GET("/discord/start", a.startDiscordFlow())
	authGroup.GET("/discord/callback", a.handleDiscordCallback())
	authGroup.POST("/logout", a.handleLogout())
}

func (a *discordAuthController) renderLoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user := currentUserFromSession(c); user != nil {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.HTML(http.StatusOK, "login.html", gin.H{})
	}
}

func (a *discordAuthController) startDiscordFlow() gin.HandlerFunc {
	return func(c *gin.Context) {
		state, err := generateOAuthState()
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to initiate Discord login")
			return
		}
		session := sessions.Default(c)
		session.Set(sessionKeyState, state)
		if err := session.Save(); err != nil {
			c.String(http.StatusInternalServerError, "failed to persist session state")
			return
		}
		authURL := a.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
		c.Redirect(http.StatusFound, authURL)
	}
}

func (a *discordAuthController) handleDiscordCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		savedState, _ := session.Get(sessionKeyState).(string)
		session.Delete(sessionKeyState)
		_ = session.Save()

		if savedState == "" || c.Query("state") != savedState {
			c.String(http.StatusBadRequest, "invalid OAuth state")
			return
		}

		code := c.Query("code")
		if code == "" {
			c.String(http.StatusBadRequest, "missing OAuth code")
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		token, err := a.oauthConfig.Exchange(ctx, code)
		if err != nil {
			c.String(http.StatusBadRequest, "failed to exchange Discord code")
			return
		}

		discordUser, err := fetchDiscordUser(ctx, token.AccessToken)
		if err != nil {
			c.String(http.StatusBadGateway, "failed to fetch Discord profile")
			return
		}

		if len(a.allowedUserIDs) > 0 {
			if _, ok := a.allowedUserIDs[discordUser.ID]; !ok {
				c.String(http.StatusForbidden, "Discord account is not authorized")
				return
			}
		}

		session.Set(sessionKeyUserID, discordUser.ID)
		session.Set(sessionKeyUserName, discordUser.DisplayName())
		session.Set(sessionKeyUserAvatar, discordUser.AvatarURL())
		if err := session.Save(); err != nil {
			c.String(http.StatusInternalServerError, "failed to update session")
			return
		}

		c.Redirect(http.StatusFound, "/")
	}
}

func (a *discordAuthController) handleLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		_ = session.Save()
		c.Redirect(http.StatusFound, "/auth/login")
	}
}

// RequireAuth ensures the user is authenticated before allowing the request to proceed.
func (a *discordAuthController) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.enabled {
			c.Next()
			return
		}
		user := currentUserFromSession(c)
		if user == nil {
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}
		c.Set(userContextKey, user)
		c.Next()
	}
}

func generateOAuthState() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

type discordUserResponse struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	GlobalName    string `json:"global_name"`
	Avatar        string `json:"avatar"`
	Discriminator string `json:"discriminator"`
}

func (u discordUserResponse) DisplayName() string {
	if strings.TrimSpace(u.GlobalName) != "" {
		return u.GlobalName
	}
	if strings.TrimSpace(u.Username) != "" {
		return u.Username
	}
	return "Discord User"
}

func (u discordUserResponse) AvatarURL() string {
	if strings.TrimSpace(u.Avatar) == "" {
		return "https://cdn.discordapp.com/embed/avatars/0.png"
	}
	format := "png"
	if strings.HasPrefix(u.Avatar, "a_") {
		format = "gif"
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.%s?size=128", u.ID, u.Avatar, format)
}

func fetchDiscordUser(ctx context.Context, accessToken string) (discordUserResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://discord.com/api/users/@me", nil)
	if err != nil {
		return discordUserResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return discordUserResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return discordUserResponse{}, fmt.Errorf("discord API returned %d", resp.StatusCode)
	}

	var payload discordUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return discordUserResponse{}, err
	}
	return payload, nil
}

func currentUserFromSession(c *gin.Context) *AuthenticatedUser {
	session := sessions.Default(c)
	id, _ := session.Get(sessionKeyUserID).(string)
	if strings.TrimSpace(id) == "" {
		return nil
	}
	name, _ := session.Get(sessionKeyUserName).(string)
	avatar, _ := session.Get(sessionKeyUserAvatar).(string)
	return &AuthenticatedUser{ID: id, Username: name, AvatarURL: avatar}
}

// CurrentUser exposes the authenticated user that the RequireAuth middleware attached to the context.
func CurrentUser(c *gin.Context) *AuthenticatedUser {
	value, exists := c.Get(userContextKey)
	if !exists {
		return nil
	}
	if user, ok := value.(*AuthenticatedUser); ok {
		return user
	}
	return nil
}
