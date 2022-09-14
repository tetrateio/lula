package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/defenseunicorns/compliance-auditor/src/cmd/execute"
)

var rootCmd = &cobra.Command{
	Use:   "compliance-auditor",
	Short: "compliance-auditor",
	Long:  `compliance-auditor`,
}

func Execute() {

	commands := []*cobra.Command{
		execute.Command(),
	}

	rootCmd.AddCommand(commands...)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
