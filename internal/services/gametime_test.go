package services

import (
	"errors"
	"testing"
)

func TestGameticks_DayPhaseString(t *testing.T) {
	tests := []struct {
		name     string
		ticks    Gameticks
		expected string
	}{
		// Day phase (1000-5999)
		{name: "day start", ticks: 1000, expected: GameTickDayLabel},
		{name: "day middle", ticks: 3000, expected: GameTickDayLabel},
		{name: "day end", ticks: 5999, expected: GameTickDayLabel},
		// Noon phase (6000-12999)
		{name: "noon start", ticks: 6000, expected: GameTickNoonLabel},
		{name: "noon middle", ticks: 9000, expected: GameTickNoonLabel},
		{name: "noon end", ticks: 12999, expected: GameTickNoonLabel},
		// Night phase (13000-17999)
		{name: "night start", ticks: 13000, expected: GameTickNightLabel},
		{name: "night middle", ticks: 15000, expected: GameTickNightLabel},
		{name: "night end", ticks: 17999, expected: GameTickNightLabel},
		// Midnight phase (18000-23999 and 0-999)
		{name: "midnight start", ticks: 18000, expected: GameTickMidnightLabel},
		{name: "midnight middle", ticks: 21000, expected: GameTickMidnightLabel},
		{name: "midnight end", ticks: 23999, expected: GameTickMidnightLabel},
		{name: "midnight wrap 0", ticks: 0, expected: GameTickMidnightLabel},
		{name: "midnight wrap 500", ticks: 500, expected: GameTickMidnightLabel},
		{name: "midnight wrap 999", ticks: 999, expected: GameTickMidnightLabel},
		// Multi-day ticks should wrap correctly
		{name: "day 2 morning", ticks: 24000 + 3000, expected: GameTickDayLabel},
		{name: "day 5 noon", ticks: 24000*5 + 8000, expected: GameTickNoonLabel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ticks.DayPhaseString()
			if got != tt.expected {
				t.Errorf("Gameticks(%d).DayPhaseString() = %q, want %q", tt.ticks, got, tt.expected)
			}
		})
	}
}

func TestGameticks_Day(t *testing.T) {
	tests := []struct {
		name     string
		ticks    Gameticks
		expected int
	}{
		{name: "day 0 start", ticks: 0, expected: 0},
		{name: "day 0 end", ticks: 23999, expected: 0},
		{name: "day 1 start", ticks: 24000, expected: 1},
		{name: "day 1 middle", ticks: 30000, expected: 1},
		{name: "day 5", ticks: 24000 * 5, expected: 5},
		{name: "day 100", ticks: 24000*100 + 12345, expected: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ticks.Day()
			if got != tt.expected {
				t.Errorf("Gameticks(%d).Day() = %d, want %d", tt.ticks, got, tt.expected)
			}
		})
	}
}

func TestGameticks_DayPhaseTicks(t *testing.T) {
	tests := []struct {
		name     string
		ticks    Gameticks
		expected Gameticks
	}{
		{name: "zero", ticks: 0, expected: 0},
		{name: "mid day", ticks: 12000, expected: 12000},
		{name: "end of day", ticks: 23999, expected: 23999},
		{name: "start of day 2", ticks: 24000, expected: 0},
		{name: "day 2 noon", ticks: 24000 + 6000, expected: 6000},
		{name: "day 10 night", ticks: 24000*10 + 15000, expected: 15000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ticks.DayPhaseTicks()
			if got != tt.expected {
				t.Errorf("Gameticks(%d).DayPhaseTicks() = %d, want %d", tt.ticks, got, tt.expected)
			}
		})
	}
}

func TestGametimeService_SetTime(t *testing.T) {
	tests := []struct {
		name        string
		timeInput   string
		expectedCmd string
		rconOutput  string
		rconErr     error
		wantErr     bool
	}{
		{
			name:        "set day",
			timeInput:   "day",
			expectedCmd: "time set day",
			rconOutput:  "Set the time to 1000",
		},
		{
			name:        "set noon",
			timeInput:   "noon",
			expectedCmd: "time set noon",
			rconOutput:  "Set the time to 6000",
		},
		{
			name:        "set night",
			timeInput:   "night",
			expectedCmd: "time set night",
			rconOutput:  "Set the time to 13000",
		},
		{
			name:        "set midnight",
			timeInput:   "midnight",
			expectedCmd: "time set midnight",
			rconOutput:  "Set the time to 18000",
		},
		{
			name:        "set specific ticks",
			timeInput:   "12345",
			expectedCmd: "time set 12345",
			rconOutput:  "Set the time to 12345",
		},
		{
			name:        "rcon error",
			timeInput:   "day",
			expectedCmd: "time set day",
			rconErr:     errors.New("connection failed"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeRconClient{
				responses: map[string]struct {
					out string
					err error
				}{
					tt.expectedCmd: {out: tt.rconOutput, err: tt.rconErr},
				},
			}

			svc := NewGametimeService(fake)
			got, err := svc.SetTime(tt.timeInput)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.rconOutput {
				t.Errorf("SetTime() = %q, want %q", got, tt.rconOutput)
			}
		})
	}
}

