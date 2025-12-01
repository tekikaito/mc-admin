package services

import (
	"fmt"
	ashcon_client "mc-admin/internal/clients"
	"mc-admin/internal/rcon"
	"mc-admin/internal/utils"
	"strings"
)

type WhitelistService struct {
	rconClient    rcon.CommandExecutor
	mojangChecker ashcon_client.MojangUserNameChecker
}

func NewWhitelistService(rconClient rcon.CommandExecutor, mojangChecker ashcon_client.MojangUserNameChecker) *WhitelistService {
	return &WhitelistService{
		rconClient:    rconClient,
		mojangChecker: mojangChecker,
	}
}

func (s *WhitelistService) GetWhitelist() ([]string, error) {
	response, err := s.rconClient.ExecuteCommand("whitelist list")
	if err != nil {
		return nil, fmt.Errorf("failed to execute whitelist list command: %w", err)
	}

	// Response format: "There are X whitelisted players: player1, player2, ..."
	parts := strings.SplitN(response, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected response format: %s", response)
	}

	commaSeparatedNames := strings.TrimSpace(parts[1])
	if commaSeparatedNames == "" {
		return []string{}, nil
	}

	playerNames := utils.SplitAndTrim(commaSeparatedNames, ",")
	return playerNames, nil
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
	currentWhitelist, err := s.GetWhitelist()
	if err != nil {
		return fmt.Errorf("failed to get current whitelist: %w", err)
	}
	for _, existingName := range currentWhitelist {
		if strings.EqualFold(existingName, trimmedName) {
			return fmt.Errorf("name '%s' is already whitelisted", trimmedName)
		}
	}

	exists, err := s.mojangChecker.CheckMojangUsernameExists(trimmedName)
	fmt.Printf("exists: %v, err: %v\n", exists, err)
	if err != nil {
		return fmt.Errorf("failed to verify if name exists: %w", err)
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
