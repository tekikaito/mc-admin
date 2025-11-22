package services

import (
	"fmt"
	"rcon-web/internal/rcon"
	"rcon-web/internal/utils"
	"strings"
)

type WhitelistService struct {
	rconClient *rcon.MinecraftRconClient
}

func NewWhitelistServiceFromRconClient(rconClient *rcon.MinecraftRconClient) *WhitelistService {
	return &WhitelistService{
		rconClient: rconClient,
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
