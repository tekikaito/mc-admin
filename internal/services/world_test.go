package services

import (
	"errors"
	"reflect"
	"testing"
)

func TestWorldService_GetWorldStats(t *testing.T) {
	tests := []struct {
		name          string
		difficultyOut string
		difficultyErr error
		gameTimeOut   string
		gameTimeErr   error
		want          WorldStats
		wantErr       bool
		wantCommands  []string
	}{
		{
			name:          "success",
			difficultyOut: "The difficulty is Hard",
			gameTimeOut:   "The time is 30000",
			want: WorldStats{
				Day:        1,
				DayPhase:   GameTickNoonLabel,
				Difficulty: "hard",
			},
			wantCommands: []string{"difficulty", "time query gametime"},
		},
		{
			name:          "difficulty error",
			difficultyErr: errors.New("boom"),
			wantErr:       true,
			wantCommands:  []string{"difficulty"},
		},
		{
			name:          "gametime error",
			difficultyOut: "The difficulty is Normal",
			gameTimeErr:   errors.New("boom"),
			wantErr:       true,
			wantCommands:  []string{"difficulty", "time query gametime"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeRconClient{
				responses: map[string]struct {
					out string
					err error
				}{
					"difficulty":          {out: tt.difficultyOut, err: tt.difficultyErr},
					"time query gametime": {out: tt.gameTimeOut, err: tt.gameTimeErr},
				},
			}

			svc := NewWorldService(fake)
			got, err := svc.GetWorldStats()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !reflect.DeepEqual(fake.received, tt.wantCommands) {
					t.Fatalf("commands = %v, want %v", fake.received, tt.wantCommands)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("stats = %+v, want %+v", got, tt.want)
			}
			if !reflect.DeepEqual(fake.received, tt.wantCommands) {
				t.Fatalf("commands = %v, want %v", fake.received, tt.wantCommands)
			}
		})
	}
}

func TestWorldService_SetTime(t *testing.T) {
	fake := &fakeRconClient{responses: map[string]struct {
		out string
		err error
	}{
		"time set day": {out: "Set the time to 1000", err: nil},
	}}

	svc := NewWorldService(fake)
	got, err := svc.SetTime("day")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Set the time to 1000" {
		t.Fatalf("SetTime() = %q, want %q", got, "Set the time to 1000")
	}
	if len(fake.received) != 1 || fake.received[0] != "time set day" {
		t.Fatalf("ExecuteCommand called with %v, want %q", fake.received, "time set day")
	}
}

func TestWorldService_GetDaytime(t *testing.T) {
	fake := &fakeRconClient{responses: map[string]struct {
		out string
		err error
	}{
		"time query daytime": {out: "The time is 6000", err: nil},
	}}

	svc := NewWorldService(fake)
	got, err := svc.GetDaytime()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 6000 {
		t.Fatalf("GetDaytime() = %d, want %d", got, 6000)
	}
}

func TestWorldService_GetPhaseFromTicks(t *testing.T) {
	svc := NewWorldService(&fakeRconClient{responses: map[string]struct {
		out string
		err error
	}{}})

	if got := svc.GetPhaseFromTicks(6000); got != GameTickNoonLabel {
		t.Fatalf("GetPhaseFromTicks(6000) = %q, want %q", got, GameTickNoonLabel)
	}
}

func TestWorldService_SetWeather(t *testing.T) {
	tests := []struct {
		name        string
		weather     string
		duration    int
		wantCommand string
		wantErr     bool
	}{
		{name: "no duration", weather: "clear", duration: 0, wantCommand: "weather clear"},
		{name: "with duration", weather: "rain", duration: 30, wantCommand: "weather rain 30"},
		{name: "rcon error", weather: "thunder", duration: 5, wantCommand: "weather thunder 5", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeRconClient{responses: map[string]struct {
				out string
				err error
			}{
				tt.wantCommand: {out: "ok", err: nil},
			}}
			if tt.wantErr {
				fake.responses[tt.wantCommand] = struct {
					out string
					err error
				}{out: "", err: errors.New("boom")}
			}

			svc := NewWorldService(fake)
			_, err := svc.SetWeather(tt.weather, tt.duration)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(fake.received) != 1 || fake.received[0] != tt.wantCommand {
				t.Fatalf("ExecuteCommand called with %v, want %q", fake.received, tt.wantCommand)
			}
		})
	}
}

