package tools

import (
	"testing"

	"strings"

	"sync"

	"fmt"

	"github.com/defenseunicorns/lula/src/test/lulatest"

	"github.com/spf13/cobra"
)

var mutex sync.Mutex

// Tests the uuidgen command.

func TestUUIDGENCmd(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name string

		args []string

		wantErr bool

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

	for _, tt := range tests {

		tt := tt

		t.Run(tt.name, func(t *testing.T) {

			t.Parallel()

			mutex.Lock() // Lock before executing command

			// defer mutex.Unlock() // Unlock after executing the command

			logOutput, err := lulatest.ExecuteTestCommand(t, func() *cobra.Command {

				fmt.Println("Debug: Executing uuidCmd")

				return uuidCmd

			}, tt.args...)

			mutex.Unlock()

			if (err != nil) != tt.wantErr {

				t.Errorf("ToolsCmd() error = %v, wantErr %v", err, tt.wantErr)

			}

			if tt.logCheck != nil {

				tt.logCheck(t, logOutput)

			}

		})

	}

}
