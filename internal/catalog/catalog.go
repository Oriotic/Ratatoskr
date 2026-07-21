// Package catalog is the single source of truth for "what can Ratatoskr
// install, and what is it called on each distro". Keeping every mapping
// here means adding a new component or supporting a new distro touches
// exactly one file.
package catalog

import "github.com/Oriotic/Ratatoskr/internal/system"

// Component is anything the user can tick in the wizard: a dev tool, a
// desktop environment, or an extra profile package.
type Component struct {
	ID          string
	Name        string
	Description string
	Pkgs        system.PkgSet
	// PostInstall, if set, is a shell one-liner run after the package
	// install step (e.g. enabling a systemd service). Empty = none.
	PostInstall string
	// CheckCmd, if set, overrides the default "is it installed" package
	// check with a custom shell command that should exit 0 when present.
	CheckCmd string
	Default  bool // pre-ticked in the wizard
}

// Desktop is a selectable desktop environment / window manager.
type Desktop struct {
	ID   string
	Name string
	Pkgs system.PkgSet
	// DisplayManager to enable afterwards, native pkg name per family.
	DisplayManager system.PkgSet
	DMService      string // systemd service name to enable, e.g. "gdm"
}

// Font is an installable typeface.
type Font struct {
	ID       string
	Name     string
	NerdFont bool          // fetched from ryanoasis/nerd-fonts GitHub releases
	NFAsset  string        // asset base name on nerd-fonts releases, e.g. "JetBrainsMono"
	Pkgs     system.PkgSet // if set, prefer native package over download
	Default  bool
}

// Profile bundles component IDs together under a named persona.
type Profile struct {
	ID          string
	Name        string
	Description string
	Components  []string // Component IDs pulled in by default
}

var (
	debian = system.FamilyDebian
	fedora = system.FamilyFedora
	arch   = system.FamilyArch
	suse   = system.FamilySUSE
)

// Desktops is every desktop/WM Ratatoskr can install.
var Desktops = []Desktop{
	{ID: "gnome", Name: "GNOME", DMService: "gdm",
		Pkgs: system.PkgSet{
			debian: {"gnome-shell", "gnome-tweaks", "gnome-terminal", "nautilus"},
			fedora: {"@gnome-desktop"},
			arch:   {"gnome", "gnome-tweaks"},
			suse:   {"-t", "patterns-gnome-gnome"},
		},
		DisplayManager: system.PkgSet{
			debian: {"gdm3"}, fedora: {"gdm"}, arch: {"gdm"}, suse: {"gdm"},
		}},
	{ID: "kde", Name: "KDE Plasma", DMService: "sddm",
		Pkgs: system.PkgSet{
			debian: {"kde-plasma-desktop"},
			fedora: {"@kde-desktop"},
			arch:   {"plasma", "kde-applications"},
			suse:   {"-t", "patterns-kde-kde_plasma"},
		},
		DisplayManager: system.PkgSet{
			debian: {"sddm"}, fedora: {"sddm"}, arch: {"sddm"}, suse: {"sddm"},
		}},
	{ID: "hyprland", Name: "Hyprland", DMService: "sddm",
		Pkgs: system.PkgSet{
			debian: {"hyprland", "waybar", "wofi", "kitty"},
			fedora: {"hyprland", "waybar", "wofi", "kitty"},
			arch:   {"hyprland", "waybar", "wofi", "kitty"},
			suse:   {"hyprland", "waybar", "wofi", "kitty"},
		},
		DisplayManager: system.PkgSet{
			debian: {"sddm"}, fedora: {"sddm"}, arch: {"sddm"}, suse: {"sddm"},
		}},
	{ID: "xfce", Name: "Xfce", DMService: "lightdm",
		Pkgs: system.PkgSet{
			debian: {"xfce4", "xfce4-goodies"},
			fedora: {"@xfce-desktop"},
			arch:   {"xfce4", "xfce4-goodies"},
			suse:   {"-t", "patterns-xfce-xfce"},
		},
		DisplayManager: system.PkgSet{
			debian: {"lightdm"}, fedora: {"lightdm"}, arch: {"lightdm", "lightdm-gtk-greeter"}, suse: {"lightdm"},
		}},
	{ID: "sway", Name: "Sway", DMService: "sddm",
		Pkgs: system.PkgSet{
			debian: {"sway", "waybar", "wofi"},
			fedora: {"sway", "waybar", "wofi"},
			arch:   {"sway", "waybar", "wofi"},
			suse:   {"sway", "waybar", "wofi"},
		},
		DisplayManager: system.PkgSet{
			debian: {"sddm"}, fedora: {"sddm"}, arch: {"sddm"}, suse: {"sddm"},
		}},
	{ID: "cinnamon", Name: "Cinnamon", DMService: "lightdm",
		Pkgs: system.PkgSet{
			debian: {"cinnamon-desktop-environment"},
			fedora: {"@cinnamon-desktop"},
			arch:   {"cinnamon"},
			suse:   {"-t", "patterns-cinnamon-cinnamon"},
		},
		DisplayManager: system.PkgSet{
			debian: {"lightdm"}, fedora: {"lightdm"}, arch: {"lightdm", "lightdm-gtk-greeter"}, suse: {"lightdm"},
		}},
	{ID: "none", Name: "None (keep current / server, no display manager)",
		Pkgs: system.PkgSet{debian: {}, fedora: {}, arch: {}, suse: {}}},
}

