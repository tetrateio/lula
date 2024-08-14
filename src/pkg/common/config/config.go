package configuration

import (
	"strings"
	"text/template"
)

// executeTemplate templates the template string with the data map
func ExecuteTemplate(data map[string]interface{}, templateString string) ([]byte, error) {
	tmpl, err := template.New("template").Funcs(funcMap()).Parse(templateString)
	if err != nil {
		return []byte{}, err
	}
	tmpl.Option("missingkey=zero")

	var buffer strings.Builder
	err = tmpl.Execute(&buffer, data)
	if err != nil {
		return []byte{}, err
	}

	return []byte(buffer.String()), nil
}
