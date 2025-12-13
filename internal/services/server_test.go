package services

import (
	"errors"
	"reflect"
	"testing"
)

func TestServerService_GetServerPlayerInfo(t *testing.T) {
	tests := []struct {
		name       string
		listOutput string
		listErr    error
		wantInfo   ServerPlayerInfo
		wantErr    bool
	}{
		{
			name:       "no players online",
			listOutput: "There are 0 of a max of 20 players online",
			wantInfo:   ServerPlayerInfo{PlayerNames: []string{}, OnlineCount: 0, MaxCount: 20},
		},
		{
			name:       "some players online",
			listOutput: "There are 2 of a max of 20 players online: Steve, Alex",
			wantInfo:   ServerPlayerInfo{PlayerNames: []string{"Steve", "Alex"}, OnlineCount: 2, MaxCount: 20},
		},
		{
			name:    "rcon error",
			listErr: errors.New("boom"),
			wantErr: true,
		},
		{
			name:       "malformed count",
			listOutput: "Players online: 1/20",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeRconClient{
				responses: map[string]struct {
					out string
					err error
				}{
					"list": {out: tt.listOutput, err: tt.listErr},
				},
			}
			svc := NewServerServiceFromRconClient(fake)

			got, err := svc.GetServerPlayerInfo()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.OnlineCount != tt.wantInfo.OnlineCount || got.MaxCount != tt.wantInfo.MaxCount {
				t.Fatalf("counts = (%d,%d), want (%d,%d)",
					got.OnlineCount, got.MaxCount, tt.wantInfo.OnlineCount, tt.wantInfo.MaxCount)
			}
			if !reflect.DeepEqual(got.PlayerNames, tt.wantInfo.PlayerNames) {
				t.Fatalf("PlayerNames = %#v, want %#v", got.PlayerNames, tt.wantInfo.PlayerNames)
			}
		})
	}
}

func TestServerService_KickPlayerByName(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		reason      string
		wantCommand string
		wantErr     bool
	}{
		{
			name:        "no reason",
			inputName:   "Steve",
			reason:      "",
			wantCommand: "kick Steve",
		},
		{
			name:        "with reason",
			inputName:   "Alex",
			reason:      "Griefing",
			wantCommand: "kick Alex Griefing",
		},
		{
			name:      "empty name",
			inputName: "   ",
			reason:    "whatever",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeRconClient{
				responses: map[string]struct {
					out string
					err error
				}{
					tt.wantCommand: {out: "", err: nil},
				},
			}
			svc := NewServerServiceFromRconClient(fake)

			err := svc.KickPlayerByName(tt.inputName, tt.reason)
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
