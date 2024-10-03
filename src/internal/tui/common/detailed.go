package common

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ContentType string

const (
	detailWidth  = 80
	detailHeight = 20
)

type DetailOpenMsg struct {
	Content string
	Height  int
	Width   int
}

type DetailModel struct {
	Open            bool
	help            HelpModel
	contentViewport viewport.Model
	content         string
	width           int
	height          int
}

func NewDetailModel() DetailModel {
	help := NewHelpModel(true)
	help.ShortHelp = ShortHelpDetail

	return DetailModel{
		help:            help,
		contentViewport: viewport.New(detailWidth, detailHeight-2),
	}
}

func (m DetailModel) Init() tea.Cmd {
	return nil
}

func (m DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.UpdateSizing(int(float64(msg.Height)*0.9), int(float64(msg.Width)*0.9))

	case tea.KeyMsg:
		k := msg.String()
		if m.Open {
			switch k {
			case ContainsKey(k, CommonKeys.Cancel.Keys()):
				m.Open = false
			}
		}

	case DetailOpenMsg:
		m.Open = true
		m.content = msg.Content
		m.UpdateSizing(int(float64(msg.Height)*0.9), int(float64(msg.Width)*0.9))
	}

	m.contentViewport, cmd = m.contentViewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m DetailModel) View() string {
	overlayDetailStyle := OverlayStyle.
		Width(m.width).
		Height(m.height)

	m.contentViewport.SetContent(m.content)

	detailContent := lipgloss.JoinVertical(lipgloss.Top, m.contentViewport.View(), m.help.View())
	return overlayDetailStyle.Render(detailContent)
}

func (m *DetailModel) UpdateSizing(height, width int) {
	m.height = height
	m.width = width

	m.contentViewport.Height = height - 2
	m.contentViewport.Width = width - 2
}