// DevComponents is the default-visible development checklist.
var DevComponents = []Component{
	{ID: "git", Name: "Git", Default: true,
		Description: "Version control. Sets pull.rebase + init.defaultBranch=main.",
		Pkgs:        system.PkgSet{debian: {"git"}, fedora: {"git"}, arch: {"git"}, suse: {"git"}}},
	{ID: "docker", Name: "Docker", Default: true,
		Description: "Container engine. Adds you to the docker group and enables the service.",
		Pkgs:        system.PkgSet{debian: {"docker.io", "docker-compose-v2"}, fedora: {"docker-ce", "docker-compose-plugin"}, arch: {"docker", "docker-compose"}, suse: {"docker", "docker-compose"}},
		PostInstall: "sudo systemctl enable --now docker && sudo usermod -aG docker $USER",
		CheckCmd:    "systemctl is-active --quiet docker"},
	{ID: "neovim", Name: "Neovim", Default: true,
		Description: "Modern Vim-based editor.",
		Pkgs:        system.PkgSet{debian: {"neovim"}, fedora: {"neovim"}, arch: {"neovim"}, suse: {"neovim"}}},
	{ID: "zsh", Name: "Zsh", Default: true,
		Description: "Shell, set up with Oh My Zsh.",
		Pkgs:        system.PkgSet{debian: {"zsh"}, fedora: {"zsh"}, arch: {"zsh"}, suse: {"zsh"}}},
	{ID: "starship", Name: "Starship", Default: true,
		Description: "Fast, minimal cross-shell prompt.",
		Pkgs:        system.PkgSet{debian: {}, fedora: {}, arch: {"starship"}, suse: {}},
		CheckCmd:    "command -v starship",
		PostInstall: "command -v starship >/dev/null || curl -sS https://starship.rs/install.sh | sh -s -- -y"},
	{ID: "direnv", Name: "direnv", Default: true,
		Description: "Per-directory environment variables.",
		Pkgs:        system.PkgSet{debian: {"direnv"}, fedora: {"direnv"}, arch: {"direnv"}, suse: {"direnv"}}},
	{ID: "rust", Name: "Rust", Default: false,
		Description: "Installed via rustup, not the distro package.",
		Pkgs:        system.PkgSet{debian: {}, fedora: {}, arch: {}, suse: {}},
		CheckCmd:    "command -v rustc",
		PostInstall: "command -v rustc >/dev/null || curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y"},
	{ID: "go", Name: "Go", Default: false,
		Description: "Go toolchain from the distro repos.",
		Pkgs:        system.PkgSet{debian: {"golang-go"}, fedora: {"golang"}, arch: {"go"}, suse: {"go"}}},
	{ID: "nodejs", Name: "Node.js", Default: false,
		Description: "Node.js + npm from the distro repos.",
		Pkgs:        system.PkgSet{debian: {"nodejs", "npm"}, fedora: {"nodejs", "npm"}, arch: {"nodejs", "npm"}, suse: {"nodejs", "npm"}}},
}

// Fonts is the font checklist.
var Fonts = []Font{
	{ID: "jetbrains-mono-nf", Name: "JetBrains Mono Nerd Font", NerdFont: true, NFAsset: "JetBrainsMono", Default: true},
	{ID: "inter", Name: "Inter", Default: true,
		Pkgs: system.PkgSet{debian: {"fonts-inter"}, fedora: {"google-inter-fonts"}, arch: {"inter-font"}, suse: {}}},
	{ID: "cascadia-code-nf", Name: "Cascadia Code Nerd Font", NerdFont: true, NFAsset: "CascadiaCode", Default: false},
}

// GPUDriverPkgs maps a detected GPU vendor to per-distro driver packages.
var GPUDriverPkgs = map[system.GPU]system.PkgSet{
	system.GPUNvidia: {
		debian: {"nvidia-driver", "nvidia-settings"},
		fedora: {"akmod-nvidia"}, // requires RPM Fusion to already be enabled
		arch:   {"nvidia", "nvidia-utils", "nvidia-settings"},
		suse:   {"x11-video-nvidiaG06"},
	},
	system.GPUAMD: {
		debian: {"mesa-vulkan-drivers", "libgl1-mesa-dri"},
		fedora: {"mesa-dri-drivers", "mesa-vulkan-drivers"},
		arch:   {"mesa", "vulkan-radeon"},
		suse:   {"Mesa-dri", "Mesa-vulkan-drivers"},
	},
	system.GPUIntel: {
		debian: {"mesa-vulkan-drivers", "intel-media-va-driver"},
		fedora: {"mesa-dri-drivers", "intel-media-driver"},
		arch:   {"mesa", "vulkan-intel", "intel-media-driver"},
		suse:   {"Mesa-dri", "intel-media-driver"},
	},
}

