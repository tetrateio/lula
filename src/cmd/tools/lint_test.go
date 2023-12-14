package tools

import (
	"testing"
	"fmt"
	"strings"

	"github.com/defenseunicorns/lula/src/test/lulatest"
)

// // Helper function to execute the Cobra command.
// func executeLintCommand(root *cobra.Command, args ...string) (string, error) {
//     buf := new(bytes.Buffer)
//     root.SetOut(buf)
//     root.SetErr(buf)
//     root.SetArgs(args)

//     _, err := root.ExecuteC()
//     return buf.String(), err
// }

// func TestLintCommand(t *testing.T ) {
// 	const oscal = "../../../test/oscal-component.yaml"
// 	const schema = ""
// 	_, err := executeLintCommand(toolsCmd, "lint", "-f" + oscal)
// 	assert.NoError(t, err)
// }

func checkLogForError(t *testing.T, logOutput string) {
    expectedError := "expected error message" // Modify as needed for each test case
    if !strings.Contains(logOutput, expectedError) {
        t.Errorf("Expected log to contain '%s', got %s", expectedError, logOutput)
    }
}

func TestLintCmd(t *testing.T) {
t.Parallel()

    tests := []struct {
        name     string
        args     []string
        wantErr  bool
        logCheck func(t *testing.T, logOutput string)
    }{
        {
            name: "Returns an error if no input file is provided",
            args: []string{"lint", "-f", " "},
            wantErr: true,
			logCheck: checkLogForError,
        },
        {
            name: "Returns an error if the input file is not a json or yaml file.",
            args: []string{"lint", "-f", "test.txt"},
            wantErr: true,
			logCheck: checkLogForError,
        },
        {
            name: "Returns an error if the input file is not a valid oscal version.",
            args: []string{"lint", "-f", lulatest.InvalidVersionComponentPath},
            wantErr: true,
			logCheck: checkLogForError,
        },
        {
            name: "returns an error if the input file is not a supported oscal version",
            args: []string{"lint", "-f", lulatest.UnsupportedVersionComponentPath},
            wantErr: true,
			logCheck: checkLogForError,
        },
        {
            name: "Returns an error if it fails to read the input file",
            args: []string{"lint", "-f", "test.yaml"},
            wantErr: true,
			logCheck: checkLogForError,
        },
        {
            name: "logs a success message if the input file is valid",
            args: []string{"lint", "-f", lulatest.ValidComponentPath},
            wantErr: false,
			logCheck: checkLogForError,
        },

    }

for _, tt := range tests {
    tt := tt
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
		fmt.Println("Executing command with args:", tt.args)
        logOutput, err := lulatest.ExecuteTestCommand(t, toolsCmd, tt.args... )
		fmt.Println("Command executed. Error:", err)

        if (err != nil) != tt.wantErr {
            t.Errorf("ToolsCmd() error = %v, wantErr %v", err, tt.wantErr)
        }

        if tt.logCheck != nil {
            tt.logCheck(t, logOutput)
        }
    })
}
}