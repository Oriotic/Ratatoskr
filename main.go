// Command ratatoskr is a friendly, idempotent setup assistant for fresh
// Linux installs. Run with no arguments (or "setup") for the interactive
// wizard; "doctor" to check system health; "repair" to fix anything
// doctor found wrong.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Oriotic/Ratatoskr/internal/catalog"
	"github.com/Oriotic/Ratatoskr/internal/doctor"
	"github.com/Oriotic/Ratatoskr/internal/installer"
	"github.com/Oriotic/Ratatoskr/internal/state"
	"github.com/Oriotic/Ratatoskr/internal/system"
	"github.com/Oriotic/Ratatoskr/internal/tui"
	"github.com/Oriotic/Ratatoskr/internal/version"
)

var (
	good = lipgloss.NewStyle().Foreground(lipgloss.Color("#86EFAC")).Bold(true)
	bad  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FCA5A5")).Bold(true)
	dim  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	acc  = lipgloss.NewStyle().Foreground(lipgloss.Color("#7DD3FC")).Bold(true)
)

func main() {
	cmd := "setup"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "setup", "":
		runSetup()
	case "doctor":
		runDoctor()
	case "repair":
		runRepair()
	case "version", "-v", "--version":
		fmt.Println("ratatoskr " + version.Version)
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("Unknown command %q\n\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`ratatoskr - a Linux setup assistant

Usage:
  ratatoskr [setup]   Run the interactive setup wizard (default)
  ratatoskr doctor    Check the health of everything Ratatoskr manages
  ratatoskr repair    Fix anything "doctor" reports as broken
  ratatoskr version   Print the version
  ratatoskr help      Show this help`)
}

// ensureSudo makes sure we have a live sudo credential 
func ensureSudo() {
	if os.Geteuid() == 0 {
		return // already root; sudo isn't part of the picture
	}
	if _, err := exec.LookPath("sudo"); err != nil {
		fmt.Fprintln(os.Stderr, "ratatoskr: this needs sudo, but it isn't installed/available")
		os.Exit(1)
	}

	fmt.Println(acc.Render("Ratatoskr") + dim.Render(" needs administrator access to install packages."))
	fmt.Println()

	cmd := exec.Command("sudo", "-v")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "\nratatoskr: couldn't get sudo access, stopping.")
		os.Exit(1)
	}
	fmt.Println()

	// Keep the credential alive for the rest of the run.
	go func() {
		for {
			time.Sleep(4 * time.Minute)
			_ = exec.Command("sudo", "-n", "-v").Run()
		}
	}()
}

func runSetup() {
	ensureSudo()
	m, err := tui.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ratatoskr: failed to start:", err)
		os.Exit(1)
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "ratatoskr: ", err)
		os.Exit(1)
	}
}

// loadSelectionForHealthCheck returns the last wizard selection if one
// exists, or a sensible default (the Development profile's defaults) so
// `doctor`/`repair` are useful even before `setup` has ever been run.
func loadSelectionForHealthCheck() (*state.State, state.Selection) {
	st, err := state.Load()
	if err != nil {
		st = &state.State{CompletedSteps: map[string]bool{}}
	}
	sel := st.Selection
	if sel.Desktop == "" && len(sel.Components) == 0 {
		for _, c := range catalog.DevComponents {
			if c.Default {
				sel.Components = append(sel.Components, c.ID)
			}
		}
		for _, f := range catalog.Fonts {
			if f.Default {
				sel.Fonts = append(sel.Fonts, f.ID)
			}
		}
	}
	return st, sel
}

func runDoctor() {
	d := system.DetectDistro()
	mgr := system.NewManager(d)
	_, sel := loadSelectionForHealthCheck()

	fmt.Println(acc.Render("Ratatoskr doctor") + dim.Render("  ·  "+d.Name))
	fmt.Println()

	results := doctor.Run(sel, mgr)
	for _, r := range results {
		if r.Healthy {
			fmt.Println(good.Render("✓ ") + r.Title)
		} else {
			fmt.Println(bad.Render("✗ ") + r.Title)
		}
	}

	failing := doctor.Failing(results)
	if len(failing) == 0 {
		fmt.Println()
		fmt.Println(good.Render("Everything looks healthy."))
		return
	}

	fmt.Println()
	fmt.Println("Suggested fixes:")
	fmt.Println()
	for i, r := range failing {
		fmt.Printf("  %d. %s\n", i+1, r.Title)
	}
	fmt.Println()
	fmt.Println("Run:")
	fmt.Println()
	fmt.Println("  " + acc.Render("ratatoskr repair"))
}

func runRepair() {
	d := system.DetectDistro()
	mgr := system.NewManager(d)
	st, sel := loadSelectionForHealthCheck()

	fmt.Println(acc.Render("Ratatoskr repair") + dim.Render("  ·  "+d.Name))
	fmt.Println()

	results := doctor.Run(sel, mgr)
	failing := doctor.Failing(results)
	if len(failing) == 0 {
		fmt.Println(good.Render("Nothing to repair — everything is already healthy."))
		return
	}

	ensureSudo() // only ask once we know there's actually something to install

	logger, closer, err := state.NewLogger()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ratatoskr: could not open log:", err)
		os.Exit(1)
	}
	defer closer()

	var steps []installer.Step
	for _, r := range failing {
		steps = append(steps, r.Step)
		delete(st.CompletedSteps, r.Step.ID)
	}

	ch := installer.Run(steps, st, mgr, logger)
	failed := false
	for u := range ch {
		switch u.Status {
		case installer.StatusRunning:
			fmt.Println(acc.Render("→ ") + u.Title)
			for _, l := range u.Explain {
				fmt.Println("    " + dim.Render(l))
			}
		case installer.StatusDone:
			fmt.Println(good.Render("✓ ") + u.Title)
		case installer.StatusSkipped:
			fmt.Println(dim.Render("· " + u.Title + " (already done)"))
		case installer.StatusFailed:
			fmt.Println(bad.Render("✗ ") + u.Title + ": " + u.Err.Error())
			failed = true
		}
	}

	fmt.Println()
	if failed {
		fmt.Println(bad.Render("Repair stopped early. See ~/.local/state/ratatoskr/ratatoskr.log for details."))
		os.Exit(1)
	}
	fmt.Println(good.Render("Repair complete."))
}
