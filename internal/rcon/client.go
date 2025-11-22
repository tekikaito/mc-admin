package rcon

import (
	"fmt"
	"log"
	"rcon-web/internal/utils"
	"strings"

	"github.com/gorcon/rcon"
)

type ServerPlayerInfo struct {
	PlayerNames []string `json:"player_names"`
	OnlineCount int      `json:"online_count"`
	MaxCount    int      `json:"max_count"`
}

type MinecraftRconClient struct {
	Host     string
	Port     string
	Password string
}

func NewMinecraftRconClient(host, port, password string) *MinecraftRconClient {
	return &MinecraftRconClient{
		Host:     host,
		Port:     port,
		Password: password,
	}
}

func (c *MinecraftRconClient) GetServerPlayerInfo() ServerPlayerInfo {
	connectionString := fmt.Sprintf("%s:%s", c.Host, c.Port)
	fmt.Println("Connecting to RCON server at", connectionString)
	
	conn, err := rcon.Dial(connectionString, c.Password)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	response, err := conn.Execute("list")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)

	var info ServerPlayerInfo
	var commaSeparatedNames string

	fmt.Sscanf(response, "There are %d of a max of %d players online", &info.OnlineCount, &info.MaxCount)

	if parts := strings.SplitN(response, ":", 2); len(parts) == 2 {
		commaSeparatedNames = strings.TrimSpace(parts[1])
	}

	if info.OnlineCount == 0 || commaSeparatedNames == "" {
		return info
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

	return info
}