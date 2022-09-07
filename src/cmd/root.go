package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "compliance-auditor",
	Short: "compliance-auditor",
	Long:  `compliance-auditor`,
}

/*
The init function is responsible to run things
which we will require before anything else
say
  - Fetch API Keys
  - Set Logging level
  - Setup any environment variable required for the app
*/
func init() {
	rootCmd.AddCommand(executeCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
