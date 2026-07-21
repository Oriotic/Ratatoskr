# Ratatoskr

A modern Linux setup assistant. Not a shell script you skim once and forget —
an interactive wizard that explains what it's doing, adapts to your distro,
and stays useful long after the first run via `doctor` and `repair`.

```
Ratatoskr
─────────────────────────────────────────

Detecting system...

✓ Arch Linux
✓ UEFI
✓ Wayland
✓ NVIDIA GPU

Choose desktop

> GNOME
  KDE Plasma
  Hyprland
  ...
```

## Features

- **Distro-aware.** Detects Debian/Ubuntu, Fedora/RHEL, Arch, and openSUSE
  families from `/etc/os-release` and drives the right package manager
  (`apt`, `dnf`, `pacman`, `zypper`) automatically.
- **Hardware-aware.** Detects UEFI vs BIOS, Wayland vs X11, GPU vendor
  (NVIDIA/AMD/Intel via `lspci`), and laptop vs desktop chassis.
- **Explains itself.** Every step shows what it's about to do and why
  before it does it — not just a spinner.
- **Idempotent & resumable.** Every step checks whether it's already
  satisfied before running. Killing Ratatoskr mid-run and re-launching it
  picks up exactly where it left off; nothing gets reinstalled needlessly.
- **Logs everything.** Full output of every command goes to
  `~/.local/state/ratatoskr/ratatoskr.log`, even though the TUI only shows
  a friendly summary.
- **Profiles.** Development, Gaming, Designer, Minimal, AI/ML, or Custom —
  pick a starting point, then fine-tune.
- **Smart dotfiles linking.** Clones your dotfiles repo and symlinks only
  recognized config files/dirs into `$HOME` — anything already there gets
  backed up to `~/.ratatoskr-backup/<timestamp>/` first, never silently
  overwritten.
- **`doctor` and `repair`.** Re-check system health anytime and fix only
  what's actually broken, using the exact same checks the installer uses.

## Install

Requires Go 1.22+ to build (nothing at runtime — it compiles to a single
static-ish binary).

```bash
git clone https://github.com/Oriotic/Ratatoskr
cd Ratatoskr
./scripts/install.sh          # builds and installs to /usr/local/bin
# or just:
./scripts/build.sh            # builds to ./bin/ratatoskr
```

## Usage

```bash
ratatoskr            # interactive setup wizard (same as `ratatoskr setup`)
ratatoskr doctor      # check the health of everything Ratatoskr manages
ratatoskr repair      # fix anything doctor reports as broken
ratatoskr version
```

### Wizard flow

1. **Detect** — distro, boot mode, display server, GPU, chassis.
2. **Profile** — Development / Gaming / Designer / Minimal / AI-ML / Custom.
   Just pre-ticks sensible defaults on the next screen.
3. **Desktop** — GNOME, KDE Plasma, Hyprland, Xfce, Sway, Cinnamon, or none.
4. **Components** — checklist of dev tools (and profile extras: Steam,
   GIMP, Python, etc. depending on your profile).
5. **Fonts** — Nerd Fonts and friends.
6. **GPU drivers** — only asked if a GPU was actually detected.
7. **Dotfiles** — optional git URL to link configs from.
8. **Summary** — recap + estimated time, then confirm.
9. **Install** — live per-step progress, skipping anything already done.

Keys: `↑/↓` or `j/k` to move, `space` to toggle a checkbox, `enter` to
continue, `q`/`ctrl+c` to quit.

### Doctor mode

```
$ ratatoskr doctor

✓ Git
✓ Docker
✗ Docker

✓ Neovim
✗ JetBrains Mono Nerd Font

Suggested fixes:

  1. Docker
  2. JetBrains Mono Nerd Font

Run:

  ratatoskr repair
```

`doctor` runs against your last saved wizard selection (or a sensible
Development-profile default if you've never run `setup`). `repair` re-runs
only the failing steps, live, and updates the log.

## Architecture

```
main.go                     CLI entry: setup / doctor / repair / version
internal/system/             distro + package manager + GPU/chassis detection
internal/catalog/            every installable thing, mapped per distro
internal/state/               persisted JSON state (idempotency/resume) + logging
internal/installer/           step runner, dotfiles linking, font downloads
internal/doctor/              health checks (reuses installer Step.Check)
internal/tui/                 the Bubble Tea wizard
```

Adding a new installable component, desktop, font, or profile means editing
exactly one file: `internal/catalog/catalog.go`. Everything else (the
wizard checklist, the installer, and `doctor`) picks it up automatically.

## Known limitations

- Some driver/package names (notably NVIDIA on Fedora and openSUSE) assume
  the relevant third-party repo (RPM Fusion, the NVIDIA repo, etc.) is
  already enabled — Ratatoskr doesn't set those up for you yet.
- Display manager selection installs one alongside the chosen desktop but
  doesn't currently ask if you'd prefer a different one than the default.
- Nerd Fonts are downloaded from GitHub releases at install time, so that
  step needs network access to `github.com`.
