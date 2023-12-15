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
	
	pathSlice = []string{ValidComponentPath, NoVersionComponentPath, InvalidVersionComponentPath, UnsupportedVersionComponentPath}	
)

// GetByteMap reads the files in PathSlice and stores them in ByteMap
func GetByteMap(t *testing.T) {
	byteMapMtx.Lock()
	defer byteMapMtx.Unlock()
	if len(ByteMap) == 0 {
		for _, path := range pathSlice {
			bytes, err := os.ReadFile(path)
			if err != nil {
				panic(err)
			}
			ByteMap[path] = bytes
		}
	}
}

// Function type for creating a new command instance
type CommandFactory func() *cobra.Command

// Helper function to execute the Cobra command.
func ExecuteTestCommand(t *testing.T, cmdFactory CommandFactory, args ...string) (string, error) {
	cmd := cmdFactory()
	// Use RedirectLog to capture log output
	logOutput := RedirectLog(t)
    
	// Sets the Cobra command args
	cmd.SetArgs(args)

    _, err := cmd.ExecuteC()

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