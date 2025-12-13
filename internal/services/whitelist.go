package services

import (
	"fmt"
	ashcon_client "mc-admin/internal/clients"
	"mc-admin/internal/rcon"
	"mc-admin/internal/utils"
	"strings"
)

type WhitelistFileSystemAccessor interface {
	ReadFile(path string) (string, error)
}

type WhitelistService struct {
	rconClient           rcon.CommandExecutor
	mojangClient         ashcon_client.MojangUserNameChecker
	minecraftFilesClient WhitelistFileSystemAccessor
	mojangCheckEnabled   bool
}

type WhitelistInfo struct {
	PlayerNames []string `json:"player_names"`
	Enabled     bool     `json:"enabled"`
}

func NewWhitelistService(rconClient rcon.CommandExecutor, mojangClient ashcon_client.MojangUserNameChecker, minecraftFilesClient WhitelistFileSystemAccessor) *WhitelistService {
	if mojangClient != nil {
		return &WhitelistService{
			rconClient:           rconClient,
			mojangCheckEnabled:   true,
			mojangClient:         mojangClient,
			minecraftFilesClient: minecraftFilesClient,
		}
	}
	return &WhitelistService{
		rconClient:           rconClient,
		mojangClient:         nil,
		mojangCheckEnabled:   false,
		minecraftFilesClient: minecraftFilesClient,
	}
}

func getWhitelistEnabledStatus(fileClient WhitelistFileSystemAccessor) (bool, error) {
	content, err := fileClient.ReadFile("server.properties")
	if err != nil {
		return false, fmt.Errorf("failed to get absolute path for server.properties: %w", err)
	}
	enabled := false
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "white-list=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[1]) == "true" {
				enabled = true
			}
			break
		}
	}
	return enabled, nil
}

func getWhitelistPlayerNames(rconClient rcon.CommandExecutor) ([]string, error) {
	response, err := rconClient.ExecuteCommand("whitelist list")
	if err != nil {
		return []string{}, fmt.Errorf("failed to execute whitelist list command: %w", err)
	}

	// Response format: "There are X whitelisted players: player1, player2, ..."
	parts := strings.SplitN(response, ":", 2)
	if len(parts) != 2 {
		return []string{}, fmt.Errorf("unexpected response format: %s", response)
	}

	commaSeparatedNames := strings.TrimSpace(parts[1])
	if commaSeparatedNames == "" {
		return []string{}, nil
	}

	playerNames := utils.SplitAndTrim(commaSeparatedNames, ",")
	return playerNames, nil
}

func (s *WhitelistService) GetWhitelistInfo() (WhitelistInfo, error) {
	enabled, err := getWhitelistEnabledStatus(s.minecraftFilesClient)
	if err != nil {
		return WhitelistInfo{}, fmt.Errorf("failed to get whitelist enabled status: %w", err)
	}

	playerNames, err := getWhitelistPlayerNames(s.rconClient)
	if err != nil {
		return WhitelistInfo{}, fmt.Errorf("failed to get whitelist player names: %w", err)
	}

	return WhitelistInfo{
		PlayerNames: playerNames,
		Enabled:     enabled,
	}, nil
}

func (s *WhitelistService) EnableWhitelist() error {
	_, err := s.rconClient.ExecuteCommand("whitelist on")
	if err != nil {
		return fmt.Errorf("failed to enable whitelist: %w", err)
	}
	return nil
}

func (s *WhitelistService) DisableWhitelist() error {
	_, err := s.rconClient.ExecuteCommand("whitelist off")
	if err != nil {
		return fmt.Errorf("failed to disable whitelist: %w", err)
	}
	return nil
}

func (s *WhitelistService) RemoveNameFromWhitelist(name string) error {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return fmt.Errorf("name cannot be empty")
	}

	command := fmt.Sprintf("whitelist remove %s", trimmedName)
	_, err := s.rconClient.ExecuteCommand(command)
	if err != nil {
		return fmt.Errorf("failed to remove name from whitelist: %w", err)
	}

	return nil
}

func (s *WhitelistService) AddNameToWhitelist(name string) error {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Check if the name is already whitelisted
	currentWhitelist, err := getWhitelistPlayerNames(s.rconClient)
	if err != nil {
		return fmt.Errorf("failed to get current whitelist: %w", err)
	}
	for _, existingName := range currentWhitelist {
		if strings.EqualFold(existingName, trimmedName) {
			return fmt.Errorf("name '%s' is already whitelisted", trimmedName)
		}
	}

	exists := true
	if s.mojangCheckEnabled {
		exists, err := s.mojangClient.CheckMojangUsernameExists(trimmedName)
		fmt.Printf("exists: %v, err: %v\n", exists, err)
		if err != nil {
			return fmt.Errorf("failed to verify if name exists: %w", err)
		}
	}
	if !exists {
		return fmt.Errorf("name '%s' does not exist", trimmedName)
	}

	command := fmt.Sprintf("whitelist add %s", trimmedName)
	_, err = s.rconClient.ExecuteCommand(command)
	if err != nil {
		return fmt.Errorf("failed to add name to whitelist: %w", err)
	}

	return nil
}
