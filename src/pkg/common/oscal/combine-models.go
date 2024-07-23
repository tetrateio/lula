package oscal

import (
	"encoding/json"
	"fmt"

	oscalTypes_1_1_2 "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
)

// CombineOscalModels takes a slice of OSCAL models and combines them into a single model
func CombineOscalModels(models []*oscalTypes_1_1_2.OscalModels) (*oscalTypes_1_1_2.OscalModels, error) {
	if len(models) == 0 {
		return nil, fmt.Errorf("no models provided to combine")
	}

	mapSlice := make([]map[string]interface{}, len(models))
	for i, model := range models {
		modelMap, err := convertOscalModelToMap(*model)
		if err != nil {
			return nil, err
		}
		mapSlice[i] = modelMap
	}

	combinedMap, err := CombineMaps(mapSlice)
	if err != nil {
		return nil, fmt.Errorf("error combining the models: %v", err)
	}

	combinedModel, err := convertMapToOscalModel(combinedMap)
	if err != nil {
		return nil, fmt.Errorf("error converting the combined map to a model: %v", err)
	}

	return combinedModel, nil
}

// CombineMaps takes a slice of map[string]interface{} and combines them into a single map
func CombineMaps(maps []map[string]interface{}) (map[string]interface{}, error) {
	if len(maps) == 0 {
		return nil, fmt.Errorf("no models provided to combine")
	}

	combindedMap := maps[0]

	// Starting with the first model, merge each model into the combined model
	for _, m := range maps[1:] {
		combindedMap = merge(combindedMap, m)
	}

	return combindedMap, nil
}

// convertOscalModelToMap converts an OSCAL model to a map[string]interface{}
func convertOscalModelToMap(model oscalTypes_1_1_2.OscalModels) (map[string]interface{}, error) {
	var modelMap map[string]interface{}
	modelBytes, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(modelBytes, &modelMap)
	if err != nil {
		return nil, err
	}

	return modelMap, nil
}

// convertMapToOscalModel converts a map[string]interface{} to an OSCAL model
func convertMapToOscalModel(modelMap map[string]interface{}) (*oscalTypes_1_1_2.OscalModels, error) {
	var model oscalTypes_1_1_2.OscalModels
	modelBytes, err := json.Marshal(modelMap)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(modelBytes, &model)
	if err != nil {
		return nil, err
	}

	return &model, nil
}

// merge recursively combines two maps into a single map
func merge(dst, src map[string]interface{}) map[string]interface{} {
	for key, value := range src {
		if dstValue, ok := dst[key]; ok {
			switch dstValueTyped := dstValue.(type) {
			case map[string]interface{}:
				srcValueTyped, ok := value.(map[string]interface{})
				if ok {
					dst[key] = merge(dstValueTyped, srcValueTyped)
					continue
				}
			}
		}
		dst[key] = value
	}
	return dst
}
