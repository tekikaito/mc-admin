package services

import (
	"errors"
	"strings"
	"testing"
)

func TestCommandService_ExecuteRawCommand(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		fakeOut      string
		fakeErr      error
		wantOut      string
		wantErr      bool
		wantErrSub   string
		wantCmdSent  string
		wantCalls    int
		checkWrapErr bool
	}{
		{
			name:       "empty command",
			input:      "",
			wantErr:    true,
			wantErrSub: "command cannot be empty",
			wantCalls:  0,
		},
		{
			name:       "whitespace only",
			input:      "   \n\t  ",
			wantErr:    true,
			wantErrSub: "command cannot be empty",
			wantCalls:  0,
		},
		{
			name:        "trims whitespace before executing",
			input:       "  list  ",
			fakeOut:     "There are 0 of a max of 20 players online",
			wantOut:     "There are 0 of a max of 20 players online",
			wantCmdSent: "list",
			wantCalls:   1,
		},
		{
			name:         "wraps rcon errors",
			input:        "say hello",
			fakeErr:      errors.New("boom"),
			wantErr:      true,
			wantErrSub:   "failed to execute command",
			wantCmdSent:  "say hello",
			wantCalls:    1,
			checkWrapErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeRconClient{responses: map[string]struct {
				out string
				err error
			}{}}

			trimmed := strings.TrimSpace(tt.input)
			if trimmed != "" {
				fake.responses[trimmed] = struct {
					out string
					err error
				}{out: tt.fakeOut, err: tt.fakeErr}
			}

			svc := NewCommandServiceFromRconClient(fake)
			got, err := svc.ExecuteRawCommand(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrSub != "" && !strings.Contains(err.Error(), tt.wantErrSub) {
					t.Fatalf("error = %q, expected to contain %q", err.Error(), tt.wantErrSub)
				}
				if tt.checkWrapErr && tt.fakeErr != nil && !errors.Is(err, tt.fakeErr) {
					t.Fatalf("expected error to wrap %v", tt.fakeErr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tt.wantOut {
					t.Fatalf("output = %q, want %q", got, tt.wantOut)
				}
			}

			if len(fake.received) != tt.wantCalls {
				t.Fatalf("ExecuteCommand call count = %d, want %d", len(fake.received), tt.wantCalls)
			}
			if tt.wantCalls > 0 {
				wantCmd := tt.wantCmdSent
				if wantCmd == "" {
					wantCmd = strings.TrimSpace(tt.input)
				}
				if fake.received[0] != wantCmd {
					t.Fatalf("ExecuteCommand called with %q, want %q", fake.received[0], wantCmd)
				}
			}
		})
	}
}
