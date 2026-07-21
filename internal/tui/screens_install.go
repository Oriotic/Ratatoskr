package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Oriotic/Ratatoskr/internal/installer"
)

type stepUpdateMsg installer.Update
type installDoneMsg struct{}

func waitForUpdate(ch <-chan installer.Update) tea.Cmd {
	return func() tea.Msg {
		u, ok := <-ch
		if !ok {
			return installDoneMsg{}
		}
		return stepUpdateMsg(u)
	}
}

func (m *Model) handleStepUpdate(msg stepUpdateMsg) (tea.Model, tea.Cmd) {
	u := installer.Update(msg)
	m.stepStatus[u.StepID] = u.Status

	if u.Status == installer.StatusRunning {
		m.currentTitle = u.Title
		m.currentLines = u.Explain
	}
	if u.Status == installer.StatusFailed {
		m.installErr = u.Err
		m.screen = scrDone
		if m.logCloser != nil {
			_ = m.logCloser()
		}
		return m, nil
	}
	return m, waitForUpdate(m.progressCh)
}

func (m *Model) viewInstall() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Installing") + "\n\n")

	for _, id := range m.stepOrder {
		status := m.stepStatus[id]
		title := titleFor(m.steps, id)
		switch status {
		case installer.StatusDone:
			b.WriteString(goodStyle.Render("✓ ") + textStyle.Render(title) + "\n")
		case installer.StatusSkipped:
			b.WriteString(dimStyle.Render("· "+title+" (already done)") + "\n")
		case installer.StatusRunning:
			b.WriteString(cursorStyle.Render("… ") + selectedStyle.Render(title) + "\n")
			for _, l := range m.currentLines {
				b.WriteString("    " + dimStyle.Render(l) + "\n")
			}
		case installer.StatusFailed:
			b.WriteString(badStyle.Render("✗ ") + textStyle.Render(title) + "\n")
		default:
			b.WriteString(dimStyle.Render("  "+title) + "\n")
		}
	}

	b.WriteString("\n" + helpStyle.Render("Ctrl+C: cancel (progress is saved and resumable)"))
	return boxStyle.Render(b.String())
}

func titleFor(steps []installer.Step, id string) string {
	for _, s := range steps {
		if s.ID == id {
			return s.Title
		}
	}
	return id
}

// ---------- Done screen ----------

func (m *Model) viewDone() string {
	var b strings.Builder

	if m.installErr != nil {
		b.WriteString(badStyle.Render("Setup stopped early") + "\n\n")
		b.WriteString(textStyle.Render(fmt.Sprintf("%s failed: %v", m.currentTitle, m.installErr)) + "\n\n")
		b.WriteString(dimStyle.Render("Full output is in " + logPathHint() + "\n"))
		b.WriteString(dimStyle.Render("Fix the issue and re-run Ratatoskr — completed steps will be skipped.") + "\n")
		b.WriteString("\n" + helpStyle.Render("Enter: exit"))
		return boxStyle.Render(b.String())
	}

	b.WriteString(titleStyle.Render("Setup complete.") + "\n\n")
	done, skipped := 0, 0
	for _, s := range m.stepStatus {
		switch s {
		case installer.StatusDone:
			done++
		case installer.StatusSkipped:
			skipped++
		}
	}
	b.WriteString(goodStyle.Render(fmt.Sprintf("✓ %d step(s) completed", done)) + "\n")
	if skipped > 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("· %d step(s) already satisfied, skipped", skipped)) + "\n")
	}

	sel := m.selection()
	if sel.Desktop != "" && sel.Desktop != "none" {
		b.WriteString("\n" + warnStyle.Render("A reboot is recommended to start your new desktop environment.") + "\n")
	}
	b.WriteString("\n" + textStyle.Render("Run:") + "\n")
	b.WriteString("  " + selectedStyle.Render("source ~/.zshrc") + "\n")
	b.WriteString("\n" + dimStyle.Render("Check on things anytime with: ratatoskr doctor") + "\n")

	b.WriteString("\n" + helpStyle.Render("Enter: exit"))
	return boxStyle.Render(b.String())
}

func logPathHint() string {
	return "~/.local/state/ratatoskr/ratatoskr.log"
}
