package trash

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestMove exercises the real Move implementation for this OS against an
// isolated fake $HOME, never the real one. Move is safe to fully
// automate everywhere: on darwin/linux it's a plain os.Rename respecting
// $HOME, and its destination folder never collides with anything real.
func TestMove(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Move on Windows goes through the real Recycle Bin via PowerShell, not an isolatable $HOME — no Windows environment available to verify this against")
	}

	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	src := filepath.Join(fakeHome, "some-cache")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "data.bin"), make([]byte, 12345), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Move(src); err != nil {
		t.Fatalf("Move: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("expected source to be gone after Move, got err=%v", err)
	}
}

// TestInfoAndEmpty covers Info/Empty only on platforms where they're
// actually isolatable from the developer's real Trash:
//
//   - linux: both go through plain files under $HOME, which t.Setenv can
//     safely redirect, so this runs the full Move -> Info -> Empty -> Info
//     cycle against a fake Trash.
//   - darwin: skipped here. Info and Empty go through Finder via
//     AppleScript specifically because macOS blocks direct filesystem
//     access to the real ~/.Trash for third-party processes (verified
//     live: even `ls ~/.Trash` fails with "operation not permitted"
//     without Full Disk Access) — and Finder automation, unlike direct
//     file access, ignores the $HOME env var and always acts on the
//     real logged-in user's real Trash. That makes it unsafe to
//     automate here. What was verified by hand instead: Info's read
//     query (count + summed size of real Trash contents) run live and
//     cross-checked against a known probe file; Empty's AppleScript
//     (`tell application "Finder" to empty the trash`) compiled cleanly
//     against this machine's real Finder dictionary via `osacompile`
//     (syntax-checked, never executed) rather than run against real data.
//   - windows: skipped — Clear-RecycleBin isn't isolatable by $HOME
//     either, and no Windows environment is available to verify against.
func TestInfoAndEmpty(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Info/Empty aren't isolatable from the real Trash on this OS — see doc comment")
	}

	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	src := filepath.Join(fakeHome, "some-cache")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "data.bin"), make([]byte, 12345), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Move(src); err != nil {
		t.Fatalf("Move: %v", err)
	}

	bytes, items, err := Info()
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if items != 1 {
		t.Fatalf("expected 1 item in trash, got %d", items)
	}
	if bytes != 12345 {
		t.Fatalf("expected 12345 bytes in trash, got %d", bytes)
	}

	if err := Empty(); err != nil {
		t.Fatalf("Empty: %v", err)
	}
	bytes, items, err = Info()
	if err != nil {
		t.Fatalf("Info after Empty: %v", err)
	}
	if items != 0 || bytes != 0 {
		t.Fatalf("expected empty trash after Empty, got bytes=%d items=%d", bytes, items)
	}
}
