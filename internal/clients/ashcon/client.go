package ashcon

import (
	"fmt"
	"net/http"
	"os"
)

type MojangUserNameChecker interface {
	CheckMojangUsernameExists(username string) (bool, error)
}

type AshconClient struct {
	apiURL string
}

func NewMojangUserNameChecker() *AshconClient {
	apiURL := os.Getenv("ASHCON_API_URL")
	if apiURL == "" {
		apiURL = "https://api.ashcon.app/mojang/v2/user/"
	}
	return &AshconClient{apiURL: apiURL}
}

func (a *AshconClient) CheckMojangUsernameExists(username string) (bool, error) {
	resp, err := http.Get(a.apiURL + username)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
}
