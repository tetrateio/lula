package tui_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aymanbagabas/go-udiff"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	oscalTypes_1_1_2 "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/defenseunicorns/lula/src/internal/tui"
	"github.com/defenseunicorns/lula/src/internal/tui/common"
	"github.com/defenseunicorns/lula/src/pkg/common/oscal"
	"github.com/muesli/termenv"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
	tea.Sequence()
}

func oscalFromPath(t *testing.T, path string) *oscalTypes_1_1_2.OscalCompleteSchema {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}
	oscalModel, err := oscal.NewOscalModel(data)
	if err != nil {
		t.Fatalf("error creating oscal model from file: %v", err)
	}

	return oscalModel
}

// TestNewComponentDefinitionModel tests that the NewOSCALModel creates the expected model from component definition file
func TestNewComponentDefinitionModel(t *testing.T) {
	oscalModel := oscalFromPath(t, "../../test/unit/common/oscal/valid-component.yaml")
	model := tui.NewOSCALModel(oscalModel)
	testModel := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(common.DefaultWidth, common.DefaultHeight))

	if err := testModel.Quit(); err != nil {
		t.Fatal(err)
	}

	fm := testModel.FinalModel(t, teatest.WithFinalTimeout(time.Second*5))

	teatest.RequireEqualOutput(t, []byte(fm.View()))
}

// TestMultiComponentDefinitionModel tests that the NewOSCALModel creates the expected model from component definition file
// and checks the component selection overlay -> new component section
func TestMultiComponentDefinitionModel(t *testing.T) {
	oscalModel := oscalFromPath(t, "../../test/unit/common/oscal/valid-multi-component.yaml")
	model := tui.NewOSCALModel(oscalModel)
	testModel := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(common.DefaultWidth, common.DefaultHeight))

	testModel.Send(tea.KeyMsg{Type: tea.KeyRight}) // Select component
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter}) // enter component selection overlay
	testModel.Send(tea.KeyMsg{Type: tea.KeyDown})  // navigate down
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter}) // select new component, exit overlay
	testModel.Send(tea.KeyMsg{Type: tea.KeyRight}) // Select framework
	testModel.Send(tea.KeyMsg{Type: tea.KeyRight}) // Select control list
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter}) // Open control

	if err := testModel.Quit(); err != nil {
		t.Fatal(err)
	}

	fm := testModel.FinalModel(t, teatest.WithFinalTimeout(time.Second*5))

	teatest.RequireEqualOutput(t, []byte(fm.View()))
}

// TestNewAssessmentResultsModel tests that the NewOSCALModel creates the expected model from assessment results file
func TestNewAssessmentResultsModel(t *testing.T) {
	oscalModel := oscalFromPath(t, "../../test/unit/common/oscal/valid-assessment-results.yaml")
	model := tui.NewOSCALModel(oscalModel)
	testModel := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(common.DefaultWidth, common.DefaultHeight))

	testModel.Send(tea.KeyMsg{Type: tea.KeyTab})

	if err := testModel.Quit(); err != nil {
		t.Fatal(err)
	}

	fm := testModel.FinalModel(t, teatest.WithFinalTimeout(time.Second*5))

	teatest.RequireEqualOutput(t, []byte(fm.View()))
}

// TestComponentControlSelect tests that the user can navigate to a control, select it, and see expected
// remarks, description, and validations
// func TestComponentControlSelect(t *testing.T) {
// 	oscalModel := oscalFromPath(t, "../../test/unit/common/oscal/valid-component.yaml")
// 	model := tui.NewOSCALModel(oscalModel)
// 	testModel := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(common.DefaultWidth, common.DefaultHeight))

// 	// Navigate to the control
// 	testModel.Send(tea.KeyMsg{Type: tea.KeyRight}) // Select component
// 	testModel.Send(tea.KeyMsg{Type: tea.KeyRight}) // Select framework
// 	testModel.Send(tea.KeyMsg{Type: tea.KeyRight}) // Select control
// 	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter}) // Open control

// 	time.Sleep(time.Second * 2)

// 	if err := testModel.Quit(); err != nil {
// 		t.Fatal(err)
// 	}

// 	testModel.WaitFinished(t, teatest.WithFinalTimeout(time.Second*5))

// 	out, err := io.ReadAll(testModel.FinalOutput(t, teatest.WithFinalTimeout(time.Second*5)))
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	teatest.RequireEqualOutput(t, out)
// }

// RequireEqualEscape is a helper function to assert the given output is
// the expected from the golden files, printing its diff in case it is not.
func requireEqualEscape(tb testing.TB, out []byte, escapes bool) {
	tb.Helper()

	out = fixLineEndings(out)

	if err := os.WriteFile("test-got.golden", out, 0o600); err != nil {
		tb.Fatal(err)
	}

	golden := filepath.Join("testdata", tb.Name()+".golden")

	goldenBts, err := os.ReadFile(golden)
	if err != nil {
		tb.Fatal(err)
	}

	goldenBts = fixLineEndings(goldenBts)
	goldenStr := string(goldenBts)
	outStr := string(out)
	if escapes {
		goldenStr = escapesSeqs(goldenStr)
		outStr = escapesSeqs(outStr)
	}

	if err := os.WriteFile("test-expected.golden", []byte(goldenStr), 0o600); err != nil {
		tb.Fatal(err)
	}

	diff := udiff.Unified("golden", "run", goldenStr, outStr)
	if diff != "" {
		tb.Fatalf("output does not match, expected:\n\n%s\n\ngot:\n\n%s\n\ndiff:\n\n%s", goldenStr, outStr, diff)
	}
}

func fixLineEndings(in []byte) []byte {
	return bytes.ReplaceAll(in, []byte("\r\n"), []byte{'\n'})
}

func escapesSeqs(in string) string {
	s := strings.Split(in, "\n")
	for i, l := range s {
		q := strconv.Quote(l)
		q = strings.TrimPrefix(q, `"`)
		q = strings.TrimSuffix(q, `"`)
		s[i] = q
	}
	return strings.Join(s, "\n")
}
