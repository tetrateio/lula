package template

import (
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/defenseunicorns/pkg/helpers"
)

const PREFIX = "LULA_"

// ExecuteTemplate templates the template string with the data map
func ExecuteTemplate(data map[string]interface{}, templateString string, templateSensitive bool) ([]byte, error) {

	// if we are not templating sensitive items - replace {{ }} with temporary delimiters
	if !templateSensitive {
		templateString = replaceDelimiters(templateString)
	}

	tmpl, err := template.New("template").Parse(templateString)
	if err != nil {
		return []byte{}, err
	}
	tmpl.Option("missingkey=default")

	var buffer strings.Builder
	err = tmpl.Execute(&buffer, data)
	if err != nil {
		return []byte{}, err
	}

	result := buffer.String()

	// replace temporary delimiters
	if !templateSensitive {
		result = revertDelimiters(result)
	}

	return []byte(result), nil
}

// Prepare the map of data for use in templating

func CollectTemplatingData(data map[string]interface{}) map[string]interface{} {

	// Get all environment variables with a specific prefix
	envMap := GetEnvVars(PREFIX)

	// Merge the data into a single map for use with templating
	mergedMap := helpers.MergeMapRecursive(envMap, data)

	return mergedMap

}

// get all environment variables with the established prefix
func GetEnvVars(prefix string) map[string]interface{} {
	envMap := make(map[string]interface{})

	// Get all environment variables
	envVars := os.Environ()

	// Iterate over environment variables
	for _, envVar := range envVars {
		// Split the environment variable into key and value
		pair := strings.SplitN(envVar, "=", 2)
		if len(pair) != 2 {
			continue
		}

		key := pair[0]
		value := pair[1]

		// Check if the key starts with the specified prefix
		if strings.HasPrefix(key, prefix) {
			// Remove the prefix from the key and convert to lowercase
			strippedKey := strings.TrimPrefix(key, prefix)
			envMap[strings.ToLower(strippedKey)] = value
		}
	}

	return envMap
}

/*
\{\{\s*\.secrets\.\w+\s*\}\}

\{\{: Matches two literal curly braces {{.
\s*: Matches any amount (including zero) of whitespace (spaces, tabs, etc.).
\.secrets\.: Matches the literal string .secrets. (with the escaped dot).
\w+: Matches one or more word characters (letters, digits, and underscores).
\s*: Matches any amount of whitespace (similar to the previous \s*).
\}\}: Matches two literal curly braces }}.

The regex pattern matches a structure like {{ .secrets.<word> }}, with optional spaces around .secrets.<word>.
*/

func replaceDelimiters(input string) string {
	// Define the regex pattern
	// This checks for any match
	pattern := `\{\{\s*\.secrets\.\w+\s*\}\}`
	re := regexp.MustCompile(pattern)

	// Replace the delimiters
	result := re.ReplaceAllStringFunc(input, func(match string) string {
		// Remove the original braces and replace with ##
		// Trim any whitespace within the match to preserve the original path
		replaced := "##" + match[2:len(match)-2] + "##"
		return replaced
	})

	return result
}

func revertDelimiters(input string) string {
	// Define the regex pattern for the ## delimited strings
	pattern := `##\s*\.secrets\.\w+\s*##`
	re := regexp.MustCompile(pattern)

	// Replace the ## delimiters with {{ and }}
	result := re.ReplaceAllStringFunc(input, func(match string) string {
		// Remove the original hashes and replace with {{}}
		// Trim any whitespace within the match to preserve the original path
		reverted := "{{" + match[2:len(match)-2] + "}}"
		return reverted
	})

	return result
}

// insertNestedKeyValue takes an underscore-delimited key and a value, and puts it in a nested map.
func insertNestedKeyValue(m map[string]interface{}, key, value string) {
	// Split the key by underscores
	parts := strings.Split(key, "_")

	// Iterate through the parts and build the nested map
	currentMap := m
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		// If the key part doesn't exist, create a new map
		if _, exists := currentMap[part]; !exists {
			currentMap[part] = make(map[string]interface{})
		}
		// Move deeper into the map
		currentMap = currentMap[part].(map[string]interface{})
	}

	// Set the final key's value
	currentMap[parts[len(parts)-1]] = value
}
