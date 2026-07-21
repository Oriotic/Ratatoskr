// Package doctor implements `ratatoskr doctor` / `ratatoskr repair`. It
// deliberately reuses the exact same Step.Check functions the installer
// uses, so "healthy" always means the same thing in both places.
package doctor

import (
	"strings"

	"github.com/Oriotic/Ratatoskr/internal/installer"
	"github.com/Oriotic/Ratatoskr/internal/state"
	"github.com/Oriotic/Ratatoskr/internal/system"
)

// Result is one line of `doctor` output.
type Result struct {
	StepID  string
	Title   string // human label for doctor/repair output, e.g. "Docker" not "Installing Docker"
	Healthy bool
	Step    installer.Step // kept around so `repair` can re-run it directly
}

func label(title string) string {
	for _, prefix := range []string{"Installing ", "Setting up the ", "Setting up "} {
		if strings.HasPrefix(title, prefix) {
			return strings.TrimPrefix(title, prefix)
		}
	}
	return title
}

// Run checks every step from the last saved selection against the live system and reports which are healthy.
func Run(sel state.Selection, mgr *system.Manager) []Result {
	run := system.DefaultRunner
	steps := installer.BuildSteps(sel, mgr)
	var results []Result
	for _, s := range steps {
		if s.SkipInDoctor {
			continue
		}
		results = append(results, Result{
			StepID:  s.ID,
			Title:   label(s.Title),
			Healthy: s.Check(run),
			Step:    s,
		})
	}
	return results
}

// Failing filters Run's results down to just the unhealthy ones.
func Failing(results []Result) []Result {
	var out []Result
	for _, r := range results {
		if !r.Healthy {
			out = append(out, r)
		}
	}
	return out
}
