package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Oriotic/Ratatoskr/internal/catalog"
)

// ---------- Detect screen ----------

func (m *Model) updateDetect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		m.screen = scrProfile
	}
	return m, nil
}

func (m *Model) viewDetect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Ratatoskr") + "\n")
	b.WriteString(rule(41) + "\n\n")
	b.WriteString(textStyle.Render("Detecting system...") + "\n\n")

	line := func(ok bool, label string) string {
		mark := goodStyle.Render("✓")
		if !ok {
			mark = dimStyle.Render("·")
		}
		return fmt.Sprintf("%s %s", mark, textStyle.Render(label))
	}

	b.WriteString(line(true, m.distro.Name) + "\n")
	b.WriteString(line(m.uefi, bootMode(m.uefi)) + "\n")
	b.WriteString(line(true, m.displayServer) + "\n")
	if len(m.gpus) > 0 {
		var names []string
		for _, g := range m.gpus {
			names = append(names, string(g))
		}
		b.WriteString(line(true, strings.Join(names, " + ")+" GPU") + "\n")
	} else {
		b.WriteString(line(false, "No dedicated GPU detected") + "\n")
	}
	chassis := "Desktop"
	if m.laptop {
		chassis = "Laptop"
	}
	b.WriteString(line(true, chassis) + "\n")
	b.WriteString(line(true, "Package manager: "+m.mgr.Name) + "\n")

	if m.st.InProgress {
		b.WriteString("\n" + warnStyle.Render("A previous run looks unfinished — Ratatoskr will pick up where it left off and skip anything already done.") + "\n")
	}

	b.WriteString("\n" + helpStyle.Render("Enter: continue   q: quit"))
	return boxStyle.Render(b.String())
}

func bootMode(uefi bool) string {
	if uefi {
		return "UEFI"
	}
	return "Legacy BIOS"
}

// ---------- Profile screen ----------

func (m *Model) updateProfile(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.profileCursor > 0 {
			m.profileCursor--
		}
	case "down", "j":
		if m.profileCursor < len(m.profiles)-1 {
			m.profileCursor++
		}
	case "enter":
		m.applyProfileDefaults()
		m.screen = scrDesktop
	}
	return m, nil
}

// applyProfileDefaults pre-ticks the components checklist based on the
// chosen profile, and builds the combined component list to show.
func (m *Model) applyProfileDefaults() {
	p := m.profiles[m.profileCursor]

	seen := map[string]bool{}
	var list []catalog.Component
	add := func(cs []catalog.Component) {
		for _, c := range cs {
			if !seen[c.ID] {
				seen[c.ID] = true
				list = append(list, c)
			}
		}
	}
	add(catalog.DevComponents)
	switch p.ID {
	case "gaming":
		add(catalog.GamingComponents)
	case "designer":
		add(catalog.DesignerComponents)
	case "aiml":
		add(catalog.AIMLComponents)
	}
	m.componentList = list

	inProfile := map[string]bool{}
	for _, id := range p.Components {
		inProfile[id] = true
	}
	for _, c := range list {
		if p.ID == "custom" {
			m.componentSelected[c.ID] = c.Default
		} else {
			m.componentSelected[c.ID] = inProfile[c.ID]
		}
	}
}

func (m *Model) viewProfile() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Choose a profile") + "\n")
	b.WriteString(subtitleStyle.Render("This just pre-selects sensible defaults — you can fine-tune everything next.") + "\n\n")

	for i, p := range m.profiles {
		cursor := "  "
		style := textStyle
		if i == m.profileCursor {
			cursor = cursorStyle.Render("> ")
			style = selectedStyle
		}
		b.WriteString(cursor + style.Render(p.Name) + "\n")
		if i == m.profileCursor {
			b.WriteString("    " + dimStyle.Render(p.Description) + "\n")
		}
	}
	b.WriteString("\n" + helpStyle.Render("↑/↓: move   Enter: continue   q: quit"))
	return boxStyle.Render(b.String())
}

// ---------- Desktop screen ----------

func (m *Model) updateDesktop(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.desktopCursor > 0 {
			m.desktopCursor--
		}
	case "down", "j":
		if m.desktopCursor < len(m.desktops)-1 {
			m.desktopCursor++
		}
	case "enter":
		m.screen = scrComponents
	}
	return m, nil
}

func (m *Model) viewDesktop() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Choose desktop") + "\n\n")

	for i, d := range m.desktops {
		cursor := "  "
		style := textStyle
		if i == m.desktopCursor {
			cursor = cursorStyle.Render("> ")
			style = selectedStyle
		}
		b.WriteString(cursor + style.Render(d.Name) + "\n")
	}
	b.WriteString("\n" + helpStyle.Render("↑/↓: move   Enter: continue   q: quit"))
	return boxStyle.Render(b.String())
}
