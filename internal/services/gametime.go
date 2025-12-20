package services

import (
	"fmt"
	"mc-admin/internal/rcon"
	"strconv"
)

type GametimeService struct {
	rconClient rcon.CommandExecutor
}

type Gameticks int64

const timeQueryCommandResponsePrefix = "The time is "

const (
	GameTickDay      Gameticks = 1000
	GameTickNoon     Gameticks = 6000
	GameTickNight    Gameticks = 13000
	GameTickMidnight Gameticks = 18000
)

const (
	GameTickDayLabel      = "day"
	GameTickNoonLabel     = "noon"
	GameTickNightLabel    = "night"
	GameTickMidnightLabel = "midnight"
)

const ticksPerMinecraftDay = 24000

func (d Gameticks) DayPhaseString() string {
	dayPhaseTicks := d.DayPhaseTicks()
	switch {
	case dayPhaseTicks >= GameTickDay && dayPhaseTicks < GameTickNoon:
		return GameTickDayLabel
	case dayPhaseTicks >= GameTickNoon && dayPhaseTicks < GameTickNight:
		return GameTickNoonLabel
	case dayPhaseTicks >= GameTickNight && dayPhaseTicks < GameTickMidnight:
		return GameTickNightLabel
	default:
		return GameTickMidnightLabel
	}
}

func (d Gameticks) Day() int {
	return int(d) / ticksPerMinecraftDay
}

func (d Gameticks) DayPhaseTicks() Gameticks {
	return Gameticks(int(d) % ticksPerMinecraftDay)
}

func NewGametimeService(rconClient rcon.CommandExecutor) GametimeService {
	return GametimeService{rconClient: rconClient}
}

// SetTime sets the time in the world
// time can be: day, night, noon, midnight, or a specific tick value
func (s *GametimeService) SetTime(time string) (string, error) {
	cmd := fmt.Sprintf("time set %s", time)
	return s.rconClient.ExecuteCommand(cmd)
}

// AddTime adds time to the world clock
func (s *GametimeService) AddTime(ticks Gameticks) (string, error) {
	cmd := fmt.Sprintf("time add %d", ticks)
	return s.rconClient.ExecuteCommand(cmd)
}

func (s *GametimeService) GetDayTime() (Gameticks, error) {
	daytimeResult, err := s.rconClient.ExecuteCommand("time query daytime")
	if err != nil {
		return 0, err
	}

	trimmedTime := daytimeResult[len(timeQueryCommandResponsePrefix):]
	actual, err := strconv.Atoi(trimmedTime)
	return Gameticks(actual), err
}

func (s *GametimeService) GetGameTime() (Gameticks, error) {
	gametimeResult, err := s.rconClient.ExecuteCommand("time query gametime")
	if err != nil {
		return 0, err
	}

	trimmedTime := gametimeResult[len(timeQueryCommandResponsePrefix):]
	actual, err := strconv.Atoi(trimmedTime)
	return Gameticks(actual), err
}

func (s *GametimeService) GetGameDay() (int, error) {
	dayResult, err := s.rconClient.ExecuteCommand("time query day")
	if err != nil {
		return 0, err
	}

	trimmedDay := dayResult[len(timeQueryCommandResponsePrefix):]
	actual, err := strconv.Atoi(trimmedDay)
	return actual, err
}
