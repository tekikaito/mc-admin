package services

import (
	"fmt"
	"rcon-web/internal/rcon"
	"rcon-web/internal/utils"
	"strings"
)

type ServerService struct {
	rconClient rcon.CommandExecutor
}

type ServerPlayerInfo struct {
	PlayerNames []string `json:"player_names"`
	OnlineCount int      `json:"online_count"`
	MaxCount    int      `json:"max_count"`
}

func NewServerServiceFromRconClient(rconClient rcon.CommandExecutor) *ServerService {
	return &ServerService{rconClient}
}

func (s *ServerService) GetServerPlayerInfo() (ServerPlayerInfo, error) {
	response, err := s.rconClient.ExecuteCommand("list")
	if err != nil {
		return ServerPlayerInfo{}, fmt.Errorf("failed to execute list command: %w", err)
	}

	var info ServerPlayerInfo
	info.PlayerNames = []string{}

	// Parse player counts from response
	n, err := fmt.Sscanf(response, "There are %d of a max of %d players online", &info.OnlineCount, &info.MaxCount)
	if err != nil || n != 2 {
		return ServerPlayerInfo{}, fmt.Errorf("failed to parse player count from response '%s': %w", response, err)
	}

	// Extract player names if any
	if info.OnlineCount > 0 {
		if parts := strings.SplitN(response, ":", 2); len(parts) == 2 {
			commaSeparatedNames := strings.TrimSpace(parts[1])
			if commaSeparatedNames != "" {
				info.PlayerNames = utils.SplitAndTrim(commaSeparatedNames, ",")
			}
		}
	}

	return info, nil
}

func (s *ServerService) KickPlayerByName(name string, reason string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("player name cannot be empty")
	}

	command := fmt.Sprintf("kick %s", trimmed)
	if r := strings.TrimSpace(reason); r != "" {
		command = fmt.Sprintf("%s %s", command, r)
	}

	_, err := s.rconClient.ExecuteCommand(command)
	if err != nil {
		return fmt.Errorf("failed to kick player '%s': %w", trimmed, err)
	}

	return nil
}
