package oscal_test

import (
	"testing"

	"github.com/defenseunicorns/lula/src/pkg/common/oscal"
	"gopkg.in/yaml.v3"
)

// Write a table test for the InjectJSONPathValues function
func TestInjectJSONPathValues(t *testing.T) {
	validSSPBytes := loadTestData(t, "../../../test/unit/common/oscal/valid-ssp.yaml")
	validMetadataBytes := loadTestData(t, "../../../test/unit/common/oscal/valid-metadata.yaml")

	// var validSSP oscalTypes_1_1_2.OscalCompleteSchema
	var validSSP map[string]interface{}
	if err := yaml.Unmarshal(validSSPBytes, &validSSP); err != nil {
		t.Fatalf("yaml.Unmarshal of validSSPBytes failed: %v", err)
	}

	var validMetadata map[string]interface{}
	if err := yaml.Unmarshal(validMetadataBytes, &validMetadata); err != nil {
		t.Fatalf("yaml.Unmarshal of validMetadataBytesfailed: %v", err)
	}

	tests := []struct {
		name   string
		model  map[string]interface{}
		path   string
		values map[string]interface{}
	}{
		{
			name:   "ssp-metadata",
			model:  validSSP,
			path:   "system-security-plan.metadata",
			values: validMetadata,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := oscal.InjectJSONPathValues(tt.model, tt.path, tt.values); err != nil {
				t.Errorf("InjectJSONPathValues() error = %v", err)
			}
		})
	}
}
