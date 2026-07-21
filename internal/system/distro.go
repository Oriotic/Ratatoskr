package system

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
)

// Family is the broad package-manager family a distro belongs to.
type Family string

const (
	FamilyDebian  Family = "debian"
	FamilyFedora  Family = "fedora"
	FamilyArch    Family = "arch"
	FamilySUSE    Family = "suse"
	FamilyUnknown Family = "unknown"
)

// Distro holds everything we learned about the running Linux system.
type Distro struct {
	ID        string // e.g. "arch", "ubuntu", "fedora", "opensuse-tumbleweed"
	Name      string // pretty name, e.g. "Arch Linux"
	VersionID string
	Family    Family
	PkgMgr    string // apt, dnf, pacman, zypper
}

// DetectDistro reads /etc/os-release and figures out the package-manager family.
func DetectDistro() Distro {
	d := Distro{ID: "unknown", Name: "Unknown Linux", Family: FamilyUnknown}

	f, err := os.Open("/etc/os-release")
	if err == nil {
		defer f.Close()
		vals := map[string]string{}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := sc.Text()
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := parts[0]
			val := strings.Trim(parts[1], `"`)
			vals[key] = val
		}
		if v, ok := vals["ID"]; ok {
			d.ID = v
		}
		if v, ok := vals["NAME"]; ok {
			d.Name = v
		}
		if v, ok := vals["PRETTY_NAME"]; ok {
			d.Name = v
		}
		if v, ok := vals["VERSION_ID"]; ok {
			d.VersionID = v
		}
		idLike := vals["ID_LIKE"]

		d.Family = classify(d.ID, idLike)
	}

	d.PkgMgr = pickPkgMgr(d.Family)
	return d
}

func classify(id, idLike string) Family {
	hay := strings.ToLower(id + " " + idLike)
	switch {
	case strings.Contains(hay, "arch"), strings.Contains(hay, "manjaro"), strings.Contains(hay, "endeavouros"):
		return FamilyArch
	case strings.Contains(hay, "fedora"), strings.Contains(hay, "rhel"), strings.Contains(hay, "centos"), strings.Contains(hay, "rocky"), strings.Contains(hay, "alma"):
		return FamilyFedora
	case strings.Contains(hay, "suse"):
		return FamilySUSE
	case strings.Contains(hay, "debian"), strings.Contains(hay, "ubuntu"), strings.Contains(hay, "mint"), strings.Contains(hay, "pop"):
		return FamilyDebian
	default:
		return FamilyUnknown
	}
}

func pickPkgMgr(fam Family) string {
	switch fam {
	case FamilyDebian:
		return "apt"
	case FamilyFedora:
		return "dnf"
	case FamilyArch:
		return "pacman"
	case FamilySUSE:
		return "zypper"
	}
	for _, mgr := range []string{"apt", "dnf", "pacman", "zypper"} {
		if _, err := exec.LookPath(mgr); err == nil {
			return mgr
		}
	}
	return ""
}

// UEFIBoot reports whether the system booted via UEFI.
func UEFIBoot() bool {
	_, err := os.Stat("/sys/firmware/efi")
	return err == nil
}

// DisplayServer returns "Wayland", "X11", or "None (TTY)".
func DisplayServer() string {
	if os.Getenv("WAYLAND_DISPLAY") != "" || os.Getenv("XDG_SESSION_TYPE") == "wayland" {
		return "Wayland"
	}
	if os.Getenv("DISPLAY") != "" || os.Getenv("XDG_SESSION_TYPE") == "x11" {
		return "X11"
	}
	return "None (TTY)"
}