// GamingComponents are extra packages pulled in by the Gaming profile.
var GamingComponents = []Component{
	{ID: "steam", Name: "Steam", Description: "Game store & Proton compatibility layer.",
		Pkgs: system.PkgSet{debian: {"steam"}, fedora: {"steam"}, arch: {"steam"}, suse: {"steam"}}},
	{ID: "mangohud", Name: "MangoHud", Description: "In-game performance overlay.",
		Pkgs: system.PkgSet{debian: {"mangohud"}, fedora: {"mangohud"}, arch: {"mangohud"}, suse: {"mangohud"}}},
	{ID: "lutris", Name: "Lutris", Description: "Open gaming platform for non-Steam titles.",
		Pkgs: system.PkgSet{debian: {"lutris"}, fedora: {"lutris"}, arch: {"lutris"}, suse: {"lutris"}}},
	{ID: "gamemode", Name: "GameMode", Description: "Optimizes system performance on demand for games.",
		Pkgs: system.PkgSet{debian: {"gamemode"}, fedora: {"gamemode"}, arch: {"gamemode"}, suse: {"gamemode"}}},
}

// DesignerComponents are extra packages for the Designer profile.
var DesignerComponents = []Component{
	{ID: "gimp", Name: "GIMP", Description: "Raster image editor.",
		Pkgs: system.PkgSet{debian: {"gimp"}, fedora: {"gimp"}, arch: {"gimp"}, suse: {"gimp"}}},
	{ID: "inkscape", Name: "Inkscape", Description: "Vector graphics editor.",
		Pkgs: system.PkgSet{debian: {"inkscape"}, fedora: {"inkscape"}, arch: {"inkscape"}, suse: {"inkscape"}}},
	{ID: "blender", Name: "Blender", Description: "3D modeling, animation, and rendering.",
		Pkgs: system.PkgSet{debian: {"blender"}, fedora: {"blender"}, arch: {"blender"}, suse: {"blender"}}},
	{ID: "krita", Name: "Krita", Description: "Digital painting.",
		Pkgs: system.PkgSet{debian: {"krita"}, fedora: {"krita"}, arch: {"krita"}, suse: {"krita"}}},
}

// AIMLComponents are extra packages for the AI/ML profile.
var AIMLComponents = []Component{
	{ID: "python3", Name: "Python 3 + pip", Description: "Interpreter and package installer.",
		Pkgs: system.PkgSet{debian: {"python3", "python3-pip", "python3-venv"}, fedora: {"python3", "python3-pip"}, arch: {"python", "python-pip"}, suse: {"python3", "python3-pip"}}},
	{ID: "pipx", Name: "pipx", Description: "Isolated installs of Python CLI tools.",
		Pkgs: system.PkgSet{debian: {"pipx"}, fedora: {"pipx"}, arch: {"python-pipx"}, suse: {"python3-pipx"}}},
	{ID: "jupyter", Name: "JupyterLab (via pipx)", Description: "Notebook environment, installed as an isolated app.",
		Pkgs:        system.PkgSet{debian: {}, fedora: {}, arch: {}, suse: {}},
		CheckCmd:    "command -v jupyter-lab",
		PostInstall: "command -v jupyter-lab >/dev/null || pipx install jupyterlab"},
}

// Profiles are the "nice extras" personas from the spec.
var Profiles = []Profile{
	{ID: "development", Name: "Development", Description: "Editors, git, containers, language toolchains.",
		Components: []string{"git", "docker", "neovim", "zsh", "starship", "direnv"}},
	{ID: "gaming", Name: "Gaming", Description: "Steam, Proton tooling, and performance overlays.",
		Components: []string{"git", "zsh", "starship", "steam", "mangohud", "lutris", "gamemode"}},
	{ID: "designer", Name: "Designer", Description: "Creative suite for image, vector, and 3D work.",
		Components: []string{"git", "zsh", "gimp", "inkscape", "blender", "krita"}},
	{ID: "minimal", Name: "Minimal", Description: "Just the essentials: git and a shell.",
		Components: []string{"git", "zsh"}},
	{ID: "aiml", Name: "AI / ML", Description: "Python, isolated tooling, and notebooks.",
		Components: []string{"git", "docker", "zsh", "python3", "pipx", "jupyter"}},
	{ID: "custom", Name: "Custom", Description: "Pick everything yourself.",
		Components: []string{}},
}

// AllComponents returns every installable component (dev + gaming +
// designer + AI/ML) keyed by ID, for looking up a Profile's members.
func AllComponents() map[string]Component {
	m := map[string]Component{}
	for _, c := range DevComponents {
		m[c.ID] = c
	}
	for _, c := range GamingComponents {
		m[c.ID] = c
	}
	for _, c := range DesignerComponents {
		m[c.ID] = c
	}
	for _, c := range AIMLComponents {
		m[c.ID] = c
	}
	return m
}
