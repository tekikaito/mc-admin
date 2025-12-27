package services

import (
	"fmt"
	"mc-admin/internal/clients/rcon"
	"strings"
)

type CommandService struct {
	rconClient rcon.CommandExecutor
}

func NewCommandServiceFromRconClient(rconClient rcon.CommandExecutor) *CommandService {
	return &CommandService{
		rconClient: rconClient,
	}
}

func (s *CommandService) ExecuteRawCommand(command string) (string, error) {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return "", fmt.Errorf("command cannot be empty")
	}

	response, err := s.rconClient.ExecuteCommand(trimmed)
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %w", err)
	}

	return response, nil
}
