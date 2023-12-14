package tools

import (
	"testing"
	"bytes"

	"github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
)

// Helper function to execute the Cobra command.
func executeLintCommand(root *cobra.Command, args ...string) (string, error) {
    buf := new(bytes.Buffer)
    root.SetOut(buf)
    root.SetErr(buf)
    root.SetArgs(args)

    _, err := root.ExecuteC()
    return buf.String(), err
}

func TestLintCommand(t *testing.T ) {
	const oscal = "../../../test/oscal-component.yaml"
	const schema = ""
	_, err := executeLintCommand(toolsCmd, "lint", "-f" + oscal)
	assert.NoError(t, err)
}