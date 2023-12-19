package tools

import (
    "testing"
    "strings"
    "bytes"

    "github.com/defenseunicorns/lula/src/test/lulatest"
    "github.com/spf13/cobra"

)

// func toolsCmdFactory() *cobra.Command {
//     // Create a new instance of cobra.Command
//     var newCmd = &cobra.Command{
// 		Use:     "uuidgen",
// 		Short:   "Generate a UUID",
// 		Long:    "Generate a UUID at random or deterministically with a provided input",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			if len(args) == 0 {
// 				fmt.Println(uuid.NewUUID())
// 				return nil
// 			} else if len(args) == 1 {
// 				fmt.Println(uuid.NewUUIDWithSource(args[0]))
// 				return nil
// 			} else {
// 				return fmt.Errorf("too many arguments")
// 			}
// 		},
//     }

//     return newCmd
// }

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
            args: []string{""},
            wantErr: false,
        },
        {
            name: "Test the uuidCmd with one argument.",
            args: []string{"https://lula.dev"},
            wantErr: false,
        },
        {
            name: "Test the uuidCmd with too many arguments.",
            args: []string{"https://lula.dev", "https://lula2.dev"},
            wantErr: true,
            logCheck: func(t *testing.T, logOutput string) {
                if !strings.Contains(logOutput, "too many arguments") {
                    t.Errorf("Expected log output to contain 'too many arguments', got: %s", logOutput)
                }
            },
        },

    }

// loops through the tests and tests checks if the wantErr is true/false

// Swapped cmdFactory CommandFactory out for func() *cobra.Command { return uuniCmd} to pass in the uuidCmd into the ExecuteTestCommand(shared function) to trigger and handle errors etc.
for _, tt := range tests {
    tt := tt
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
     logOutput, err := lulatest.ExecuteTestCommand(t, func() *cobra.Command {
            return uuidCmd
        }, tt.args... )

        if (err != nil) != tt.wantErr {
            t.Errorf("ToolsCmd() error = %v, wantErr %v", err, tt.wantErr)
        }

        if tt.logCheck != nil {
            tt.logCheck(t, logOutput)
        }
    })
}
}

// Testing uuidCmd individually(Excluding the shared ExecuteTestCommand) trying to narrow down if issue is the shared function or in uuidCmd errors.
// Current thoughts is the way errors are handled in the uuidCmd or how I am grabbing that error in the test.
func TestUUIDCmdTooManyArguments(t *testing.T) {
    cmd := uuidCmd
    cmd.SetArgs([]string{"https://lula.dev", "https://lula2.dev"})

    stdoutBuf := new(bytes.Buffer)
    stderrBuf := new(bytes.Buffer)
    cmd.SetOut(stdoutBuf)
    cmd.SetErr(stderrBuf)

    err := cmd.Execute()

    if err == nil {
        t.Fatalf("Expected error, got nil")
    } else if err.Error() != "too many arguments" {
        t.Fatalf("Expected 'too many arguments' error, got: %v", err)
    }

    t.Logf("Stdout: %s", stdoutBuf.String())
    t.Logf("Stderr: %s", stderrBuf.String())
}

// Testing the RunE in uuidCmd (This actually works but it dosen't test the cmd as a whole)
func TestUUIDCmdDirectRunE(t *testing.T) {
    err := uuidCmd.RunE(nil, []string{"arg1", "arg2"})
    
    if err == nil {
        t.Fatal("Expected error, got nil")
    } else if err.Error() != "too many arguments" {
        t.Fatalf("Expected 'too many arguments' error, got: %v", err)
    }
}

