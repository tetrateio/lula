package configuration

import (
	"github.com/brandtkeller/text-template/text/template"
	"regexp"
	"strings"
)

// executeTemplate templates the template string with the data map
func ExecuteTemplate(data map[string]interface{}, templateString string) ([]byte, error) {
	tmpl, err := template.New("template").Parse(templateString)
	if err != nil {
		return []byte{}, err
	}
	tmpl.Option("missingkey=ignore")

	var buffer strings.Builder
	err = tmpl.Execute(&buffer, data)
	if err != nil {
		return []byte{}, err
	}

	return []byte(buffer.String()), nil
}

func ReplaceDelimiters(input string) string {
	// Define the regex pattern
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

func RevertDelimiters(input string) string {
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
