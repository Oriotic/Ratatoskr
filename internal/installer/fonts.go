package installer

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Oriotic/Ratatoskr/internal/catalog"
)

func fontDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "fonts", "ratatoskr")
}

// FontInstalled does a best-effort check for whether a font's files are
// already present, so re-running Ratatoskr doesn't re-download it.
func FontInstalled(f catalog.Font) bool {
	dir := fontDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	needle := strings.ToLower(strings.ReplaceAll(f.Name, " ", ""))
	for _, e := range entries {
		if strings.Contains(strings.ToLower(strings.ReplaceAll(e.Name(), " ", "")), needle) {
			return true
		}
	}
	return false
}

// InstallFontManually downloads a font (Nerd Font release zip, or a plain
// TTF/OTF release for non-Nerd fonts we don't have a package for) and
// unpacks it into ~/.local/share/fonts/ratatoskr, then refreshes the font cache
func InstallFontManually(f catalog.Font) ([]byte, error) {
	var log strings.Builder
	dir := fontDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return []byte(log.String()), err
	}

	if f.NerdFont && f.NFAsset != "" {
		url := fmt.Sprintf("https://github.com/ryanoasis/nerd-fonts/releases/latest/download/%s.zip", f.NFAsset)
		fmt.Fprintf(&log, "Downloading %s\n", url)
		zipPath := filepath.Join(os.TempDir(), f.NFAsset+".zip")
		if err := downloadFile(url, zipPath); err != nil {
			return []byte(log.String()), fmt.Errorf("download %s: %w", f.Name, err)
		}
		defer os.Remove(zipPath)
		if err := unzip(zipPath, dir); err != nil {
			return []byte(log.String()), fmt.Errorf("unzip %s: %w", f.Name, err)
		}
		fmt.Fprintf(&log, "Extracted to %s\n", dir)
	} else {
		fmt.Fprintf(&log, "No download source configured for %s; skipping.\n", f.Name)
	}

	refreshOut, _ := runShell("fc-cache -f " + dir)
	log.Write(refreshOut)
	return []byte(log.String()), nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("http %d fetching %s", resp.StatusCode, url)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !(strings.HasSuffix(strings.ToLower(f.Name), ".ttf") || strings.HasSuffix(strings.ToLower(f.Name), ".otf")) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		outPath := filepath.Join(dest, filepath.Base(f.Name))
		out, err := os.Create(outPath)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
