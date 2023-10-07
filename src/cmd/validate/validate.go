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

		results := types.ResultObject{
			FilePaths: componentDefinitionPaths,
		}

		// Primary expected path for validation of OSCAL documents
		err := ValidateOnPaths(&results)
		if err != nil {
			return fmt.Errorf("Validation error: %w\n", err)
		}

		// Convert Results to expected report format
		report, err := GenerateReportFromResults(result)
		if err != nil {
			return fmt.Errorf("Generate error: %w\n", err)
		}

		// Write report(s) to file
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

/*
	To tell the validation story:
		Lula is currently evaluating controls identified in the Implemented-Requirements of a component-definition.
		We would then be looking to retain information that may be required for relation of component-definition (input) to an assessment-results (output).
		In order to get there - we have to traverse and possibly track UUIDs at a minimum:

		Lula accepts 1 -> N paths to OSCAL component-definition files
		For each component definition:
			There are 1 -> N Components
			For each component:
				There are 1 -> N control-Implementations
				For each control-implementation:
					There are 1-> N implemented-requirements
					For each implemented-requirement:
						There are 1 -> N validations
							This allows for breaking complex query and policy into smaller  chunks
						Validations are evaluated individually with passing/failing resources
					Pass/Fail results from all validations is evaluated for a pass/fail status in the report

	As such, building a ReportObject to collect and retain the relational information could be preferred

*/

// ValidateOnPath takes 1 -> N paths to OSCAL component-definition files
// It will then read those files to perform validation and return an ResultObject
func ValidateOnPaths(obj *types.ReportObject) error {
	resultMap := make(map[string]map[string][]types.Result, 0)
	// for each path
	for _, path := range obj.FilePaths {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			fmt.Printf("Path: %v does not exist - unable to digest document\n", path)
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		compDef, err := oscal.NewOscalComponentDefinition(data)
		err = ValidateOnCompDef(obj, compDef)

		obj.UUID = compDef.UUID
	}

	return nil

}

// ValidateOnCompDef takes a single ComponentDefinition object
// It will perform a validation and add data to a referenced report object
func ValidateOnCompDef(obj *types.ReportObject, compDef oscalTypes.ComponentDefinition) error {
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

func ValidateOnCompDef(obj *types.ReportObject, compDef oscalTypes.ComponentDefinition) error {

	// Loops all the way down
	for _, component := range compDef.Components {
		for _, controlImplementation := range component.ControlImplementations {
			for _, implementedRequirement := range controlImplementation.ImplementedRequirements {
				for _, target := range implementedRequirement.Rules {

				}
			}
		}
	}

	return nil
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
