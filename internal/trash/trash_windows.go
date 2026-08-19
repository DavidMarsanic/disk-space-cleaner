//go:build windows

package trash

import (
	"fmt"
	"os/exec"
	"strings"
)

// Move sends path to the Recycle Bin via
// Microsoft.VisualBasic.FileIO.FileSystem.DeleteFile, rather than hand-
// rolling the SHFileOperationW struct layout directly — a subtly wrong
// field offset there corrupts an OS call in a way that's very hard to
// diagnose, and this codebase has no Windows environment to verify it
// against. The VB FileIO helper is the standard, documented way to reach
// the Recycle Bin from a script.
func Move(path string) error {
	script := fmt.Sprintf(
		`Add-Type -AssemblyName Microsoft.VisualBasic; `+
			`[Microsoft.VisualBasic.FileIO.FileSystem]::DeleteFile('%s', `+
			`'OnlyErrorDialogs', 'SendToRecycleBin')`,
		strings.ReplaceAll(path, "'", "''"),
	)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("moving to Recycle Bin: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// Info can't report a real size on Windows: the Recycle Bin is a hidden,
// per-volume, per-user-SID structure (`$Recycle.Bin\<SID>`) with no
// simple, reliably-permissioned folder to just list and sum — unlike
// macOS/Linux, where Trash is a plain folder Move already writes to
// directly. Returning -1 tells the UI "size unknown" rather than
// guessing. Emptying is still safe and well-defined (see Empty), so
// that's offered without a precomputed size.
func Info() (bytes int64, items int, err error) {
	return -1, -1, nil
}

// Empty empties the Recycle Bin via PowerShell's built-in Clear-RecycleBin
// cmdlet (Windows 10+) — the one genuinely irreversible action in this
// app, gated behind its own explicit confirmation in the UI rather than
// living under the reversible "move to Trash" flow every other action
// here uses.
func Empty() error {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command",
		"Clear-RecycleBin -Force -ErrorAction Stop")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("emptying Recycle Bin: %s", strings.TrimSpace(string(out)))
	}
	return nil
}