func TestGametimeService_AddTime(t *testing.T) {
	tests := []struct {
		name        string
		ticks       Gameticks
		expectedCmd string
		rconOutput  string
		rconErr     error
		wantErr     bool
	}{
		{
			name:        "add 1000 ticks",
			ticks:       1000,
			expectedCmd: "time add 1000",
			rconOutput:  "Added 1000 to the time",
		},
		{
			name:        "add 0 ticks",
			ticks:       0,
			expectedCmd: "time add 0",
			rconOutput:  "Added 0 to the time",
		},
		{
			name:        "add large amount",
			ticks:       100000,
			expectedCmd: "time add 100000",
			rconOutput:  "Added 100000 to the time",
		},
		{
			name:        "rcon error",
			ticks:       1000,
			expectedCmd: "time add 1000",
			rconErr:     errors.New("connection failed"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeRconClient{
				responses: map[string]struct {
					out string
					err error
				}{
					tt.expectedCmd: {out: tt.rconOutput, err: tt.rconErr},
				},
			}

			svc := NewGametimeService(fake)
			got, err := svc.AddTime(tt.ticks)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.rconOutput {
				t.Errorf("AddTime() = %q, want %q", got, tt.rconOutput)
			}
		})
	}
}

func TestGametimeService_GetDayTime(t *testing.T) {
	tests := []struct {
		name       string
		rconOutput string
		rconErr    error
		expected   Gameticks
		wantErr    bool
	}{
		{
			name:       "morning",
			rconOutput: "The time is 3000",
			expected:   3000,
		},
		{
			name:       "noon",
			rconOutput: "The time is 6000",
			expected:   6000,
		},
		{
			name:       "midnight",
			rconOutput: "The time is 0",
			expected:   0,
		},
		{
			name:       "end of day",
			rconOutput: "The time is 23999",
			expected:   23999,
		},
		{
			name:    "rcon error",
			rconErr: errors.New("connection failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeRconClient{
				responses: map[string]struct {
					out string
					err error
				}{
					"time query daytime": {out: tt.rconOutput, err: tt.rconErr},
				},
			}

			svc := NewGametimeService(fake)
			got, err := svc.GetDayTime()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("GetDayTime() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestGametimeService_GetGameTime(t *testing.T) {
	tests := []struct {
		name       string
		rconOutput string
		rconErr    error
		expected   Gameticks
		wantErr    bool
	}{
		{
			name:       "new world",
			rconOutput: "The time is 1000",
			expected:   1000,
		},
		{
			name:       "old world",
			rconOutput: "The time is 1234567",
			expected:   1234567,
		},
		{
			name:    "rcon error",
			rconErr: errors.New("connection failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeRconClient{
				responses: map[string]struct {
					out string
					err error
				}{
					"time query gametime": {out: tt.rconOutput, err: tt.rconErr},
				},
			}

			svc := NewGametimeService(fake)
			got, err := svc.GetGameTime()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("GetGameTime() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestGametimeService_GetGameDay(t *testing.T) {
	tests := []struct {
		name       string
		rconOutput string
		rconErr    error
		expected   int
		wantErr    bool
	}{
		{
			name:       "day 0",
			rconOutput: "The time is 0",
			expected:   0,
		},
		{
			name:       "day 1",
			rconOutput: "The time is 1",
			expected:   1,
		},
		{
			name:       "day 100",
			rconOutput: "The time is 100",
			expected:   100,
		},
		{
			name:    "rcon error",
			rconErr: errors.New("connection failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeRconClient{
				responses: map[string]struct {
					out string
					err error
				}{
					"time query day": {out: tt.rconOutput, err: tt.rconErr},
				},
			}

			svc := NewGametimeService(fake)
			got, err := svc.GetGameDay()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("GetGameDay() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestGameTickConstants(t *testing.T) {
	// Verify constants match expected Minecraft tick values
	if GameTickDay != 1000 {
		t.Errorf("GameTickDay = %d, want 1000", GameTickDay)
	}
	if GameTickNoon != 6000 {
		t.Errorf("GameTickNoon = %d, want 6000", GameTickNoon)
	}
	if GameTickNight != 13000 {
		t.Errorf("GameTickNight = %d, want 13000", GameTickNight)
	}
	if GameTickMidnight != 18000 {
		t.Errorf("GameTickMidnight = %d, want 18000", GameTickMidnight)
	}
}

func TestGameTickLabels(t *testing.T) {
	if GameTickDayLabel != "day" {
		t.Errorf("GameTickDayLabel = %q, want %q", GameTickDayLabel, "day")
	}
	if GameTickNoonLabel != "noon" {
		t.Errorf("GameTickNoonLabel = %q, want %q", GameTickNoonLabel, "noon")
	}
	if GameTickNightLabel != "night" {
		t.Errorf("GameTickNightLabel = %q, want %q", GameTickNightLabel, "night")
	}
	if GameTickMidnightLabel != "midnight" {
		t.Errorf("GameTickMidnightLabel = %q, want %q", GameTickMidnightLabel, "midnight")
	}
}
