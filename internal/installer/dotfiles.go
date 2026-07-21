package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func dotfilesCacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "ratatoskr", "dotfiles")
}

func dotfilesMarkerPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "ratatoskr", "dotfiles-applied")
}

// DotfilesMarkerExists reports whether a previous run already applied a dotfiles repo (idempotency check)
func DotfilesMarkerExists() bool {
	_, err := os.Stat(dotfilesMarkerPath())
	return err == nil
}

// knownDotfileNames are the config files/dirs Ratatoskr will link from a
// dotfiles repo into $HOME if present, rather than blindly copying
// everything (which would also copy the repo's README, LICENSE, etc).
var knownDotfileNames = []string{
	".zshrc", ".bashrc", ".bash_profile", ".profile",
	".gitconfig", ".gitignore_global",
	".config/nvim", ".config/starship.toml", ".config/kitty",
	".config/alacritty", ".config/hypr", ".config/waybar", ".config/wofi",
	".config/direnv", ".tmux.conf", ".config/tmux",
	".vimrc", ".p10k.zsh",
}

// ApplyDotfiles clones (or pulls) a dotfiles git repo and symlinks the
// recognized files/directories into $HOME, backing up anything that would
// be overwritten first.
func ApplyDotfiles(repoURL string) ([]byte, error) {
	var log strings.Builder
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cacheDir := dotfilesCacheDir()

	if _, err := os.Stat(filepath.Join(cacheDir, ".git")); err == nil {
		fmt.Fprintf(&log, "Repo already cloned, pulling latest changes.\n")
		out, err := exec.Command("git", "-C", cacheDir, "pull", "--ff-only").CombinedOutput()
		log.Write(out)
		if err != nil {
			return []byte(log.String()), fmt.Errorf("git pull: %w", err)
		}
	} else {
		fmt.Fprintf(&log, "Cloning %s\n", repoURL)
		if err := os.MkdirAll(filepath.Dir(cacheDir), 0o755); err != nil {
			return []byte(log.String()), err
		}
		out, err := exec.Command("git", "clone", "--depth", "1", repoURL, cacheDir).CombinedOutput()
		log.Write(out)
		if err != nil {
			return []byte(log.String()), fmt.Errorf("git clone: %w", err)
		}
	}

	backupDir := filepath.Join(home, ".ratatoskr-backup", time.Now().Format("20060102-150405"))
	linked := 0
	for _, rel := range knownDotfileNames {
		src := filepath.Join(cacheDir, rel)
		if _, err := os.Stat(src); err != nil {
			continue // repo doesn't have this one, skip quietly
		}
		dst := filepath.Join(home, rel)

		if info, err := os.Lstat(dst); err == nil {
			// Already a symlink into our cache dir: nothing to do.
			if info.Mode()&os.ModeSymlink != 0 {
				if target, err := os.Readlink(dst); err == nil && target == src {
					continue
				}
			}
			if err := os.MkdirAll(backupDir, 0o755); err != nil {
				return []byte(log.String()), err
			}
			backupDst := filepath.Join(backupDir, rel)
			if err := os.MkdirAll(filepath.Dir(backupDst), 0o755); err != nil {
				return []byte(log.String()), err
			}
			if err := os.Rename(dst, backupDst); err != nil {
				fmt.Fprintf(&log, "warning: could not back up existing %s: %v\n", rel, err)
				continue
			}
			fmt.Fprintf(&log, "Backed up existing %s -> %s\n", rel, backupDst)
		}

		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return []byte(log.String()), err
		}
		if err := os.Symlink(src, dst); err != nil {
			fmt.Fprintf(&log, "warning: could not link %s: %v\n", rel, err)
			continue
		}
		fmt.Fprintf(&log, "Linked %s\n", rel)
		linked++
	}
	fmt.Fprintf(&log, "Linked %d dotfile(s) from %s\n", linked, repoURL)

	if err := os.MkdirAll(filepath.Dir(dotfilesMarkerPath()), 0o755); err != nil {
		return []byte(log.String()), err
	}
	if err := os.WriteFile(dotfilesMarkerPath(), []byte(repoURL), 0o644); err != nil {
		return []byte(log.String()), err
	}

	return []byte(log.String()), nil
}
