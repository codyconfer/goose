package game

import tea "github.com/charmbracelet/bubbletea"

type screen interface {
	handleKey(m *Model, msg tea.KeyMsg) tea.Cmd

	view(m *Model) string

	simulates() bool
}
