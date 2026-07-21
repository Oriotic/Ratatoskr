package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Oriotic/Ratatoskr/internal/catalog"
)

// ---------- Components screen ----------

func (m *Model) updateComponents(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.componentCursor > 0 {
			m.componentCursor--
		}
	case "down", "j":
		if m.componentCursor < len(m.componentList)-1 {
			m.componentCursor++
		}
	case " ":
		if m.componentCursor < len(m.componentList) {
			id := m.componentList[m.componentCursor].ID
			m.componentSelected[id] = !m.componentSelected[id]
		}
	case "enter":
		m.screen = scrFonts
	}
	return m, nil
}

func (m *Model) viewComponents() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Development") + "\n\n")

	for i, c := range m.componentList {
		b.WriteString(checklistLine(i == m.componentCursor, m.componentSelected[c.ID], c.Name))
		if i == m.componentCursor {
			b.WriteString("      " + dimStyle.Render(c.Description) + "\n")
		}
	}
	b.WriteString("\n" + helpStyle.Render("↑/↓: move   Space: toggle   Enter: continue   q: quit"))
	return boxStyle.Render(b.String())
}

// ---------- Fonts screen ----------

func (m *Model) updateFonts(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.fontCursor > 0 {
			m.fontCursor--
		}
	case "down", "j":
		if m.fontCursor < len(catalog.Fonts)-1 {
			m.fontCursor++
		}
	case " ":
		if m.fontCursor < len(catalog.Fonts) {
			id := catalog.Fonts[m.fontCursor].ID
			m.fontSelected[id] = !m.fontSelected[id]
		}
	case "enter":
		if len(m.gpus) > 0 {
			m.screen = scrGPU
		} else {
			m.screen = scrDotfiles
		}
	}
	return m, nil
}

func (m *Model) viewFonts() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Fonts") + "\n\n")

	for i, f := range catalog.Fonts {
		b.WriteString(checklistLine(i == m.fontCursor, m.fontSelected[f.ID], f.Name))
	}
	b.WriteString("\n" + helpStyle.Render("↑/↓: move   Space: toggle   Enter: continue   q: quit"))
	return boxStyle.Render(b.String())
}

// checklistLine renders one "[x] Name" row with cursor highlighting.
func checklistLine(cursor, checked bool, name string) string {
	c := "  "
	style := textStyle
	if cursor {
		c = cursorStyle.Render("> ")
		style = selectedStyle
	}
	box := "[ ]"
	if checked {
		box = goodStyle.Render("[x]")
	}
	return c + box + " " + style.Render(name) + "\n"
}
