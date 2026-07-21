package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Oriotic/Ratatoskr/internal/installer"
	"github.com/Oriotic/Ratatoskr/internal/state"
)

// ---------- GPU confirm screen ----------

func (m *Model) updateGPU(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "right", "h", "l", "tab":
		m.gpuConfirm = !m.gpuConfirm
	case "y":
		m.gpuConfirm = true
	case "n":
		m.gpuConfirm = false
	case "enter":
		m.screen = scrDotfiles
	}
	return m, nil
}

func (m *Model) viewGPU() string {
	var names []string
	for _, g := range m.gpus {
		names = append(names, string(g))
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("GPU drivers") + "\n\n")
	b.WriteString(textStyle.Render(fmt.Sprintf("Detected: %s", strings.Join(names, ", "))) + "\n")
	b.WriteString(dimStyle.Render("Install the matching driver packages for these GPUs?") + "\n\n")

	yes, no := "  Yes", "  No"
	if m.gpuConfirm {
		yes = cursorStyle.Render("> ") + selectedStyle.Render("Yes")
		no = "  " + textStyle.Render("No")
	} else {
		yes = "  " + textStyle.Render("Yes")
		no = cursorStyle.Render("> ") + selectedStyle.Render("No")
	}
	b.WriteString(yes + "    " + no + "\n")
	b.WriteString("\n" + helpStyle.Render("←/→: choose   Enter: continue   q: quit"))
	return boxStyle.Render(b.String())
}

// ---------- Dotfiles screen ----------

func (m *Model) updateDotfiles(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.screen = scrSummary
	case "backspace":
		if len(m.dotfilesInput) > 0 {
			m.dotfilesInput = m.dotfilesInput[:len(m.dotfilesInput)-1]
		}
	case "ctrl+u":
		m.dotfilesInput = ""
	default:
		if len(msg.Runes) > 0 {
			m.dotfilesInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m *Model) viewDotfiles() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Dotfiles") + "\n\n")
	b.WriteString(textStyle.Render("Config repository (optional):") + "\n\n")
	b.WriteString("  " + m.dotfilesInput + cursorStyle.Render("│") + "\n\n")
	b.WriteString(dimStyle.Render("e.g. https://github.com/you/dotfiles — leave blank to skip.") + "\n")
	b.WriteString(dimStyle.Render("Known files (.zshrc, .gitconfig, .config/nvim, ...) get linked in; anything\nyou already have gets backed up first, never overwritten silently.") + "\n")
	b.WriteString("\n" + helpStyle.Render("Type to edit   Enter: continue   q: quit"))
	return boxStyle.Render(b.String())
}

// ---------- Summary screen ----------

func (m *Model) updateSummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		sel := m.selection()
		m.st.Selection = sel
		_ = m.st.Save()
		m.steps = installer.BuildSteps(sel, m.mgr)
		for _, s := range m.steps {
			m.stepOrder = append(m.stepOrder, s.ID)
			m.stepStatus[s.ID] = installer.StatusPending
		}
		ch := installer.Run(m.steps, m.st, m.mgr, m.logger)
		m.progressCh = ch
		m.screen = scrInstall
		return m, waitForUpdate(ch)
	}
	return m, nil
}

func (m *Model) viewSummary() string {
	sel := m.selection()
	var b strings.Builder
	b.WriteString(titleStyle.Render("Summary") + "\n\n")

	desktopName := "None"
	for _, d := range m.desktops {
		if d.ID == sel.Desktop {
			desktopName = d.Name
		}
	}
	b.WriteString(textStyle.Render("Desktop: ") + dimStyle.Render(desktopName) + "\n")
	b.WriteString(textStyle.Render(fmt.Sprintf("Packages: %d selected", len(sel.Components))) + "\n")
	b.WriteString(textStyle.Render(fmt.Sprintf("Fonts: %d selected", len(sel.Fonts))) + "\n")
	if sel.GPUDrivers {
		b.WriteString(textStyle.Render("GPU drivers: ") + goodStyle.Render("yes") + "\n")
	}
	if sel.DotfilesURL != "" {
		b.WriteString(textStyle.Render("Config repository: ") + dimStyle.Render(sel.DotfilesURL) + "\n")
	}
	b.WriteString(textStyle.Render(fmt.Sprintf("Estimated time: ~%d min", estimateMinutes(sel))) + "\n")

	b.WriteString("\n" + textStyle.Render("Proceed?") + "\n")
	b.WriteString("\n" + helpStyle.Render("Enter: install   q: quit"))
	return boxStyle.Render(b.String())
}

func estimateMinutes(sel state.Selection) int {
	minutes := 1
	if sel.Desktop != "" && sel.Desktop != "none" {
		minutes += 4
	}
	minutes += len(sel.Components) / 2
	minutes += len(sel.Fonts)
	if sel.GPUDrivers {
		minutes += 2
	}
	if sel.DotfilesURL != "" {
		minutes++
	}
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}
