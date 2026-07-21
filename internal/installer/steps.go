package installer

import (
	"fmt"

	"github.com/Oriotic/Ratatoskr/internal/catalog"
	"github.com/Oriotic/Ratatoskr/internal/state"
	"github.com/Oriotic/Ratatoskr/internal/system"
)

// Step is one unit of installation work: idempotent (Check tells us if
// it's already satisfied) and self-describing (Explain is shown live in
// the TUI so the user knows *why* something is happening, not just that
// a command is running).
type Step struct {
	ID      string
	Title   string
	Explain []string
	Check   func(run system.Runner) bool
	Run     func(run system.Runner) ([]byte, error)
	// SkipInDoctor marks steps that don't represent a meaningful health
	// state (e.g. "refresh package indexes" always looks "unhealthy"
	// because it's designed to always re-run) so `doctor`/`repair` don't
	// report them as broken.
	SkipInDoctor bool
}

// BuildSteps turns a saved Selection into the ordered list of Steps the installer will execute.
func BuildSteps(sel state.Selection, mgr *system.Manager) []Step {
	var steps []Step
	comps := catalog.AllComponents()

	steps = append(steps, Step{
		ID:           "pkg-update",
		Title:        "Refreshing package indexes",
		Explain:      []string{"Syncing the local package database so installs pull the latest versions."},
		Check:        func(run system.Runner) bool { return false }, // always safe/cheap to re-run
		Run:          func(run system.Runner) ([]byte, error) { return mgr.Update(run) },
		SkipInDoctor: true,
	})

	if sel.Desktop != "" && sel.Desktop != "none" {
		for _, d := range catalog.Desktops {
			if d.ID != sel.Desktop {
				continue
			}
			d := d
			pkgs := mgr.Pkgs(d.Pkgs)
			steps = append(steps, Step{
				ID:      "desktop-" + d.ID,
				Title:   "Installing " + d.Name,
				Explain: []string{fmt.Sprintf("Installing the %s desktop environment and its core apps.", d.Name)},
				Check:   func(run system.Runner) bool { return allInstalled(mgr, run, pkgs) },
				Run:     func(run system.Runner) ([]byte, error) { return mgr.Install(run, pkgs) },
			})
			dmPkgs := mgr.Pkgs(d.DisplayManager)
			if len(dmPkgs) > 0 && d.DMService != "" {
				svc := d.DMService
				steps = append(steps, Step{
					ID:    "displaymanager-" + d.ID,
					Title: "Setting up the " + svc + " login screen",
					Explain: []string{
						fmt.Sprintf("Installing %s so you get a graphical login prompt.", svc),
						"Enabling it as the default display manager.",
					},
					Check: func(run system.Runner) bool {
						return allInstalled(mgr, run, dmPkgs) && shellOK(run, "systemctl is-enabled --quiet "+svc)
					},
					Run: func(run system.Runner) ([]byte, error) {
						out, err := mgr.Install(run, dmPkgs)
						if err != nil {
							return out, err
						}
						enableOut, enableErr := run("sudo", "systemctl", "enable", svc)
						return append(out, enableOut...), enableErr
					},
				})
			}
		}
	}

	for _, id := range sel.Components {
		c, ok := comps[id]
		if !ok {
			continue
		}
		steps = append(steps, componentStep(c, mgr))
	}

	if sel.GPUDrivers {
		for _, g := range system.DetectGPUs() {
			set, ok := catalog.GPUDriverPkgs[g]
			if !ok {
				continue
			}
			g := g
			pkgs := mgr.Pkgs(set)
			if len(pkgs) == 0 {
				continue
			}
			steps = append(steps, Step{
				ID:      "gpu-" + string(g),
				Title:   fmt.Sprintf("Installing %s GPU drivers", g),
				Explain: []string{fmt.Sprintf("Detected a %s GPU; installing the matching driver packages.", g)},
				Check:   func(run system.Runner) bool { return allInstalled(mgr, run, pkgs) },
				Run:     func(run system.Runner) ([]byte, error) { return mgr.Install(run, pkgs) },
			})
		}
	}

	for _, fontID := range sel.Fonts {
		for _, f := range catalog.Fonts {
			if f.ID != fontID {
				continue
			}
			steps = append(steps, fontStep(f, mgr))
		}
	}

	if sel.DotfilesURL != "" {
		steps = append(steps, Step{
			ID:      "dotfiles",
			Title:   "Copying dotfiles from " + sel.DotfilesURL,
			Explain: []string{"Cloning your dotfiles repo.", "Backing up any existing files it would overwrite.", "Linking the rest into place."},
			Check:   func(run system.Runner) bool { return DotfilesMarkerMatches(sel.DotfilesURL) },
			Run:     func(run system.Runner) ([]byte, error) { return ApplyDotfiles(sel.DotfilesURL) },
		})
	}

	return steps
}

func componentStep(c catalog.Component, mgr *system.Manager) Step {
	pkgs := mgr.Pkgs(c.Pkgs)
	explain := []string{c.Description}
	if len(pkgs) > 0 {
		explain = append(explain, fmt.Sprintf("Installing package(s): %v", pkgs))
	}
	if c.PostInstall != "" {
		explain = append(explain, "Running post-install setup (service enable / installer script).")
	}
	return Step{
		ID:      "component-" + c.ID,
		Title:   "Installing " + c.Name,
		Explain: explain,
		Check: func(run system.Runner) bool {
			if c.CheckCmd != "" {
				return shellOK(run, c.CheckCmd)
			}
			return allInstalled(mgr, run, pkgs)
		},
		Run: func(run system.Runner) ([]byte, error) {
			var all []byte
			if len(pkgs) > 0 {
				out, err := mgr.Install(run, pkgs)
				all = append(all, out...)
				if err != nil {
					return all, err
				}
			}
			if c.PostInstall != "" {
				out, err := run("sh", "-c", c.PostInstall)
				all = append(all, out...)
				if err != nil {
					return all, err
				}
			}
			return all, nil
		},
	}
}

func fontStep(f catalog.Font, mgr *system.Manager) Step {
	nativePkgs := mgr.Pkgs(f.Pkgs)
	return Step{
		ID:      "font-" + f.ID,
		Title:   "Installing " + f.Name,
		Explain: []string{fontExplain(f)},
		Check: func(run system.Runner) bool {
			if len(nativePkgs) > 0 && allInstalled(mgr, run, nativePkgs) {
				return true
			}
			return FontInstalled(f)
		},
		Run: func(run system.Runner) ([]byte, error) {
			if len(nativePkgs) > 0 {
				out, err := mgr.Install(run, nativePkgs)
				if err == nil {
					return out, nil
				}
				// fall through to manual download if the distro lacks the package
			}
			return InstallFontManually(f)
		},
	}
}

func fontExplain(f catalog.Font) string {
	if f.NerdFont {
		return "Downloading the patched Nerd Font archive and installing it to ~/.local/share/fonts."
	}
	return "Installing via the distro package manager, or downloading it if unavailable."
}

func allInstalled(mgr *system.Manager, run system.Runner, pkgs []string) bool {
	if len(pkgs) == 0 {
		return true
	}
	for _, p := range pkgs {
		if p == "-t" { // zypper pattern install syntax marker, not a real pkg name
			continue
		}
		if !mgr.IsInstalled(run, p) {
			return false
		}
	}
	return true
}

func shellOK(run system.Runner, cmd string) bool {
	_, err := run("sh", "-c", cmd)
	return err == nil
}
