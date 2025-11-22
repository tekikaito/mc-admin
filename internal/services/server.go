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
		return ServerPlayerInfo{}, err
	}

	var info ServerPlayerInfo
	var commaSeparatedNames string

	fmt.Sscanf(response, "There are %d of a max of %d players online", &info.OnlineCount, &info.MaxCount)

	if parts := strings.SplitN(response, ":", 2); len(parts) == 2 {
		commaSeparatedNames = strings.TrimSpace(parts[1])
	}

	if info.OnlineCount == 0 || commaSeparatedNames == "" {
		return info, nil
	}
	playerNames := utils.SplitAndTrim(commaSeparatedNames, ",")
	defer func() {
		if len(playerNames) > 0 {
			info.PlayerNames = append([]string(nil), playerNames...)
		}
	}()
	info.PlayerNames = []string{}

	if info.OnlineCount == 1 {
		info.PlayerNames = append(info.PlayerNames, commaSeparatedNames)
	} else if info.OnlineCount > 1 {
		info.PlayerNames = append(info.PlayerNames, utils.SplitAndTrim(commaSeparatedNames, ",")...)
	}

	return info, nil
}
