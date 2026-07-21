package system

import (
	"fmt"
	"os"
	"os/exec"
)

// PkgSet maps a distro Family to the package name(s)
type PkgSet map[Family][]string

// Manager is the uniform interface Ratatoskr uses to talk to whatever
// native package manager the host distro has.
type Manager struct {
	Name   string
	Family Family
}

// NewManager builds a Manager for the detected distro.
func NewManager(d Distro) *Manager {
	return &Manager{Name: d.PkgMgr, Family: d.Family}
}

// Runner lets the installer capture combined stdout+stderr for logging
// while still streaming nothing to the terminal (the TUI owns the screen).
type Runner func(name string, args ...string) ([]byte, error)

// DefaultRunner shells out for real.
func DefaultRunner(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	return cmd.CombinedOutput()
}

// Update refreshes package indexes.
func (m *Manager) Update(run Runner) ([]byte, error) {
	switch m.Name {
	case "apt":
		return run("sudo", "apt-get", "update", "-y")
	case "dnf":
		return run("sudo", "dnf", "check-update", "-y")
	case "pacman":
		return run("sudo", "pacman", "-Sy", "--noconfirm")
	case "zypper":
		return run("sudo", "zypper", "--non-interactive", "refresh")
	}
	return nil, fmt.Errorf("no package manager detected")
}

// installs one or more packages by native name for this distro.
func (m *Manager) Install(run Runner, pkgs []string) ([]byte, error) {
	if len(pkgs) == 0 {
		return nil, nil
	}
	switch m.Name {
	case "apt":
		args := append([]string{"apt-get", "install", "-y"}, pkgs...)
		return run("sudo", args...)
	case "dnf":
		args := append([]string{"dnf", "install", "-y"}, pkgs...)
		return run("sudo", args...)
	case "pacman":
		args := append([]string{"pacman", "-S", "--noconfirm", "--needed"}, pkgs...)
		return run("sudo", args...)
	case "zypper":
		args := append([]string{"zypper", "--non-interactive", "install"}, pkgs...)
		return run("sudo", args...)
	}
	return nil, fmt.Errorf("no package manager detected")
}

// IsInstalled checks whether a native package name is already installed.
func (m *Manager) IsInstalled(run Runner, pkg string) bool {
	var out []byte
	var err error
	switch m.Name {
	case "apt":
		out, err = run("dpkg-query", "-W", "-f=${Status}", pkg)
		if err != nil {
			return false
		}
		return containsInstalledOk(string(out))
	case "dnf":
		_, err = run("rpm", "-q", pkg)
		return err == nil
	case "pacman":
		_, err = run("pacman", "-Q", pkg)
		return err == nil
	case "zypper":
		_, err = run("rpm", "-q", pkg)
		return err == nil
	}
	return false
}

func containsInstalledOk(status string) bool {
	for i := 0; i+9 <= len(status); i++ {
		if status[i:i+9] == "installed" {
			return true
		}
	}
	return false
}

// Pkgs resolves a PkgSet to the concrete package list for this manager's family.
func (m *Manager) Pkgs(set PkgSet) []string {
	return set[m.Family]
}
