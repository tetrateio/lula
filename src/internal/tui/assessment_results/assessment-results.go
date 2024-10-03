package assessmentresults

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	blist "github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	oscalTypes_1_1_2 "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/defenseunicorns/lula/src/internal/tui/common"
	"github.com/evertras/bubble-table/table"
)

type Satisfaction string

const (
	height           = 20
	width            = 12
	pickerHeight     = 20
	pickerWidth      = 80
	dialogFixedWidth = 40
)

const (
	resultPicker                 common.PickerKind = "result"
	comparedResultPicker         common.PickerKind = "compared result"
	columnKeyName                                  = "name"
	columnKeyStatus                                = "status"
	columnKeyDescription                           = "description"
	columnKeyStatusChange                          = "status_change"
	columnKeyFinding                               = "finding"
	columnKeyComparedFinding                       = "compared_finding"
	columnKeyObservation                           = "observation"
	columnKeyComparedObservation                   = "compared_observation"

	satisfied    Satisfaction = "satisfied"
	notSatisfied Satisfaction = "not-satisfied"
)

var (
	styleBase = lipgloss.NewStyle().
			Foreground(common.Text).
			BorderForeground(common.Highlight).
			Align(lipgloss.Left)

	satisfiedColors = map[Satisfaction]string{
		satisfied:    "#3ad33c",
		notSatisfied: "e36750",
	}
)

