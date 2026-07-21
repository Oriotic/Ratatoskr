package system

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// GPU is a detected graphics adapter vendor.
type GPU string

const (
	GPUNvidia  GPU = "NVIDIA"
	GPUAMD     GPU = "AMD"
	GPUIntel   GPU = "Intel"
	GPUUnknown GPU = ""
)

// DetectGPUs returns the set of GPU vendors found on the system, in the
// order lspci reported them. Falls back to /sys scanning if lspci is absent.
func DetectGPUs() []GPU {
	seen := map[GPU]bool{}
	var out []GPU
	add := func(g GPU) {
		if g != GPUUnknown && !seen[g] {
			seen[g] = true
			out = append(out, g)
		}
	}

	if path, err := exec.LookPath("lspci"); err == nil {
		cmd := exec.Command(path, "-mm")
		data, err := cmd.Output()
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				low := strings.ToLower(line)
				if !strings.Contains(low, "vga") && !strings.Contains(low, "3d controller") && !strings.Contains(low, "display controller") {
					continue
				}
				switch {
				case strings.Contains(low, "nvidia"):
					add(GPUNvidia)
				case strings.Contains(low, "amd"), strings.Contains(low, "ati "), strings.Contains(low, "advanced micro devices"):
					add(GPUAMD)
				case strings.Contains(low, "intel"):
					add(GPUIntel)
				}
			}
			return out
		}
	}

	// Fallback: scan /sys/class/drm vendor IDs.
	entries, err := os.ReadDir("/sys/class/drm")
	if err == nil {
		for _, e := range entries {
			vendorPath := "/sys/class/drm/" + e.Name() + "/device/vendor"
			b, err := os.ReadFile(vendorPath)
			if err != nil {
				continue
			}
			switch strings.TrimSpace(string(b)) {
			case "0x10de":
				add(GPUNvidia)
			case "0x1002":
				add(GPUAMD)
			case "0x8086":
				add(GPUIntel)
			}
		}
	}
	return out
}

// IsLaptop guesses whether the machine is a laptop using DMI chassis type
// and the presence of a battery, which is far more reliable than chassis
// type alone on many boards.
func IsLaptop() bool {
	if b, err := os.ReadFile("/sys/class/dmi/id/chassis_type"); err == nil {
		if n, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil {
			// DMTF chassis types: 8=Portable,9=Laptop,10=Notebook, 11=Handheld,14=Sub-Notebook
			switch n {
			case 8, 9, 10, 11, 14:
				return true
			}
		}
	}
	entries, err := os.ReadDir("/sys/class/power_supply")
	if err == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "BAT") {
				return true
			}
		}
	}
	return false
}
