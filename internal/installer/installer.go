package installer

import (
	"log"
	"os/exec"

	"github.com/Oriotic/Ratatoskr/internal/state"
	"github.com/Oriotic/Ratatoskr/internal/system"
)

// StepStatus is where a Step currently stands in its lifecycle.
type StepStatus int

const (
	StatusPending StepStatus = iota
	StatusRunning
	StatusSkipped // already satisfied on a prior run
	StatusDone
	StatusFailed
)

// Update is emitted onto the progress channel every time a step changes
// state, so the TUI can render it live without polling.
type Update struct {
	StepID  string
	Title   string
	Explain []string
	Status  StepStatus
	Err     error
}

// Run executes every step in order, skipping ones already marked complete
// in state (resume) or already satisfied on the live system (idempotency),
// and streams an Update for each transition on the returned channel. The
// channel is closed when every step has finished or one fails.
func Run(steps []Step, st *state.State, mgr *system.Manager, logger *log.Logger) <-chan Update {
	ch := make(chan Update)
	run := system.DefaultRunner

	go func() {
		defer close(ch)
		st.InProgress = true
		_ = st.Save()

		for _, step := range steps {
			if st.IsDone(step.ID) && step.Check(run) {
				ch <- Update{StepID: step.ID, Title: step.Title, Explain: step.Explain, Status: StatusSkipped}
				continue
			}
			if step.Check(run) {
				ch <- Update{StepID: step.ID, Title: step.Title, Explain: step.Explain, Status: StatusSkipped}
				_ = st.MarkDone(step.ID)
				continue
			}

			ch <- Update{StepID: step.ID, Title: step.Title, Explain: step.Explain, Status: StatusRunning}
			out, err := step.Run(run)
			state.LogStep(logger, step.ID, step.Title, out, err)

			if err != nil {
				ch <- Update{StepID: step.ID, Title: step.Title, Explain: step.Explain, Status: StatusFailed, Err: err}
				st.InProgress = false
				_ = st.Save()
				return
			}
			_ = st.MarkDone(step.ID)
			ch <- Update{StepID: step.ID, Title: step.Title, Explain: step.Explain, Status: StatusDone}
		}
		st.InProgress = false
		_ = st.Save()
	}()

	return ch
}

// runShell is a tiny convenience used by the font installer for
// non-privileged local commands (no sudo needed).
func runShell(cmd string) ([]byte, error) {
	return exec.Command("sh", "-c", cmd).CombinedOutput()
}