func NewAssessmentResultsModel(assessmentResults *oscalTypes_1_1_2.AssessmentResults) Model {
	results := make([]result, 0)
	findings := make([]blist.Item, 0)
	var selectedResult result

	if assessmentResults != nil {
		for _, r := range assessmentResults.Results {
			var findingsRows []table.Row
			var observationsRows []table.Row
			for _, f := range *r.Findings {
				findingString, err := common.ToYamlString(f)
				if err != nil {
					common.PrintToLog("error converting finding to yaml: %v", err)
					findingString = ""
				}
				findingsRows = append(findingsRows, table.NewRow(table.RowData{
					columnKeyName:        f.Target.TargetId,
					columnKeyStatus:      f.Target.Status.State,
					columnKeyDescription: strings.ReplaceAll(f.Description, "\n", " "),
					columnKeyFinding:     findingString,
				}))
			}
			for _, o := range *r.Observations {
				state := "undefined"
				var remarks strings.Builder
				if o.RelevantEvidence != nil {
					for _, e := range *o.RelevantEvidence {
						if e.Description == "Result: satisfied\n" {
							state = "satisfied"
						} else if e.Description == "Result: not-satisfied\n" {
							state = "not-satisfied"
						}
						if e.Remarks != "" {
							remarks.WriteString(strings.ReplaceAll(e.Remarks, "\n", " "))
						}
					}
				}
				observationsRows = append(observationsRows, table.NewRow(table.RowData{
					columnKeyName:        o.Description,
					columnKeyStatus:      state,
					columnKeyDescription: remarks.String()}))
			}

			results = append(results, result{
				uuid:             r.UUID,
				title:            r.Title,
				timestamp:        r.Start.Format(time.RFC3339),
				findings:         r.Findings,
				observations:     r.Observations,
				findingsRows:     findingsRows,
				observationsRows: observationsRows,
			})
		}
	}

	// TODO - fix this so that findings table is associated to the selected result; also results table
	// possibly need to add another table type for the "comparison" table
	if len(results) != 0 {
		selectedResult = results[0]
		observationMap := makeObservationMap(selectedResult.observations)
		if selectedResult.findings != nil {
			for _, f := range *selectedResult.findings {
				// get the related observations
				observations := make([]observation, 0)
				if f.RelatedObservations != nil {
					for _, o := range *f.RelatedObservations {
						observationUuid := o.ObservationUuid
						if _, ok := observationMap[observationUuid]; ok {
							observations = append(observations, observationMap[observationUuid])
						}
					}
				}
				findings = append(findings, finding{
					title:        f.Title,
					uuid:         f.UUID,
					controlId:    f.Target.TargetId,
					state:        f.Target.Status.State,
					observations: observations,
				})
			}
		}
	}

	resultItems := make([]string, len(results))
	for i, c := range results {
		resultItems[i] = getResultText(c)
	}
	resultsPicker := common.NewPickerModel("Select a Result", resultPicker, resultItems, 0)

	comparedResultItems := getComparedResults(results, selectedResult)
	comparedResultsPicker := common.NewPickerModel("Select a Result to Compare", comparedResultPicker, comparedResultItems, 0)

	findingsSummary := viewport.New(width, height)
	findingsSummary.Style = common.PanelStyle
	observationsSummary := viewport.New(width, height)
	observationsSummary.Style = common.PanelStyle

	help := common.NewHelpModel(false)
	help.OneLine = true
	help.ShortHelp = shortHelpNoFocus

	findingsTableColumns := []table.Column{
		table.NewFlexColumn(columnKeyName, "Control", 1).WithFiltered(true),
		table.NewFlexColumn(columnKeyStatus, "Status", 1),
		table.NewFlexColumn(columnKeyDescription, "Description", 4),
	}

	observationsTableColumns := []table.Column{
		table.NewFlexColumn(columnKeyName, "Observation", 2).WithFiltered(true),
		table.NewFlexColumn(columnKeyStatus, "Status", 1),
		table.NewFlexColumn(columnKeyDescription, "Remarks", 3),
	}

	findingsTable := table.New(findingsTableColumns).
		WithRows(selectedResult.findingsRows).
		WithBaseStyle(styleBase).
		Filtered(true).
		SortByAsc(columnKeyName)

	observationsTable := table.New(observationsTableColumns).WithRows(selectedResult.observationsRows).Filtered(true)

	return Model{
		keys:                  assessmentKeys,
		help:                  help,
		results:               results,
		resultsPicker:         resultsPicker,
		comparedResultsPicker: comparedResultsPicker,
		selectedResult:        selectedResult,
		findingsTable:         findingsTable,
		findingsSummary:       findingsSummary,
		observationsTable:     observationsTable,
		observationsSummary:   observationsSummary,
		detailView:            common.NewDetailModel(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.UpdateSizing(msg.Height-common.TabOffset, msg.Width)

	case tea.KeyMsg:
		if m.open {
			common.DumpToLog(msg)
			k := msg.String()
			switch k {
			case common.ContainsKey(k, m.keys.Help.Keys()):
				m.help.ShowAll = !m.help.ShowAll

			case common.ContainsKey(k, m.keys.NavigateLeft.Keys()):
				if !m.resultsPicker.Open && !m.comparedResultsPicker.Open && !m.detailView.Open {
					if m.focus == 0 {
						m.focus = maxFocus
					} else {
						m.focus--
					}
					m.updateKeyBindings()
				}

			case common.ContainsKey(k, m.keys.NavigateRight.Keys()):
				if !m.resultsPicker.Open && !m.comparedResultsPicker.Open && !m.detailView.Open {
					m.focus = (m.focus + 1) % (maxFocus + 1)
					m.updateKeyBindings()
				}

			case common.ContainsKey(k, m.keys.Confirm.Keys()):
				switch m.focus {
				case focusResultSelection:
					if len(m.results) > 1 && !m.resultsPicker.Open {
						return m, func() tea.Msg {
							return common.PickerOpenMsg{
								Kind: resultPicker,
							}
						}
					}

				case focusCompareSelection:
					if len(m.results) > 1 && m.comparedResultsPicker.Open {
						return m, func() tea.Msg {
							return common.PickerOpenMsg{
								Kind: comparedResultPicker,
							}
						}
					}

				case focusSummary:
					// do stuff
					// Update the selected findings, update the observations table
					// If 'd', pull up detailed view

				}

			case common.ContainsKey(k, m.keys.Detail.Keys()):
				common.PrintToLog("detail key pressed")
				switch m.focus {
				case focusSummary:
					selected := m.findingsTable.HighlightedRow().Data[columnKeyFinding].(string)
					common.PrintToLog("selected: \n%s", selected)
					return m, func() tea.Msg {
						return common.DetailOpenMsg{
							Content: selected,
							Height:  (m.height + common.TabOffset),
							Width:   m.width,
						}
					}
				}
			}
		}
	}

	mdl, cmd := m.resultsPicker.Update(msg)
	m.resultsPicker = mdl.(common.PickerModel)
	cmds = append(cmds, cmd)

	mdl, cmd = m.comparedResultsPicker.Update(msg)
	m.comparedResultsPicker = mdl.(common.PickerModel)
	cmds = append(cmds, cmd)

	mdl, cmd = m.detailView.Update(msg)
	m.detailView = mdl.(common.DetailModel)
	cmds = append(cmds, cmd)

	m.findingsTable, cmd = m.findingsTable.Update(msg)
	cmds = append(cmds, cmd)

	m.observationsTable, cmd = m.observationsTable.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.resultsPicker.Open {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.resultsPicker.View(), lipgloss.WithWhitespaceChars(" "))
	}
	if m.comparedResultsPicker.Open {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.comparedResultsPicker.View(), lipgloss.WithWhitespaceChars(" "))
	}
	if m.detailView.Open {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.detailView.View(), lipgloss.WithWhitespaceChars(" "))
	}
	return m.mainView()
}

