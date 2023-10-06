package oscal

import (
	"encoding/json"
	"fmt"

	oscalTypes "github.com/defenseunicorns/lula/src/types/oscal"
	yaml2 "github.com/ghodss/yaml"
)

// NewOscalComponentDefintion consumes a byte arrray and returns a new single OscalComponentDefinitionModel object
// Standard use is to read a file from the filesystem and pass the []byte to this function
func NewOscalComponentDefinition(data []byte) (oscalTypes.OscalComponentDefinitionModel, error) {
	var oscalComponentDefinition oscalTypes.OscalComponentDefinitionModel

	// TODO: see if we unmarshall yaml data more effectively
	jsonDoc, err := yaml2.YAMLToJSON(data)
	if err != nil {
		fmt.Printf("Error converting YAML to JSON: %s\n", err.Error())
		return oscalComponentDefinition, err
	}

	err = json.Unmarshal(jsonDoc, &oscalComponentDefinition)

	if err != nil {
		fmt.Printf("Error unmarshalling JSON: %s\n", err.Error())
	}

	return oscalComponentDefinition, nil
}

// Collect all implemented-requirements from the component-definition
func GetImplementedRequirements(componentDefinition oscalTypes.ComponentDefinition) (map[string][]oscalTypes.ImplementedRequirement, error) {
	controlImplementations := make(map[string][]oscalTypes.ImplementedRequirement, 0)

	for _, component := range componentDefinition.Components {
		for _, controlImplementation := range component.ControlImplementations {
			controlImplementations[controlImplementation.UUID] = controlImplementation.ImplementedRequirements
		}

	}
	return controlImplementations, nil
}
