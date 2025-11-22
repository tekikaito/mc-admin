package rcon

import (
	"fmt"

	"github.com/gorcon/rcon"
)

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

func getConnectionString(c *MinecraftRconClient) string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func (c *MinecraftRconClient) ExecuteCommand(command string) (string, error) {
	connectionString := getConnectionString(c)
	conn, err := rcon.Dial(connectionString, c.Password)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	response, err := conn.Execute(command)
	if err != nil {
		return "", err
	}

	return response, nil
}
