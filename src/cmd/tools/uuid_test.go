package tools

import (
    "testing"

    "github.com/defenseunicorns/lula/src/test/lulatest"

)
// Tests the uuidgen command.
func TestUUIDGENCmd(t *testing.T) {
t.Parallel()

    tests := []struct {
        name     string
        args     []string
        wantErr  bool
        logCheck func(t *testing.T, logOutput string)
    }{
        {
            name: "Test the uuidCmd with no arguments.",
            args: []string{"uuidgen"},
            wantErr: false,
        },
        {
            name: "Test the uuidCmd with one argument.",
            args: []string{"uuidgen", "https://lula.dev"},
            wantErr: false,
        },
        {
            name: "Test the uuidCmd with too many arguments.",
            args: []string{"uuidgen", "https://lula.dev", "https://lula.dev"},
            wantErr: true,
        },

    }
// loops through the tests and tests checks if the wantErr is true/false
for _, tt := range tests {
    tt := tt
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        logOutput, err := lulatest.ExecuteTestCommand(t, toolsCmd, tt.args... )

        if (err != nil) != tt.wantErr {
            t.Errorf("ToolsCmd() error = %v, wantErr %v", err, tt.wantErr)
        }

        if tt.logCheck != nil {
            tt.logCheck(t, logOutput)
        }
    })
}
}