package services

import (
	"fmt"
	"mc-admin/internal/rcon"
)

// WorldService handles common Minecraft world commands
type WorldService struct {
	rconClient      rcon.CommandExecutor
	gametimeService GametimeService
}

type WorldStats struct {
	Day        int
	DayPhase   string
	Difficulty string
}

// NewWorldService creates a new WorldService instance
func NewWorldService(rconClient rcon.CommandExecutor) *WorldService {
	return &WorldService{
		rconClient:      rconClient,
		gametimeService: NewGametimeService(rconClient),
	}
}

func (s *WorldService) GetWorldStats() (WorldStats, error) {
	difficultyResult, err := s.GetDifficulty()
	if err != nil {
		return WorldStats{}, err
	}
	tickTime, err := s.gametimeService.GetGameTime()
	if err != nil {
		return WorldStats{}, err
	}
	day := tickTime.Day()
	dayPhase := tickTime.DayPhaseString()
	stats := WorldStats{
		Day:        day,
		DayPhase:   dayPhase,
		Difficulty: difficultyResult,
	}
	return stats, nil

}

// SetTime sets the world time
// time can be: day, night, noon, midnight, or a specific tick value
func (s *WorldService) SetTime(time string) (string, error) {
	return s.gametimeService.SetTime(time)
}

// SetWeather sets the weather in the world
// weather can be: clear, rain, thunder
// duration is optional and specifies the duration in seconds
func (s *WorldService) SetWeather(weather string, duration int) (string, error) {
	cmd := fmt.Sprintf("weather %s", weather)
	if duration > 0 {
		cmd = fmt.Sprintf("weather %s %d", weather, duration)
	}
	return s.rconClient.ExecuteCommand(cmd)
}

// SetDifficulty sets the game difficulty
// difficulty can be: peaceful, easy, normal, hard
func (s *WorldService) SetDifficulty(difficulty string) (string, error) {
	cmd := fmt.Sprintf("difficulty %s", difficulty)
	return s.rconClient.ExecuteCommand(cmd)
}

func (s *WorldService) GetDifficulty() (string, error) {
	difficultyResp, err := s.rconClient.ExecuteCommand("difficulty")
	if err != nil {
		return "", err
	}

	actualDifficulty := difficultyResp[len("The difficulty is "):]
	return actualDifficulty, nil
}

// SetGameRule sets a game rule value
func (s *WorldService) SetGameRule(rule string, value string) (string, error) {
	cmd := fmt.Sprintf("gamerule %s %s", rule, value)
	return s.rconClient.ExecuteCommand(cmd)
}

// GetGameRule gets the current value of a game rule
func (s *WorldService) GetGameRule(rule string) (string, error) {
	cmd := fmt.Sprintf("gamerule %s", rule)
	return s.rconClient.ExecuteCommand(cmd)
}

// SetWorldSpawn sets the world spawn point
func (s *WorldService) SetWorldSpawn(x, y, z int) (string, error) {
	cmd := fmt.Sprintf("setworldspawn %d %d %d", x, y, z)
	return s.rconClient.ExecuteCommand(cmd)
}

// SetWorldBorder sets the world border size
func (s *WorldService) SetWorldBorder(size float64) (string, error) {
	cmd := fmt.Sprintf("worldborder set %f", size)
	return s.rconClient.ExecuteCommand(cmd)
}

// SetWorldBorderCenter sets the world border center
func (s *WorldService) SetWorldBorderCenter(x, z float64) (string, error) {
	cmd := fmt.Sprintf("worldborder center %f %f", x, z)
	return s.rconClient.ExecuteCommand(cmd)
}

// Say broadcasts a message to all players
func (s *WorldService) Say(message string) (string, error) {
	cmd := fmt.Sprintf("say %s", message)
	return s.rconClient.ExecuteCommand(cmd)
}

// Title displays a title to a player or all players
func (s *WorldService) Title(player string, titleType string, text string) (string, error) {
	cmd := fmt.Sprintf("title %s %s {\"text\":\"%s\"}", player, titleType, text)
	return s.rconClient.ExecuteCommand(cmd)
}

// Save saves the world
func (s *WorldService) Save() (string, error) {
	return s.rconClient.ExecuteCommand("save-all")
}

// ToggleAutoSave enables or disables auto-save
func (s *WorldService) ToggleAutoSave(enable bool) (string, error) {
	if enable {
		return s.rconClient.ExecuteCommand("save-on")
	}
	return s.rconClient.ExecuteCommand("save-off")
}

// Stop stops the server
func (s *WorldService) Stop() (string, error) {
	return s.rconClient.ExecuteCommand("stop")
}
