// Package state persists what Ratatoskr has already done, so re-running it
// is safe (idempotent) and an interrupted run can resume instead of starting over
package state

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Dir returns the directory Ratatoskr keeps its state and logs in
func Dir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".local", "state", "ratatoskr")
}

func StatePath() string { return filepath.Join(Dir(), "state.json") }
func LogPath() string   { return filepath.Join(Dir(), "ratatoskr.log") }

// Selection is what the user picked in the wizard, saved so a resumed run doesn't need to ask again.
type Selection struct {
	Profile     string   `json:"profile"`
	Desktop     string   `json:"desktop"`
	Components  []string `json:"components"`
	Fonts       []string `json:"fonts"`
	DotfilesURL string   `json:"dotfiles_url"`
	GPUDrivers  bool     `json:"gpu_drivers"`
}

// State is the full on-disk record.
type State struct {
	Version        int             `json:"version"`
	StartedAt      time.Time       `json:"started_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	Selection      Selection       `json:"selection"`
	CompletedSteps map[string]bool `json:"completed_steps"`
	InProgress     bool            `json:"in_progress"`
}

// Load reads existing state from disk, or returns a fresh empty State if none exists yet
func Load() (*State, error) {
	b, err := os.ReadFile(StatePath())
	if os.IsNotExist(err) {
		return &State{Version: 1, CompletedSteps: map[string]bool{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var s State
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	if s.CompletedSteps == nil {
		s.CompletedSteps = map[string]bool{}
	}
	return &s, nil
}

// Save writes state to disk atomically-ish (write temp, rename).
func (s *State) Save() error {
	if err := os.MkdirAll(Dir(), 0o755); err != nil {
		return err
	}
	s.UpdatedAt = time.Now()
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := StatePath() + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, StatePath())
}

// MarkDone records a completed step ID and saves.
func (s *State) MarkDone(stepID string) error {
	s.CompletedSteps[stepID] = true
	return s.Save()
}

// IsDone reports whether a step ID was already completed in a prior run.
func (s *State) IsDone(stepID string) bool {
	return s.CompletedSteps[stepID]
}

// Reset clears completed-step tracking, e.g. for a deliberate re-run.
func (s *State) Reset() {
	s.CompletedSteps = map[string]bool{}
}

// NewLogger opens (creating/appending) the Ratatoskr log file and returns a
// standard logger plus a close func. Every install step's full output goes
// here, regardless of what the TUI shows on screen.
func NewLogger() (*log.Logger, func() error, error) {
	if err := os.MkdirAll(Dir(), 0o755); err != nil {
		return nil, nil, err
	}
	f, err := os.OpenFile(LogPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, err
	}
	logger := log.New(f, "", log.LstdFlags)
	logger.Println("──────── ratatoskr session start ────────")
	return logger, f.Close, nil
}

// LogStep writes a formatted section for one step's output to the log.
func LogStep(logger *log.Logger, stepID, stepName string, output []byte, err error) {
	status := "OK"
	if err != nil {
		status = fmt.Sprintf("FAILED: %v", err)
	}
	logger.Printf("[%s] %s -> %s\n", stepID, stepName, status)
	if len(output) > 0 {
		logger.Printf("---- output ----\n%s\n----------------\n", string(output))
	}
}
