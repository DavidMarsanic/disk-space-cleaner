package trash

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// dirTotal sums the size of every file under dir and counts its
// top-level, non-hidden entries — used by Info on platforms where Trash
// is a real, directly-readable folder (macOS, Linux).
func dirTotal(dir string) (bytes int64, items int, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		items++
	}

	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil {
			bytes += info.Size()
		}
		return nil
	})
	return bytes, items, nil
}
