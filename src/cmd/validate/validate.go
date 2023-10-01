package validate

import (
	"fmt"
	"os"

	"github.com/defenseunicorns/lula/src/pkg/oscal"
	"github.com/defenseunicorns/lula/src/pkg/providers/cel"
	"github.com/spf13/cobra"
)

var validateHelp = `
To validate on a cluster:
	lula validate ./oscal-component.yaml

To validate on a resource:
	lula validate ./oscal-component.yaml -r resource.yaml

To validate without creation of any report files
	lula validate ./oscal-component.yaml -d
`

var resourcePaths []string
var cluster, dryRun bool

// this is a temporary struct until a reporting model is selected
type ComplianceReport struct {
	UUID        string `json:"uuid" yaml:"uuid"`
	ControlId   string `json:"control-id" yaml:"control-id"`
	Description string `json:"description" yaml:"description"`
	Result      string `json:"result" yaml:"result"`
}

var ValidateCmd = &cobra.Command{
	Use:     "validate",
	Short:   "validate",
	Example: validateHelp,
	Run: func(cmd *cobra.Command, componentDefinitionPaths []string) {
		// Conduct further error checking here (IE flags/arguments)
		if len(componentDefinitionPaths) == 0 {
			fmt.Println("Path to the local OSCAL file must be present")
			fmt.Print(validateHelp)
			os.Exit(1)
		}

		err := conductValidate(componentDefinitionPaths, resourcePaths, dryRun)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func ValidateCommand() *cobra.Command {
	ValidateCmd.Flags().StringArrayVarP(&resourcePaths, "resource", "r", []string{}, "Path to resource files")
	ValidateCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Specifies whether to write reports to filesystem")

	return ValidateCmd
}

func conductValidate(componentDefinitionPaths []string, resourcePaths []string, dryRun bool) error {

	// process the static vs live query switch

	// can we create a map of map[path][]byte and pass that to an oscal function that returns a type?\
	fileData := make(map[string][]byte, 0)

	for _, path := range componentDefinitionPaths {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			fmt.Printf("Path: %v does not exist - unable to digest document\n", path)
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fileData[path] = data

	}

	// unmarshall all documents to types.OscalComponentDocument into a slice of component documents
	// Declare empty slice of oscalComponentDocuments
	oscalComponentDefinitions, err := oscal.ManyNewOscalComponentDefinitions(fileData)
	if err != nil {
		return err
	}

	var complianceReports []ComplianceReport

	// with an array/slice of oscalComponentDocuments, search each implemented requirement for a props.name/value (hardcoded and specific value for now)
	// copy struct to slice of implementedRequirements
	// foreach oscalComponentDocument -- foreach implemented requirement -- if props.name == compliance validator
	implementedReqs, err := oscal.GetImplementedRequirements(oscalComponentDefinitions)

	for _, implementedReq := range implementedReqs {

		// for each validation
		for _, target := range implementedReq.Rules {
			// for each rule
			fmt.Println(target["provider"])
			// identify the provider
			// simple conditional until more providers are introduced
			if provider, ok := target["provider"].(string); ok && provider == "cel" {
				err := cel.Validate(target)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Println("Provider not found")
				continue
			}
			// perform the validation and get the results

			// process the result

			var currentReport ComplianceReport

			// var resultString string
			// if results.Match > 0 && results.NonMatch <= 0 {
			// 	resultString = "Pass"
			// } else {
			// 	resultString = "Fail"
			// }

			currentReport.UUID = implementedReq.UUID
			currentReport.ControlId = implementedReq.ControlId
			// currentReport.Result = resultString
			currentReport.Result = "Pass"

			complianceReports = append(complianceReports, currentReport)

			// fmt.Printf("UUID: %v\n\tResources Matching: %v\n\tResources non-matching: %v\n\tStatus: %v\n", implementedReq.UUID, results.Match, results.NonMatch, resultString)
		}

	}
	if err != nil {
		return err
	}

	return nil
}
