package lulatest

import (
	"bytes"
	"log"
	"os"
	"testing"
	"sync"
	"io"

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
	writers                         = []io.Writer{}
	
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

// Capture standard output
    stdoutBuf := new(bytes.Buffer)
    cmd.SetOut(stdoutBuf)

	// Use RedirectLog to capture log output
	logOutput := RedirectLog(t)

	// Log the arguments for debugging
    t.Logf("Executing command with arguments: %v", args)
    
	// Sets the Cobra command args
	cmd.SetArgs(args)

	// Execute the command and log errors
    err := cmd.Execute()

	if err != nil {
        t.Logf("Command execution error: %v", err)
    }

	// Reads the captured Log output
	capturedLog := ReadLog(t, logOutput)
	capturedStdOut := stdoutBuf.String()

	// Log the captured standard output for debugging
    t.Logf("Captured standard output: %s", capturedStdOut)

    return string(capturedLog), err
}

func ResetTestCommandState(cmd *cobra.Command) {
	cmd.ResetFlags()
	cmd.SetArgs(nil)
}

func RedirectLog(t *testing.T) *bytes.Buffer {
	logMtx.Lock()
	defer logMtx.Unlock()
	logOutput := new(bytes.Buffer)
	writers = append(writers, logOutput)
	multiWriter := io.MultiWriter(writers...)
	log.SetOutput(multiWriter)

	return logOutput
}

func ReadLog(t *testing.T, logOutput *bytes.Buffer) []byte {
	logMtx.Lock()
	defer logMtx.Unlock()
	return logOutput.Bytes()
}