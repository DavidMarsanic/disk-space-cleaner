//go:build darwin

// Package trash moves files to the OS's real, recoverable trash/recycle
// bin. Move is the only way this app removes anything on its own —
// reversible, exactly like dragging something to Trash yourself. Empty
// is the one deliberate exception: it empties the Trash itself, which is
// unavoidably permanent, and every caller of it is expected to get its
// own explicit confirmation first.
package trash

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Move sends path to ~/.Trash, the same place Finder's "Move to Trash"
// puts things — no AppleScript/Finder automation needed, and no
// permission prompt it would trigger. A name collision (something already
// in Trash with that name) gets a numeric suffix rather than overwriting
// whatever's already there.
func Move(path string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolving home directory: %w", err)
	}
	trashDir := filepath.Join(home, ".Trash")
	if err := os.MkdirAll(trashDir, 0o755); err != nil {
		return fmt.Errorf("preparing Trash: %w", err)
	}

	dest := uniqueDest(trashDir, filepath.Base(path))
	if err := os.Rename(path, dest); err != nil {
		return fmt.Errorf("moving to Trash: %w", err)
	}
	return nil
}

func uniqueDest(dir, name string) string {
	candidate := filepath.Join(dir, name)
	if _, err := os.Stat(candidate); err != nil {
		return candidate
	}
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	for i := 2; ; i++ {
		candidate = filepath.Join(dir, stem+" "+strconv.Itoa(i)+ext)
		if _, err := os.Stat(candidate); err != nil {
			return candidate
		}
	}
}

// Info reports how much is currently sitting in ~/.Trash. This goes
// through Finder via AppleScript rather than reading the folder
// directly — verified live: a plain os.ReadDir(~/.Trash) (and even `ls`
// from a Terminal without Full Disk Access) fails with "operation not
// permitted" under macOS's privacy protections, even though Move's
// os.Rename *into* the same folder works fine. Finder itself is exempt
// from that restriction, so asking it is the only reliable way to read
// Trash's contents without requiring the user to grant this app Full
// Disk Access.
func Info() (totalBytes int64, itemCount int, err error) {
	out, err := runAppleScript(`tell application "Finder"
set totalSize to 0
set totalCount to 0
repeat with theItem in (get items of trash)
	set totalSize to totalSize + (size of theItem)
	set totalCount to totalCount + 1
end repeat
return (totalSize as string) & " " & (totalCount as string)
end tell`)
	if err != nil {
		return 0, 0, fmt.Errorf("reading Trash: %w", err)
	}
	parts := strings.Fields(out)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("reading Trash: unexpected response %q", out)
	}
	fmt.Sscanf(parts[0], "%d", &totalBytes)
	fmt.Sscanf(parts[1], "%d", &itemCount)
	return totalBytes, itemCount, nil
}

// Empty permanently deletes everything in the Trash via Finder's own
// "empty trash" command — the one genuinely irreversible action in this
// app, gated behind its own explicit confirmation in the UI rather than
// living under the reversible "move to Trash" flow every other action
// here uses. Going through Finder (rather than os.RemoveAll on ~/.Trash's
// contents) sidesteps the same read-permission restriction Info works
// around above, and it's the exact mechanism Finder's own "Empty Trash"
// menu item uses.
func Empty() error {
	_, err := runAppleScript(`tell application "Finder" to empty the trash`)
	if err != nil {
		return fmt.Errorf("emptying Trash: %w", err)
	}
	return nil
}

func runAppleScript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(out)), nil
}
