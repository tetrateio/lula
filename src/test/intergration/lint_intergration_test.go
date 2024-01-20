package toolstest

import (
	"strings"
	"os/exec"
	"testing"

)

// Tests the lint command.
func TestLintCmd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		wantOutput string
	}{
		{

			name: "Test the Lint command with no file flag.",
			args: []string{"tools", "lint", ""},
			wantErr: true,
			wantOutput: "",
		},
		{
			name: "Test the Lint command with file flag no file path.",
			args: []string{"tools", "lint", "-f"},
			wantErr: true,
			wantOutput: "",
		},

		{
			name: "Test the Lint command with file flag and one valid OSCAL file path.",
			args: []string{"tools", "lint", "-f", "../../../test/valid-component-definition.yaml"},
			wantErr: false,
			wantOutput: "",
		},

		{
			name: "Test the Lint command with file flag and one invalid version OSCAL file path.",
			args: []string{"tools", "lint", "-f", "../../../test/invalid-version-component-definition.yaml"},
			wantErr: true,
			wantOutput: "",
		},

		{
			name: "Test the Lint command with file flag and one unsupported OSCAL version file path.",
			args: []string{"tools", "lint", "-f", "../../../test/unsupported-version-component-definition.yaml"},
			wantErr: true,
			wantOutput: "",
		},

		{
			name: "Test the Lint command with file flag and two OSCAL file paths.",
			args: []string{"tools", "lint", "-f", "../../../valid-component-definition.yaml", "../../../test/oscal-component.yaml"},
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
