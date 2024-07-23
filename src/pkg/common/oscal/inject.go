package oscal

import (
	"fmt"

	oscalTypes_1_1_2 "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"k8s.io/client-go/util/jsonpath"
)

// InjectJSONPathValues injects values into an OSCAL model using JSONPath
func InjectJSONPathValues(model *oscalTypes_1_1_2.OscalModels, path string, values map[string]interface{}) error {
	jp := getJsonPath(path)
	if jp == nil {
		return fmt.Errorf("failed to create jsonpath from %s", path)
	}
	jp.Execute(model, values) // this isn't right.
	return nil
}

// getJsonPath returns the jsonpath from a string
func getJsonPath(path string) *jsonpath.JSONPath {
	return jsonpath.New(path)
}