func (m Model) mainView() string {
	// Add help panel at the top left
	helpStyle := common.HelpStyle(m.width)
	helpView := helpStyle.Render(m.help.View())

	// Add viewport styles
	focusedViewport := common.PanelStyle.BorderForeground(common.Focused)
	focusedViewportHeaderColor := common.Focused
	focusedDialogBox := common.DialogBoxStyle.BorderForeground(common.Focused)

	selectedResultDialogBox := common.DialogBoxStyle
	comparedResultDialogBox := common.DialogBoxStyle
	summaryViewport := common.PanelStyle
	summaryViewportHeader := common.Highlight
	observationsViewport := common.PanelStyle
	observationsViewportHeader := common.Highlight

	switch m.focus {
	case focusResultSelection:
		selectedResultDialogBox = focusedDialogBox
	case focusCompareSelection:
		comparedResultDialogBox = focusedDialogBox
	case focusSummary:
		summaryViewport = focusedViewport
		summaryViewportHeader = focusedViewportHeaderColor
	case focusObservations:
		observationsViewport = focusedViewport
		observationsViewportHeader = focusedViewportHeaderColor
	}

	// add panels at the top for selecting a result, selecting a comparison result
	const dialogFixedWidth = 40

	selectedResultLabel := common.LabelStyle.Render("Selected Result")
	selectedResultText := common.TruncateText(getResultText(m.selectedResult), dialogFixedWidth)
	selectedResultContent := selectedResultDialogBox.Width(dialogFixedWidth).Render(selectedResultText)
	selectedResult := lipgloss.JoinHorizontal(lipgloss.Top, selectedResultLabel, selectedResultContent)

	comparedResultLabel := common.LabelStyle.Render("Compare Result")
	comparedResultText := common.TruncateText(getResultText(m.comparedResult), dialogFixedWidth)
	comparedResultContent := comparedResultDialogBox.Width(dialogFixedWidth).Render(comparedResultText)
	comparedResult := lipgloss.JoinHorizontal(lipgloss.Top, comparedResultLabel, comparedResultContent)

	resultSelectionContent := lipgloss.JoinHorizontal(lipgloss.Top, selectedResult, comparedResult)

	// Add Tables
	m.findingsSummary.Style = summaryViewport
	m.findingsSummary.SetContent(m.findingsTable.View())
	summaryPanel := fmt.Sprintf("%s\n%s", common.HeaderView("Summary", m.findingsSummary.Width-common.PanelStyle.GetPaddingRight(), summaryViewportHeader), m.findingsSummary.View())

	m.observationsSummary.Style = observationsViewport
	m.observationsSummary.SetContent(m.observationsTable.View())
	observationsPanel := fmt.Sprintf("%s\n%s", common.HeaderView("Observations", m.observationsSummary.Width-common.PanelStyle.GetPaddingRight(), observationsViewportHeader), m.observationsSummary.View())

	bottomContent := lipgloss.JoinVertical(lipgloss.Top, summaryPanel, observationsPanel)

	return lipgloss.JoinVertical(lipgloss.Top, helpView, resultSelectionContent, bottomContent)
}

// func (m *Model) renderSummary() string {
// 	return m.findingsTable.View()
// 	// return "⚠️ Summary Under Construction ⚠️"
// }

// func (m *Model) renderObservations() string {
// 	return m.observationsTable.View()
// 	// return "⚠️ Observations Under Construction ⚠️"
// }

func getComparedResults(results []result, selectedResult result) []string {
	comparedResults := []string{"No Compared Result"}
	for _, r := range results {
		if r.uuid != selectedResult.uuid {
			comparedResults = append(comparedResults, getResultText(r))
		}
	}
	return comparedResults
}

func getResultText(result result) string {
	if result.uuid == "" {
		return "No Result Selected"
	}
	return fmt.Sprintf("%s - %s", result.title, result.timestamp)
}

func makeObservationMap(observations *[]oscalTypes_1_1_2.Observation) map[string]observation {
	observationMap := make(map[string]observation)

	for _, o := range *observations {
		validationId := findUuid(o.Description)
		state := "not-satisfied"
		remarks := strings.Builder{}
		if o.RelevantEvidence != nil {
			for _, re := range *o.RelevantEvidence {
				if re.Description == "Result: satisfied\n" {
					state = "satisfied"
				} else if re.Description == "Result: not-satisfied\n" {
					state = "not-satisfied"
				}
				remarks.WriteString(re.Remarks)
			}
		}
		observationMap[o.UUID] = observation{
			uuid:         o.UUID,
			description:  o.Description,
			remarks:      remarks.String(),
			state:        state,
			validationId: validationId,
		}
	}
	return observationMap
}

func findUuid(input string) string {
	uuidPattern := `[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}`

	re := regexp.MustCompile(uuidPattern)

	return re.FindString(input)
}
