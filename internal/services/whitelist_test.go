package services

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestWhitelistService_GetServerPlayerInfo(t *testing.T) {
	tests := []struct {
		name             string
		listOutput       string
		listErr          error
		whitelistedNames []string
		wantErr          bool
	}{
		{
			name:             "2 whitelisted players",
			listOutput:       "There are 2 whitelisted player(s): Steve, Alex",
			whitelistedNames: []string{"Steve", "Alex"},
		},
		{
			name:             "no whitelisted players",
			listOutput:       "There are 0 whitelisted player(s):",
			whitelistedNames: []string{},
		},
		{
			name:    "rcon error",
			listErr: errors.New("boom"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRconClient := &fakeRconClient{
				responses: map[string]struct {
					out string
					err error
				}{
					"whitelist list": {out: tt.listOutput, err: tt.listErr},
				},
			}
			fakeMojangChecker := &fakeMojangChecker{
				existsMap: map[string]bool{},
				errMap:    map[string]error{},
			}
			fakeFileClient := &fakeFileClient{
				files: map[string]string{
					"server.properties": "white-list=true",
				},
			}
			svc := NewWhitelistService(fakeRconClient, fakeMojangChecker, fakeFileClient)
			got, err := svc.GetWhitelistInfo()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got.PlayerNames, tt.whitelistedNames) {
				t.Fatalf("whitelisted names = %#v, want %#v", got.PlayerNames, tt.whitelistedNames)
			}
		})
	}
}

func TestWhitelistService_RemoveNameFromWhitelist(t *testing.T) {
	tests := []struct {
		name          string
		inputName     string
		expectedCmd   string
		executeOutput string
		executeErr    error
		wantErr       bool
	}{
		{
			name:        "valid name",
			inputName:   "Steve",
			expectedCmd: "whitelist remove Steve",
		},
		{
			name:      "empty name",
			inputName: "   ",
			wantErr:   true,
		},
		{
			name:        "rcon error",
			inputName:   "Alex",
			expectedCmd: "whitelist remove Alex",
			executeErr:  errors.New("boom"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRconClient := &fakeRconClient{
				responses: map[string]struct {
					out string
					err error
				}{
					tt.expectedCmd: {out: tt.executeOutput, err: tt.executeErr},
				},
			}
			fakeMojangChecker := &fakeMojangChecker{
				existsMap: map[string]bool{},
				errMap:    map[string]error{},
			}
			fakeFileClient := &fakeFileClient{
				files: map[string]string{
					"server.properties": "white-list=true",
				},
			}
			svc := NewWhitelistService(fakeRconClient, fakeMojangChecker, fakeFileClient)
			err := svc.RemoveNameFromWhitelist(tt.inputName)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestWhitelistService_AddNameToWhitelist(t *testing.T) {
	tests := []struct {
		name              string
		inputName         string
		existingWhitelist []string
		mojangExists      bool
		mojangErr         error
		expectedCmd       string
		executeOutput     string
		executeErr        error
		wantErr           bool
	}{
		{
			name:              "valid name",
			inputName:         "Steve",
			existingWhitelist: []string{},
			mojangExists:      true,
			expectedCmd:       "whitelist add Steve",
		},
		{
			name:      "empty name",
			inputName: "   ",
			wantErr:   true,
		},
		{
			name:              "name already whitelisted",
			inputName:         "Alex",
			existingWhitelist: []string{"Alex"},
			wantErr:           true,
		},
		{
			name:      "mojang check error",
			inputName: "Herobrine",
			mojangErr: errors.New("boom"),
			wantErr:   true,
		},
		{
			name:         "mojang name does not exist",
			inputName:    "NotARealPlayer",
			mojangExists: false,
			wantErr:      true,
		},
		{
			name:         "rcon error",
			inputName:    "Zombie",
			mojangExists: true,
			expectedCmd:  "whitelist add Zombie",
			executeErr:   errors.New("boom"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRconClient := &fakeRconClient{
				responses: map[string]struct {
					out string
					err error
				}{
					"whitelist list": {out: formatWhitelistListOutput(tt.existingWhitelist), err: nil},
					tt.expectedCmd:   {out: tt.executeOutput, err: tt.executeErr},
				},
			}
			fakeMojangChecker := &fakeMojangChecker{
				existsMap: map[string]bool{},
				errMap:    map[string]error{},
			}
			fakeMojangChecker.existsMap[tt.inputName] = tt.mojangExists
			if tt.mojangErr != nil {
				fakeMojangChecker.errMap[tt.inputName] = tt.mojangErr
			}
			fakeFileClient := &fakeFileClient{
				files: map[string]string{
					"server.properties": "white-list=true",
				},
			}

			svc := NewWhitelistService(fakeRconClient, fakeMojangChecker, fakeFileClient)
			err := svc.AddNameToWhitelist(tt.inputName)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func formatWhitelistListOutput(names []string) string {
	if len(names) == 0 {
		return "There are 0 whitelisted player(s):"
	}
	return "There are " + fmt.Sprintf("%d", len(names)) + " whitelisted player(s): " + joinNames(names)
}

func joinNames(names []string) string {
	result := ""
	for i, name := range names {
		if i > 0 {
			result += ", "
		}
		result += name
	}
	return result
}
