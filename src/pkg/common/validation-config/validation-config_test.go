package validationConfig_test

import (
	"testing"

	validationConfig "github.com/defenseunicorns/lula/src/pkg/common/validation-config"
	"github.com/stretchr/testify/assert"
)

func TestExecuteValidationTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     map[string]interface{}
		template []byte
		expected []byte
	}{
		{
			name: "test-validation-with-config-data",
			data: map[string]interface{}{
				"name":    "test",
				"version": "1.0.0",
			},
			template: []byte(`
name: {{ .name }}
version: {{ .version }}
`),
			expected: []byte(`
name: test
version: 1.0.0
`),
		},
		{
			name: "test-validation-with-env-data",
			data: map[string]interface{}{
				"name":    "test",
				"version": "1.0.0",
			},
			template: []byte(`
name: {{ .name }}
version: {{ .version }}
username: {{ .env.USER }}
`),
			expected: []byte(`
name: test
version: 1.0.0
username: meganwolf
`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validationConfig.ExecuteValidationTemplate(tt.data, string(tt.template))
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			assert.Equal(t, result, string(tt.expected))
		})
	}
}

func TestExecuteBuildTimeTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     map[string]interface{}
		template []byte
		expected []byte
	}{
		{
			name: "test-validation-with-config-data",
			data: map[string]interface{}{
				"name":    "test",
				"version": "1.0.0",
			},
			template: []byte(`
name: {{ .name }}
version: {{ .version }}
username: {{ .env.USER }}
`),
			expected: []byte(`
name: test
version: 1.0.0
username: {{ .env.USER }}
`),
		},
		{
			name: "test-validation-differet-spacing",
			data: map[string]interface{}{
				"name":    "test",
				"version": "1.0.0",
			},
			template: []byte(`
name: {{ .name }}
version: {{ .version }}
username: {{.env.USER}}
`),
			expected: []byte(`
name: test
version: 1.0.0
username: {{ .env.USER }}
`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validationConfig.ExecuteBuildTimeTemplate(tt.data, string(tt.template))
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			assert.Equal(t, result, string(tt.expected))
		})
	}
}
