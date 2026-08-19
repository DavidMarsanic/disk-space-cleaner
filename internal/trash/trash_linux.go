//go:build linux

package trash

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Move implements the freedesktop.org Trash spec (the mechanism every
// Linux file manager — Nautilus, Dolphin, etc. — reads from): the file
// itself moves to ~/.local/share/Trash/files, and a matching .trashinfo
// sidecar records its original path and deletion time so "restore" works.
func Move(path string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolving home directory: %w", err)
	}
	base := filepath.Join(home, ".local", "share", "Trash")
	filesDir := filepath.Join(base, "files")
	infoDir := filepath.Join(base, "info")
	if err := os.MkdirAll(filesDir, 0o755); err != nil {
		return fmt.Errorf("preparing Trash: %w", err)
	}
	if err := os.MkdirAll(infoDir, 0o755); err != nil {
		return fmt.Errorf("preparing Trash: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	name, destFile, destInfo := uniqueDest(filesDir, infoDir, filepath.Base(path))

	info := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		url.PathEscape(absPath), time.Now().Format("2006-01-02T15:04:05"))
	if err := os.WriteFile(destInfo, []byte(info), 0o644); err != nil {
		return fmt.Errorf("writing trash info: %w", err)
	}

	if err := os.Rename(path, destFile); err != nil {
		_ = os.Remove(destInfo)
		return fmt.Errorf("moving to Trash: %w", err)
	}
	_ = name
	return nil
}

func uniqueDest(filesDir, infoDir, name string) (finalName, filePath, infoPath string) {
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	candidate := name
	for i := 1; ; i++ {
		filePath = filepath.Join(filesDir, candidate)
		infoPath = filepath.Join(infoDir, candidate+".trashinfo")
		_, fileErr := os.Stat(filePath)
		_, infoErr := os.Stat(infoPath)
		if fileErr != nil && infoErr != nil {
			return candidate, filePath, infoPath
		}
		i++
		candidate = stem + " " + strconv.Itoa(i) + ext
	}
}

// Info reports how much is currently sitting in the freedesktop.org Trash
// — the same location Move sends things to and Nautilus/Dolphin/etc show.
func Info() (bytes int64, items int, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, 0, fmt.Errorf("resolving home directory: %w", err)
	}
	return dirTotal(filepath.Join(home, ".local", "share", "Trash", "files"))
}

// Empty permanently deletes everything in the freedesktop.org Trash (both
// the file contents and their .trashinfo sidecars) — the one genuinely
// irreversible action in this app, gated behind its own explicit
// confirmation in the UI rather than living under the reversible
// "move to Trash" flow every other action here uses.
func Empty() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolving home directory: %w", err)
	}
	base := filepath.Join(home, ".local", "share", "Trash")
	for _, sub := range []string{"files", "info"} {
		entries, err := os.ReadDir(filepath.Join(base, sub))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("emptying Trash: %w", err)
		}
		for _, e := range entries {
			if err := os.RemoveAll(filepath.Join(base, sub, e.Name())); err != nil {
				return fmt.Errorf("emptying Trash: %w", err)
			}
		}
	}
	return nil
}
