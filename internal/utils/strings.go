package utils

import (
	"fmt"
	"net/http"
	"strings"
)

func SplitAndTrim(commaSeperatedElements string, seperator string) []string {
	var element []string
	for _, name := range strings.Split(commaSeperatedElements, seperator) {
		element = append(element, strings.TrimSpace(name))
	}
	return element
}

// Check if the name exists by fetching https://api.ashcon.app/mojang/v2/user/<name>
func CheckMojangUsernameExists(username string) (bool, error) {
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
