package tools

import (
	"os"

	oscalTypes_1_1_2 "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/defenseunicorns/lula/src/pkg/common/oscal"

	"github.com/defenseunicorns/lula/src/config"
	"github.com/defenseunicorns/lula/src/pkg/message"
	"github.com/spf13/cobra"
)

type combineFlags struct {
	InputFiles []string // -f --input-files
	OutputFile string   // -o --output-file
}

var combineOpts = &combineFlags{}

var combineHelp = `
To lint existing OSCAL files:
	lula tools combine -f <path1>,<path2>,<path3> [-o <output-file>]
`

func init() {
	combineCmd := &cobra.Command{
		Use:   "combine",
		Short: "Combine OSCAL models",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			config.SkipLogFile = true
		},
		Long:    "Combine multiple OSCAL documents into a single model",
		Example: combineHelp,
		Run: func(cmd *cobra.Command, args []string) {
			models := []*oscalTypes_1_1_2.OscalModels{}
			if len(combineOpts.InputFiles) == 0 {
				message.Fatalf(nil, "No input files specified")
			}

			for _, inputFile := range combineOpts.InputFiles {
				spinner := message.NewProgressSpinner("Combining OSCAL files %s", inputFile)
				defer spinner.Stop()

				data, err := os.ReadFile(inputFile)
				if err != nil {
					message.Fatalf(err, "Failed to read %s", inputFile)
				}

				model, err := oscal.NewOscalModel(data)
				if err != nil {
					message.Fatalf(err, "Failed to load oscal from file: %s, %v", inputFile, err)
				}

				models = append(models, model)
			}

			combinedModel, err := oscal.CombineOscalModels(models)
			if err != nil {
				message.Fatalf(err, "Failed to combine models: %v", err)
			}

			// If result file is specified, write the validation results to the file
			if opts.ResultFile != "" {
				oscal.WriteOscalModel(opts.ResultFile, combinedModel)
			} else {
				oscal.WriteOscalModel("combined.yaml", combinedModel)
			}
		},
	}

	toolsCmd.AddCommand(combineCmd)

	combineCmd.Flags().StringSliceVarP(&combineOpts.InputFiles, "input-files", "f", []string{}, "the paths to oscal files (comma-separated)")
	combineCmd.Flags().StringVarP(&combineOpts.OutputFile, "output-file", "r", "", "the path to write the combined model")
}
