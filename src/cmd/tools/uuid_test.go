package tools

import (
    "testing"
    "fmt"

    "github.com/defenseunicorns/lula/src/test/lulatest"
    "github.com/defenseunicorns/go-oscal/src/pkg/uuid"
    "github.com/spf13/cobra"

)

func toolsCmdFactory() *cobra.Command {
    // Create a new instance of cobra.Command
    var newCmd = &cobra.Command{
		Use:     "uuidgen",
		Short:   "Generate a UUID",
		Long:    "Generate a UUID at random or deterministically with a provided input",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Println(uuid.NewUUID())
				return nil
			} else if len(args) == 1 {
				fmt.Println(uuid.NewUUIDWithSource(args[0]))
				return nil
			} else {
				return fmt.Errorf("too many arguments")
			}
		},
    }

    return newCmd
}

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
        },

    }
// loops through the tests and tests checks if the wantErr is true/false
for _, tt := range tests {
    tt := tt
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        logOutput, err := lulatest.ExecuteTestCommand(t, toolsCmdFactory, tt.args... )

        if (err != nil) != tt.wantErr {
            t.Errorf("ToolsCmd() error = %v, wantErr %v", err, tt.wantErr)
        }

        if tt.logCheck != nil {
            tt.logCheck(t, logOutput)
        }
    })
}
}