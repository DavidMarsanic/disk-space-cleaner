// Package engine finds well-known, documented-safe cache directories
// (build tools, package managers, browsers) and reports how much space
// each is using — never anything outside those specific, hardcoded
// locations under the user's own home directory, and never anything
// requiring elevated permissions.
package engine

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Progress is one step of a scan, streamed to the caller as it happens —
// walking a directory like the Go module cache or Xcode's DerivedData
// can genuinely take several seconds, so silence would read as "stuck".
type Progress struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
}

// Scan checks every known category for this OS and returns the ones that
// actually exist here, largest first. A category that isn't installed or
// is already empty is simply omitted.
func Scan(ctx context.Context, onProgress func(Progress)) ([]Category, error) {
	publish := func(p Progress) {
		if onProgress != nil {
			onProgress(p)
		}
	}

	defs := platformCategories()
	results := make([]Category, 0, len(defs))

	for i, def := range defs {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		publish(Progress{Stage: "scanning", Message: fmt.Sprintf("checking %s… (%d/%d)", def.Name, i+1, len(defs))})

		candidates := def.resolve()
		var existing []string
		var total int64
		for _, p := range candidates {
			info, err := os.Stat(p)
			if err != nil || !info.IsDir() {
				continue
			}
			existing = append(existing, p)
			total += dirSize(ctx, p)
		}
		if len(existing) == 0 || total == 0 {
			continue
		}
		results = append(results, Category{
			ID: def.ID, Name: def.Name, Description: def.Description,
			Paths: existing, Bytes: total,
		})
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Bytes > results[j].Bytes })
	return results, nil
}

func dirSize(ctx context.Context, root string) int64 {
	var total int64
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // unreadable entry (permissions, broken symlink) — skip it, don't abort the whole scan
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() || d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		if info, err := d.Info(); err == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}
