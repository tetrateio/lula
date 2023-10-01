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

// Allow for the creation and return of many objects without having to define types in each location
func ManyNewOscalComponentDefinitions(data map[string][]byte) ([]oscalTypes.OscalComponentDefinitionModel, error) {

	var oscalComponentDefinitions []oscalTypes.OscalComponentDefinitionModel

	for _, item := range data {
		componentDefinition, err := NewOscalComponentDefinition(item)
		if err != nil {
			fmt.Printf("Error creating new OscalComponentDefinition: %s\n", err.Error())
		}
		oscalComponentDefinitions = append(oscalComponentDefinitions, componentDefinition)
	}

	return oscalComponentDefinitions, nil
}

// Collect all implemented-requirements from the component-definition documents
func GetImplementedRequirements(componentDefinitions []oscalTypes.OscalComponentDefinitionModel) ([]oscalTypes.ImplementedRequirement, error) {
	var implementedReqs []oscalTypes.ImplementedRequirement

	for _, componentDefinition := range componentDefinitions {
		for _, component := range componentDefinition.ComponentDefinition.Components {
			for _, controlImplementation := range component.ControlImplementations {
				implementedReqs = append(implementedReqs, controlImplementation.ImplementedRequirements...)
			}
		}
	}
	return implementedReqs, nil
}
