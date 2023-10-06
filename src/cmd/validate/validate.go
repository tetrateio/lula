package validate

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/defenseunicorns/lula/src/pkg/oscal"
	"github.com/defenseunicorns/lula/src/types"
	"github.com/defenseunicorns/lula/src/types/oscal"
	"github.com/spf13/cobra"
	yaml1 "sigs.k8s.io/yaml"
)

var validateHelp = `
To validate on a cluster:
	lula validate ./oscal-component.yaml
`

var cluster bool

var ValidateCmd = &cobra.Command{
	Use:     "validate",
	Short:   "validate",
	Example: validateHelp,
	RunE: func(cmd *cobra.Command, componentDefinitionPaths []string) error {
		// Conduct further error checking here (IE flags/arguments)
		if len(componentDefinitionPaths) == 0 {
			fmt.Println(cmd.Long)
			return errors.New("Path to OSCAL component definition(s) required")
		}

		result, err := ValidateOnPaths(componentDefinitionPaths)
		if err != nil {
			return fmt.Errorf("Validation error: %w\n", err)
		}

		report, err := GenerateReportFromResults(result)
		if err != nil {
			return fmt.Errorf("Generate error: %w\n", err)
		}

		err = WriteReport(report)
		if err != nil {
			return fmt.Errorf("Write error: %w\n", err)
		}
		return nil
	},
}

func ValidateCommand() *cobra.Command {

	// insert flag options here
	return ValidateCmd
}

// ValidateOnPath takes 1 -> N paths to OSCAL component-definition files
// It will then read those files to perform validation and return an ResultObject
func ValidateOnPaths(componentDefinitionPaths []string) (map[string]map[string][]types.Result, error) {
	resultMap := make(map[string]map[string][]types.Result, 0)
	// for each path
	for _, path := range componentDefinitionPaths {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			fmt.Printf("Path: %v does not exist - unable to digest document\n", path)
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return resultMap, err
		}

		compDef, err := oscal.NewOscalComponentDefinition(data)
		results, err := ValidateOnCompDef(compDef.ComponentDefinition)

		resultMap[compDef.ComponentDefinition.UUID] = results
	}

	return resultMap, nil

}

// ValidateOnCompDef takes a single ComponentDefinition object
// It will perform a validation and return a map[string][]Result
// This keep track of the control-implementation UUID for which each component may have many
func ValidateOnCompDef(compDef oscalTypes.ComponentDefinition) (map[string][]types.Result, error) {
	results := make(map[string][]types.Result, 0)

	controlImplementations, err := oscal.GetImplementedRequirements(compDef)
	if err != nil {
		return results, nil
	}

	for _, implementedReqs := range controlImplementations {
		for _, implementedReq := range implementedReqs {
			for _, target := range implementedReq.Rules {
				result, err := ValidateOnTarget(target)

				if err != nil {
					return results, nil
				}

				results[implementedReq.UUID] = append(results[implementedReq.UUID], result)
			}
		}

		// for each validation - this is what will change when we become OSCAL schema compliant

	}
	return results, nil
}

// ValidateOnTarget takes a map[string]interface{}
// It will return a single Result
func ValidateOnTarget(target map[string]interface{}) (types.Result, error) {
	var result types.Result
	// for each rule
	// identify the provider
	// simple conditional until more providers are introduced
	// if provider, ok := target["provider"].(string); ok && provider == "opa" {
	// 	results, err = opa.Validate(ctx, target)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// } else {
	// 	fmt.Println("Provider not found")
	// 	continue
	// }

	// mock result until until initial provider is created
	result.UUID = "12345"
	result.ControlId = "cm4.1"
	result.Failing = 0
	result.Passing = 1
	result.Description = "This control ensures results can be processed"

	return result, nil

}

// TODO: this needs to evolve quite a bit
func GenerateReportFromResults(results map[string]map[string][]types.Result) ([]types.ComplianceReport, error) {
	var complianceReports []types.ComplianceReport
	// TODO: need to grab identifying information about the component-definition and component
	// component-definition -> component -> control-implementation -> implemented-requirements -> targets
	for _, controlImplementation := range results {

		for id, results := range controlImplementation {
			currentReport := types.ComplianceReport{
				UUID: id,
			}
			var pass, fail int
			for _, result := range results {
				currentReport.ControlId = result.ControlId
				pass += result.Passing
				fail += result.Failing

			}
			var resultString string
			if pass > 0 && fail <= 0 {
				resultString = "Pass"
			} else {
				resultString = "Fail"
			}
			currentReport.Result = resultString
			complianceReports = append(complianceReports, currentReport)
			fmt.Printf("UUID: %v\n\tResources Passing: %v\n\tResources Failing: %v\n\tStatus: %v\n", currentReport.UUID, pass, fail, resultString)
		}
	}

	return complianceReports, nil
}

// This is the OSCAL document generation for final output.
// This should include some ability to consolidate controls met in multiple input documents under single control entries
// This should include fields that reference the source of the control to the original document ingested
func WriteReport(compiledReport []types.ComplianceReport) error {
	reportData, err := yaml1.Marshal(&compiledReport)
	if err != nil {
		return err
	}

	currentTime := time.Now()
	fileName := "compliance_report-" + currentTime.Format("01-02-2006-15:04:05") + ".yaml"

	err = os.WriteFile(fileName, reportData, 0644)
	if err != nil {
		return err
	}
	return nil
}
