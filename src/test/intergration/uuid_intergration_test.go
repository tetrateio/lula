package toolstest

import (
	"strings"
	"os/exec"
	"testing"

)

// Tests the uuidgen command.
func TestUUIDGENCmd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		wantOutput string
	}{
		{

			name: "Test the uuidCmd with no arguments.",
			args: []string{"tools", "uuidgen", ""},
			wantErr: false,
			wantOutput: "",
		},
		{
			name: "Test the uuidCmd with one argument.",
			args: []string{"tools", "uuidgen", "https://lula.dev"},
			wantErr: false,
			wantOutput: "",
		},

		{
			name: "Test the uuidCmd with too many arguments.",
			args: []string{"tools", "uuidgen", "https://lula.dev", "https://lula2.dev"},
			wantErr: true,
			wantOutput: "",
		},
	}

	// loops through the tests and tests checks if the wantErr is true/false
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command("../../../bin/lula", tt.args...)

			// Execute the command
            output, err := cmd.CombinedOutput()
			
			// Check for expected error state
            if (err != nil) != tt.wantErr {
                t.Errorf("uuidgen() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

			if tt.wantOutput != "" && !strings.Contains(string(output), tt.wantOutput) {
                t.Errorf("uuidgen() got output = %v, want %v", string(output), tt.wantOutput)
            }
		})
	}
}
