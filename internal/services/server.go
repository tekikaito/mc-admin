package services

import (
	"fmt"
	"rcon-web/internal/rcon"
	"rcon-web/internal/utils"
	"strings"
)

type ServerService struct {
	rconClient *rcon.MinecraftRconClient
}

type ServerPlayerInfo struct {
	PlayerNames []string `json:"player_names"`
	OnlineCount int      `json:"online_count"`
	MaxCount    int      `json:"max_count"`
}

func NewServerServiceFromRconClient(rconClient *rcon.MinecraftRconClient) *ServerService {
	return &ServerService{
		rconClient: rconClient,
	}
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
