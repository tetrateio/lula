package validationConfig

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
)

// ExecuteValidationTemplate templates a validation with data config and env vars
// Note: run on `validate` with config data
func ExecuteValidationTemplate(config map[string]interface{}, templateString string) (result string, err error) {
	envVars := loadEnvVars()
	config["env"] = envVars["env"]

	return executeTemplate(config, templateString)
}

// ExecuteRuntimeTemplate templates a validation with env vars - assumes config data is already loaded via compose or non-existent
// IDK if this is needed, couold probably just use ExecuteValidationTemplate...
// Note: only run on `validate` with no config data, and/or after composed
func ExecuteRuntimeTemplate(templateString string) (result string, err error) {
	envVars := loadEnvVars()

	return executeTemplate(envVars, templateString)
}

// ExecuteBuildTimeTemplate just execs the template w config data - should maintain env vars templating
// Note: only run on `compose` with config data
func ExecuteBuildTimeTemplate(config map[string]interface{}, templateString string) (result string, err error) {
	// Check for env key in config
	err = checkForEnvKey(config)
	if err != nil {
		return "", err
	}

	// Find anything {{ .env.KEY }} and replace with {{ "{{ .env.KEY }}" }}
	re := regexp.MustCompile(`{{\s*\.env\.([a-zA-Z0-9_]+)\s*}}`)
	templateString = re.ReplaceAllString(templateString, "{{ \"{{ .env.$1 }}\" }}")

	return executeTemplate(config, templateString)
}

// executeTemplate templates the template string with the data map
func executeTemplate(data map[string]interface{}, templateString string) (string, error) {
	tmpl, err := template.New("template").Parse(templateString)
	if err != nil {
		return "", err
	}
	tmpl.Option("missingkey=zero")

	var buffer strings.Builder
	err = tmpl.Execute(&buffer, data)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// loadEnvVars loads all environment variables into a map[string]interface{}
func loadEnvVars() map[string]interface{} {
	envVars := make(map[string]interface{})
	envMap := make(map[string]string)

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		envMap[pair[0]] = pair[1]
	}

	envVars["env"] = envMap
	return envVars
}

// checkForEnvKey checks if the config map contains an "env" key and throws an error if found
func checkForEnvKey(data map[string]interface{}) error {
	if _, exists := data["env"]; exists {
		return fmt.Errorf("'env' key found in data map")
	}
	return nil
}
