package services

import (
	"errors"
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
			svc := NewWhitelistService(fakeRconClient, fakeMojangChecker)
			got, err := svc.GetWhitelist()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.whitelistedNames) {
				t.Fatalf("whitelisted names = %#v, want %#v", got, tt.whitelistedNames)
			}
		})
	}
}