func TestWorldService_SetDifficultyAndGetDifficulty(t *testing.T) {
	fake := &fakeRconClient{responses: map[string]struct {
		out string
		err error
	}{
		"difficulty easy": {out: "Set difficulty to easy", err: nil},
		"difficulty":      {out: "The difficulty is Easy", err: nil},
	}}

	svc := NewWorldService(fake)
	if _, err := svc.SetDifficulty("easy"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := svc.GetDifficulty()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Easy" {
		t.Fatalf("GetDifficulty() = %q, want %q", got, "Easy")
	}

	wantCommands := []string{"difficulty easy", "difficulty"}
	if !reflect.DeepEqual(fake.received, wantCommands) {
		t.Fatalf("commands = %v, want %v", fake.received, wantCommands)
	}
}

func TestWorldService_SetDifficultyAndGetDifficultyWithNewline(t *testing.T) {
	fake := &fakeRconClient{responses: map[string]struct {
		out string
		err error
	}{
		"difficulty easy": {out: "Set difficulty to easy", err: nil},
		"difficulty":      {out: "The difficulty is Easy\n", err: nil},
	}}

	svc := NewWorldService(fake)
	if _, err := svc.SetDifficulty("easy"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := svc.GetDifficulty()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Easy" {
		t.Fatalf("GetDifficulty() = %q, want %q", got, "Easy")
	}

	wantCommands := []string{"difficulty easy", "difficulty"}
	if !reflect.DeepEqual(fake.received, wantCommands) {
		t.Fatalf("commands = %v, want %v", fake.received, wantCommands)
	}
}

func TestWorldService_SetGameRuleAndGetGameRule(t *testing.T) {
	fake := &fakeRconClient{responses: map[string]struct {
		out string
		err error
	}{
		"gamerule keepInventory true": {out: "Game rule keepInventory is now set to: true", err: nil},
		"gamerule keepInventory":      {out: "keepInventory = true", err: nil},
	}}

	svc := NewWorldService(fake)
	if _, err := svc.SetGameRule("keepInventory", "true"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.GetGameRule("keepInventory"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"gamerule keepInventory true", "gamerule keepInventory"}
	if !reflect.DeepEqual(fake.received, want) {
		t.Fatalf("commands = %v, want %v", fake.received, want)
	}
}

func TestWorldService_SetWorldSpawn(t *testing.T) {
	fake := &fakeRconClient{responses: map[string]struct {
		out string
		err error
	}{
		"setworldspawn 1 2 3": {out: "Set world spawn to (1,2,3)", err: nil},
	}}

	svc := NewWorldService(fake)
	if _, err := svc.SetWorldSpawn(1, 2, 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.received) != 1 || fake.received[0] != "setworldspawn 1 2 3" {
		t.Fatalf("ExecuteCommand called with %v, want %q", fake.received, "setworldspawn 1 2 3")
	}
}

func TestWorldService_WorldBorderCommands(t *testing.T) {
	fake := &fakeRconClient{responses: map[string]struct {
		out string
		err error
	}{
		"worldborder set 128.500000":             {out: "ok", err: nil},
		"worldborder center 10.000000 20.250000": {out: "ok", err: nil},
	}}

	svc := NewWorldService(fake)
	if _, err := svc.SetWorldBorder(128.5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.SetWorldBorderCenter(10, 20.25); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"worldborder set 128.500000", "worldborder center 10.000000 20.250000"}
	if !reflect.DeepEqual(fake.received, want) {
		t.Fatalf("commands = %v, want %v", fake.received, want)
	}
}

func TestWorldService_SayAndTitle(t *testing.T) {
	fake := &fakeRconClient{responses: map[string]struct {
		out string
		err error
	}{
		"say hello":                           {out: "ok", err: nil},
		"title Steve title {\"text\":\"Hi\"}": {out: "ok", err: nil},
	}}

	svc := NewWorldService(fake)
	if _, err := svc.Say("hello"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.Title("Steve", "title", "Hi"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"say hello", "title Steve title {\"text\":\"Hi\"}"}
	if !reflect.DeepEqual(fake.received, want) {
		t.Fatalf("commands = %v, want %v", fake.received, want)
	}
}

func TestWorldService_SaveToggleAutoSaveStop(t *testing.T) {
	fake := &fakeRconClient{responses: map[string]struct {
		out string
		err error
	}{
		"save-all": {out: "Saved the game", err: nil},
		"save-on":  {out: "Turned on auto-saving", err: nil},
		"save-off": {out: "Turned off auto-saving", err: nil},
		"stop":     {out: "Stopping the server", err: nil},
	}}

	svc := NewWorldService(fake)
	if _, err := svc.Save(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.ToggleAutoSave(true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.ToggleAutoSave(false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.Stop(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"save-all", "save-on", "save-off", "stop"}
	if !reflect.DeepEqual(fake.received, want) {
		t.Fatalf("commands = %v, want %v", fake.received, want)
	}
}
