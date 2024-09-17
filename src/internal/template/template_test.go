package template_test

import (
	"os"
	"strings"
	"testing"

	"github.com/defenseunicorns/lula/src/internal/template"
)

func TestExecuteTemplate(t *testing.T) {

	test := func(t *testing.T, data map[string]interface{}, sensitive bool, preTemplate string, expected string) {
		t.Helper()
		// templateData returned
		got, err := template.ExecuteTemplate(data, preTemplate, sensitive)
		if err != nil {
			t.Fatalf("error templating data: %s\n", err.Error())
		}

		if string(got) != expected {
			t.Fatalf("Expected %s - Got %s\n", expected, string(got))
		}
	}

	t.Run("Test {{ .testVar }} with data", func(t *testing.T) {
		data := map[string]interface{}{
			"testVar": "testing",
		}

		test(t, data, true, "{{ .testVar }}", "testing")
	})

	t.Run("Test {{ .testVar }} but empty data", func(t *testing.T) {
		data := map[string]interface{}{}

		test(t, data, true, "{{ .testVar }}", "<no value>")
	})

	t.Run("Sensitive templating with data - sensitive true", func(t *testing.T) {
		// represent .secrets.lulakey
		data := map[string]interface{}{
			"secrets": map[string]interface{}{
				"lulakey": "lulavalue",
			},
		}

		test(t, data, true, "{{ .secrets.lulakey }}", "lulavalue")

	})

	t.Run("Sensitive templating with data - sensitive false", func(t *testing.T) {
		// represent .secrets.lulakey
		data := map[string]interface{}{
			"secrets": map[string]interface{}{
				"lulakey": "lulavalue",
			},
		}

		test(t, data, false, "{{ .secrets.lulakey }}", "{{ .secrets.lulakey }}")

	})

	t.Run("Sensitive templating with no data - sensitive true", func(t *testing.T) {
		// represent .secrets.lulakey
		data := map[string]interface{}{}

		test(t, data, true, "{{ .secrets.lulakey }}", "<no value>")

	})

	t.Run("Sensitive templating with no data - sensitive false", func(t *testing.T) {
		// represent .secrets.lulakey
		data := map[string]interface{}{}

		test(t, data, false, "{{ .secrets.lulakey }}", "{{ .secrets.lulakey }}")

	})
}

func TestGetEnvVars(t *testing.T) {

	test := func(t *testing.T, prefix string, key string, value string) {
		t.Helper()

		os.Setenv(key, value)
		envMap := template.GetEnvVars(prefix)

		// convert key to expected format
		strippedKey := strings.TrimPrefix(key, prefix)

		if envMap[strings.ToLower(strippedKey)] != value {
			t.Fatalf("Expected %s - Got %s\n", value, envMap[strings.ToLower(strippedKey)])
		}
		os.Unsetenv(key)
	}

	t.Run("Test LULA_RESOURCE - Pass", func(t *testing.T) {
		test(t, "LULA_", "LULA_RESOURCE", "pods")
	})

	t.Run("Test OTHER_RESOURCE - Pass", func(t *testing.T) {
		test(t, "OTHER_", "OTHER_RESOURCE", "deployments")
	})
}
