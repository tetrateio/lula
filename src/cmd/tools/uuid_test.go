package tools

import (
    "testing"

    // "github.com/defenseunicorns/lula/src/test/lulatest"
    "/src/test/lulatest"

)

// // Helper function to execute the Cobra command.
// func executeUUIDCommand(root *cobra.Command, args ...string) (string, error) {
//     buf := new(bytes.Buffer)
//     root.SetOut(buf)
//     root.SetArgs(args)

//     _, err := root.ExecuteC()
//     return buf.String(), err
// }

// // Test the uuidCmd with no arguments.
// func TestUUIDCmdNoArgs(t *testing.T) {
//     _, err := executeUUIDCommand(toolsCmd, "uuidgen")
//     assert.NoError(t, err)
// }


// // Test the uuidCmd with one argument.
// func TestUUIDCmdWithSource(t *testing.T) {
//     const source = "https://lula.dev"
//     _, err := executeUUIDCommand(toolsCmd, "uuidgen", source)
//     assert.NoError(t, err)
// }


// // Test the uuidCmd with too many arguments.
// func TestUUIDCmdTooManyArgs(t *testing.T) {
//     _, err := executeUUIDCommand(toolsCmd, "uuidgen", "arg1", "arg2")
//     assert.Error(t, err)
// }

// func TestUUIDGENCmd(t *testing.T) {
//     t.Parallel()

// }

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
            args: []string{},
            wantErr: false,
        },
        {
            name: "Test the uuidCmd with one argument.",
            args: []string{"https://lula.dev"},
            wantErr: false,
        },
        {
            name: "Test the uuidCmd with too many arguments.",
            args: []string{"https://lula.dev", "https://lula.dev"},
            wantErr: true,
        },

    }

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        logOutput, err := lulatest.executeTestCommand(t, toolsCmd, tt.args... )

        if (err != nil) != tt.wantErr {
            t.Errorf("ToolsCmd() error = %v, wantErr %v", err, tt.wantErr)
        }

        if tt.logCheck != nil {
            tt.logCheck(t, logOutput)
        }
    })
}
}