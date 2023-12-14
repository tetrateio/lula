package lulatest

import (
	"bytes"
	"log"
	"os"
	"testing"
	"sync"

	"github.com/spf13/cobra"

)

var (
	byteMapMtx                      = sync.Mutex{}
	logMtx                          = sync.Mutex{}
	ByteMap                         = map[string][]byte{}
	ValidComponentPath              = "../../../test/valid-component-definition.yaml"
	NoVersionComponentPath          = "../../../test/no-version-component-definition.yaml"
	InvalidVersionComponentPath     = "../../../test/invalid-version-component-definition.yaml"
	UnsupportedVersionComponentPath = "../../../test/unsupported-version-component-definition.yaml"

)

// Helper function to execute the Cobra command.
func ExecuteTestCommand(t *testing.T, root *cobra.Command, args ...string) (string, error) {
	
	// Use RedirectLog to capture log output
	logOutput := RedirectLog(t)
    
	// Sets the Cobra command args
	root.SetArgs(args)

    _, err := root.ExecuteC()

	// Reads the captured Log output
	capturedLog := ReadLog(t, logOutput)
    return string(capturedLog), err
}

func RedirectLog(t *testing.T) *bytes.Buffer {
	logOutput := new(bytes.Buffer)
	log.SetOutput(logOutput)

	t.Cleanup(func() {
		log.SetOutput(os.Stderr)
	})
	return logOutput
}

func ReadLog(t *testing.T, logOutput *bytes.Buffer) []byte {
	logMtx.Lock()
	defer logMtx.Unlock()
	return logOutput.Bytes()
}