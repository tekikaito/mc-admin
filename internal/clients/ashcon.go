package ashcon_client

import (
	"fmt"
	"net/http"
)

type MojangUserNameChecker interface {
	CheckMojangUsernameExists(username string) (bool, error)
}

type AshconClient struct{}

func NewMojangUserNameChecker() *AshconClient {
	return &AshconClient{}
}

func (a *AshconClient) CheckMojangUsernameExists(username string) (bool, error) {
	resp, err := http.Get("https://api.ashcon.app/mojang/v2/user/" + username)
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
