package validate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	types "github.com/defenseunicorns/lula/src/internal/types"
	kube "github.com/defenseunicorns/lula/src/pkg/k8s"
	"github.com/defenseunicorns/lula/src/pkg/opa"
	yaml2 "github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	log "sigs.k8s.io/controller-runtime/pkg/log"
)

type Resource struct {
	Name   string            `json:"name"`
	Values map[string]string `json:"values"`
}

type Policy struct {
	Name      string     `json:"name"`
	Resources []Resource `json:"resources"`
}

type Values struct {
	Policies []Policy `json:"policies"`
}

type SkippedInvalidPolicies struct {
	skipped []string
	invalid []string
}

var validateHelp = `
To validate on a cluster:
	lula validate ./oscal-component.yaml

To validate on a resource:
	lula validate ./oscal-component.yaml -r resource.yaml

To validate without creation of any report files
	lula validate ./oscal-component.yaml -d
`

var generateHelp = `
To generate kyverno policies:
	lula generate ./oscal-component.yaml -o ./out
`

var resourcePaths []string
var cluster, dryRun bool

var osExit = os.Exit

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
			log.Log.Error(err, "error string")
		}
	},
}

func ValidateCommand() *cobra.Command {
	ValidateCmd.Flags().StringArrayVarP(&resourcePaths, "resource", "r", []string{}, "Path to resource files")
	ValidateCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Specifies whether to write reports to filesystem")

	return ValidateCmd
}

var outDirectory string

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func conductValidate(componentDefinitionPaths []string, resourcePaths []string, dryRun bool) error {

	// process the static vs live query switch

	// unmarshall all documents to types.OscalComponentDocument into a slice of component documents
	// Declare empty slice of oscalComponentDocuments
	oscalComponentDefinitions, err := oscalComponentDefinitionsFromPaths(componentDefinitionPaths)
	check(err)

	var complianceReports []types.ComplianceReport

	// with an array/slice of oscalComponentDocuments, search each implemented requirement for a props.name/value (hardcoded and specific value for now)
	// copy struct to slice of implementedRequirements
	// foreach oscalComponentDocument -- foreach implemented requirement -- if props.name == compliance validator
	implementedReqs, err := getImplementedReqs(oscalComponentDefinitions)
	for _, implementedReq := range implementedReqs {
		ctx := context.Background()

		// TODO: add the functionality here for doing the actual processing

		// for each target
		for _, target := range implementedReq.Rules {
			var resources []unstructured.Unstructured
			// TODO - Per target - process domain and execute query accordingly
			switch domain := target.Domain; domain {
			case "kubernetes":
				resources, err = queryKube(ctx, target)
				if err != nil {
					fmt.Println(err)
				}
			default:
				fmt.Printf("No domain connector available for %s", domain)
			}

			// Maybe silly? marshall to json and unmarshall to []map[string]interface{}
			jsonData, err := json.Marshal(resources)
			if err != nil {
				fmt.Println(err)
			}
			var data []map[string]interface{}
			err = json.Unmarshal(jsonData, &data)

			if err != nil {
				fmt.Println(err)
			}

			var includedData []map[string]interface{}
			for _, value := range data {
				resourceNamespace := value["metadata"].(map[string]interface{})["namespace"]

				exclude := false
				for _, exns := range target.Exclude {
					if exns == resourceNamespace {
						exclude = true
					}
				}
				if !exclude {
					includedData = append(includedData, value)
				}
			}

			// Call GetMatchedAssets()
			results, err := opa.GetMatchedAssets(ctx, string(target.Rego), includedData)
			if err != nil {
				fmt.Println(err)
			}

			// Now let's do something with this
			fmt.Println(results.Match)

			var currentReport types.ComplianceReport

			var resultString string
			if results.Match > 0 && results.NonMatch <= 0 {
				resultString = "Pass"
			} else {
				resultString = "Fail"
			}

			currentReport.SourceRequirements = implementedReq
			currentReport.Result = resultString

			complianceReports = append(complianceReports, currentReport)

			fmt.Printf("UUID: %v\n\tResources Matching: %v\n\tResources non-matching: %v\n\tStatus: %v\n", implementedReq.UUID, results.Match, results.NonMatch, resultString)
		}

		// END OF opa logic

	}
	if err != nil {
		log.Log.Error(err, "error string")
	}

	return nil
}

// Open files and attempt to unmarshall to oscal component definition structs
func oscalComponentDefinitionsFromPaths(filepaths []string) (oscalComponentDefinitions []types.OscalComponentDefinitionModel, err error) {
	for _, path := range filepaths {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			fmt.Printf("Path: %v does not exist - unable to digest document\n", path)
			continue
		}

		rawDoc, err := os.ReadFile(path)
		check(err)

		var oscalComponentDefinition types.OscalComponentDefinitionModel

		jsonDoc, err := yaml2.YAMLToJSON(rawDoc)
		if err != nil {
			fmt.Printf("Error converting YAML to JSON: %s\n", err.Error())
		}

		err = json.Unmarshal(jsonDoc, &oscalComponentDefinition)
		check(err)

		oscalComponentDefinitions = append(oscalComponentDefinitions, oscalComponentDefinition)
	}

	return
}

// Parse the ingested documents (POC = 1) for applicable information
// Knowns = this will be a yaml file
// return a slice of Control objects
func getImplementedReqs(componentDefinitions []types.OscalComponentDefinitionModel) (implementedReqs []types.ImplementedRequirement, err error) {
	for _, componentDefinition := range componentDefinitions {
		for _, component := range componentDefinition.ComponentDefinition.Components {
			for _, controlImplementation := range component.ControlImplementations {
				implementedReqs = append(implementedReqs, controlImplementation.ImplementedRequirements...)
			}
		}
	}
	return
}

func queryKube(ctx context.Context, target types.RegoTargets) (resources []unstructured.Unstructured, err error) {

	config := ctrl.GetConfigOrDie()
	dynamic := dynamic.NewForConfigOrDie(config)

	for _, kind := range target.Kinds {
		// check for group/version combo - there is only ever one `/` right?
		var group, version string
		if strings.Contains(target.ApiGroup, "/") {
			split := strings.Split(target.ApiGroup, "/")
			group = split[0]
			version = split[1]
		} else {
			group = ""
			version = target.ApiGroup
		}

		// TODO - Better way to get proper lowercase + plural of a resource
		// Pod and Pods and pod are not acceptable inputs - maybe more kubernetes native?
		resource := strings.ToLower(kind) + "s"

		items, err := kube.GetResourcesDynamically(dynamic, ctx,
			group, version, resource, target.Namespace)
		if err != nil {
			fmt.Println(err)
		}
		resources = append(resources, items...)
	}
	return resources, nil
}
