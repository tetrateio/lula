package tui_test

import (
	"io"
	"os"
	"testing"
	"time"

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

func TestNewOSCALModel(t *testing.T) {
	oscalModel := oscalFromPath(t, "../../test/unit/common/oscal/valid-component.yaml")
	model := tui.NewOSCALModel(oscalModel)
	testModel := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(common.DefaultWidth, common.DefaultHeight))

	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})

	out, err := io.ReadAll(testModel.FinalOutput(t, teatest.WithFinalTimeout(time.Second*5)))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestComponentControlSelect(t *testing.T) {
	oscalModel := oscalFromPath(t, "../../test/unit/common/oscal/valid-component.yaml")
	model := tui.NewOSCALModel(oscalModel)
	testModel := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(common.DefaultWidth, common.DefaultHeight))

	// Navigate to the control
	testModel.Send(tea.KeyMsg{Type: tea.KeyRight}) // Select component
	testModel.Send(tea.KeyMsg{Type: tea.KeyRight}) // Select framework
	testModel.Send(tea.KeyMsg{Type: tea.KeyRight}) // Select control
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter}) // Open control

	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})

	out, err := io.ReadAll(testModel.FinalOutput(t, teatest.WithFinalTimeout(time.Second*5)))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}
