package system

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"
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

// runningMu guards runningPID: the process-group ID of whatever command
// DefaultRunner currently has in flight (0 = nothing running). Ratatoskr
// only ever runs one installer step at a time, so a single slot is enough.
var (
	runningMu  sync.Mutex
	runningPID int
)

func DefaultRunner(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Start(); err != nil {
		return buf.Bytes(), err
	}

	runningMu.Lock()
	runningPID = cmd.Process.Pid
	runningMu.Unlock()

	err := cmd.Wait()

	runningMu.Lock()
	runningPID = 0
	runningMu.Unlock()

	return buf.Bytes(), err
}

var (
	geteuid    = os.Geteuid
	runKillCmd = func(args ...string) error { return exec.Command(args[0], args[1:]...).Run() }
)


func sendGroupSignal(pid int, sig syscall.Signal, sigName string) {
	if geteuid() == 0 {
		_ = syscall.Kill(-pid, sig)
		return
	}
	_ = runKillCmd("sudo", "-n", "kill", "-"+sigName, "-"+strconv.Itoa(pid))
}

// CancelRunning terminates whatever command DefaultRunner currently has in flight
func CancelRunning() {
	runningMu.Lock()
	pid := runningPID
	runningMu.Unlock()
	if pid == 0 {
		return
	}

	sendGroupSignal(pid, syscall.SIGTERM, "TERM")

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		runningMu.Lock()
		stillRunning := runningPID == pid
		runningMu.Unlock()
		if !stillRunning {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	sendGroupSignal(pid, syscall.SIGKILL, "KILL")
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

func (m *Manager) lockPaths() []string {
	switch m.Name {
	case "apt":
		return []string{"/var/lib/dpkg/lock-frontend", "/var/lib/dpkg/lock", "/var/cache/apt/archives/lock"}
	case "dnf":
		return []string{"/var/cache/dnf/metadata_lock.pid"}
	case "pacman":
		return []string{"/var/lib/pacman/db.lck"}
	case "zypper":
		return []string{"/var/run/zypp.pid"}
	}
	return nil
}

// processNames are the process name(s) that indicate this package manager
// is still active somewhere on the system
func (m *Manager) processNames() []string {
	switch m.Name {
	case "apt":
		return []string{"apt-get", "apt", "dpkg", "unattended-upgr"}
	case "dnf":
		return []string{"dnf", "dnf5", "rpm"}
	case "pacman":
		return []string{"pacman"}
	case "zypper":
		return []string{"zypper"}
	}
	return nil
}

// CleanupStaleLock removes this package manager's lock file(s)
func (m *Manager) CleanupStaleLock(run Runner) {
	names := m.processNames()
	if len(names) == 0 {
		return
	}
	for _, n := range names {
		if _, err := run("pgrep", "-x", n); err == nil {
			return // something by that name is still alive; leave the lock alone
		}
	}
	for _, p := range m.lockPaths() {
		_, _ = run("sudo", "rm", "-f", p)
	}
}
