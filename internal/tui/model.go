package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Oriotic/Ratatoskr/internal/catalog"
	"github.com/Oriotic/Ratatoskr/internal/installer"
	"github.com/Oriotic/Ratatoskr/internal/state"
	"github.com/Oriotic/Ratatoskr/internal/system"
)

type screen int

const (
	scrDetect screen = iota
	scrProfile
	scrDesktop
	scrComponents
	scrFonts
	scrGPU
	scrDotfiles
	scrSummary
	scrInstall
	scrDone
	scrQuit
)

type Model struct {
	width, height int

	// System facts, gathered once at startup.
	distro        system.Distro
	gpus          []system.GPU
	laptop        bool
	uefi          bool
	displayServer string

	mgr *system.Manager
	st  *state.State

	screen screen

	// Profile screen
	profiles      []catalog.Profile
	profileCursor int

	// Desktop screen
	desktops      []catalog.Desktop
	desktopCursor int

	// Components screen
	componentList     []catalog.Component
	componentCursor   int
	componentSelected map[string]bool

	// Fonts screen
	fontCursor   int
	fontSelected map[string]bool

	// GPU confirm screen
	gpuConfirm bool // true = yes install drivers

	// Dotfiles screen
	dotfilesInput string

	// Install screen
	steps         []installer.Step
	stepStatus    map[string]installer.StepStatus
	stepOrder     []string
	currentTitle  string
	currentLines  []string
	progressCh    <-chan installer.Update
	installErr    error
	installOutput []byte
	logger        *log.Logger
	logCloser     func() error

	quitting bool
}

// New builds the initial model: detects the system, loads persisted state, and starts on the detect screen.
func New() (*Model, error) {
	d := system.DetectDistro()
	mgr := system.NewManager(d)

	st, err := state.Load()
	if err != nil {
		return nil, err
	}

	logger, closer, err := state.NewLogger()
	if err != nil {
		return nil, err
	}

	m := &Model{
		distro:            d,
		gpus:              system.DetectGPUs(),
		laptop:            system.IsLaptop(),
		uefi:              system.UEFIBoot(),
		displayServer:     system.DisplayServer(),
		mgr:               mgr,
		st:                st,
		screen:            scrDetect,
		profiles:          catalog.Profiles,
		desktops:          catalog.Desktops,
		componentSelected: map[string]bool{},
		fontSelected:      map[string]bool{},
		stepStatus:        map[string]installer.StepStatus{},
		logger:            logger,
		logCloser:         closer,
		gpuConfirm:        len(system.DetectGPUs()) > 0,
	}
	for _, f := range catalog.Fonts {
		m.fontSelected[f.ID] = f.Default
	}
	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case stepUpdateMsg:
		return m.handleStepUpdate(msg)

	case installDoneMsg:
		m.screen = scrDone
		if m.logCloser != nil {
			_ = m.logCloser()
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.screen == scrInstall {
		// Only allow a clean interrupt while installing; don't let stray keys interfere with the running step.
		if msg.String() == "ctrl+c" {
			m.quitting = true
			if m.logCloser != nil {
				_ = m.logCloser()
			}
			return m, tea.Quit
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c", "q":
		if m.screen == scrDotfiles {
			break // let 'q' be typed into the URL field
		}
		m.quitting = true
		if m.logCloser != nil {
			_ = m.logCloser()
		}
		return m, tea.Quit
	}

	switch m.screen {
	case scrDetect:
		return m.updateDetect(msg)
	case scrProfile:
		return m.updateProfile(msg)
	case scrDesktop:
		return m.updateDesktop(msg)
	case scrComponents:
		return m.updateComponents(msg)
	case scrFonts:
		return m.updateFonts(msg)
	case scrGPU:
		return m.updateGPU(msg)
	case scrDotfiles:
		return m.updateDotfiles(msg)
	case scrSummary:
		return m.updateSummary(msg)
	case scrDone:
		if msg.String() == "enter" || msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) View() string {
	if m.quitting {
		return ""
	}
	switch m.screen {
	case scrDetect:
		return m.viewDetect()
	case scrProfile:
		return m.viewProfile()
	case scrDesktop:
		return m.viewDesktop()
	case scrComponents:
		return m.viewComponents()
	case scrFonts:
		return m.viewFonts()
	case scrGPU:
		return m.viewGPU()
	case scrDotfiles:
		return m.viewDotfiles()
	case scrSummary:
		return m.viewSummary()
	case scrInstall:
		return m.viewInstall()
	case scrDone:
		return m.viewDone()
	}
	return ""
}

// selection builds the state.Selection struct from whatever the wizard screens have gathered so far.
func (m *Model) selection() state.Selection {
	var comps []string
	for id, on := range m.componentSelected {
		if on {
			comps = append(comps, id)
		}
	}
	var fonts []string
	for id, on := range m.fontSelected {
		if on {
			fonts = append(fonts, id)
		}
	}
	profile := ""
	if m.profileCursor >= 0 && m.profileCursor < len(m.profiles) {
		profile = m.profiles[m.profileCursor].ID
	}
	desktop := ""
	if m.desktopCursor >= 0 && m.desktopCursor < len(m.desktops) {
		desktop = m.desktops[m.desktopCursor].ID
	}
	return state.Selection{
		Profile:     profile,
		Desktop:     desktop,
		Components:  comps,
		Fonts:       fonts,
		DotfilesURL: m.dotfilesInput,
		GPUDrivers:  len(m.gpus) > 0 && m.gpuConfirm,
	}
}
