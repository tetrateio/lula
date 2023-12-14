package tools

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestInclude(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	Include(rootCmd)
	assert.True(t, containsCommand(rootCmd.Commands(), toolsCmd), "toolsCmd should be a subcommand of rootCmd")
}

func containsCommand(commands []*cobra.Command, cmd *cobra.Command) bool {
	for _, c := range commands {
		if c == cmd {
			return true
		}
	}

	return false
}
