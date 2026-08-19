package engine

import (
	"os"
	"path/filepath"
)

func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

// at joins homeDir with parts into a single-path resolver. Every
// category's location this app knows about lives under the user's own
// home directory — never a system-wide path that would need elevated
// permissions to touch. Returns a resolver yielding no paths (rather
// than a broken relative one) if the home directory can't be resolved.
func at(parts ...string) func() []string {
	return func() []string {
		h := homeDir()
		if h == "" {
			return nil
		}
		return []string{filepath.Join(append([]string{h}, parts...)...)}
	}
}

// glob is like at, but the final path segment is a glob pattern —
// needed for locations with a randomized component, like a Firefox
// profile folder name.
func glob(parts ...string) func() []string {
	return func() []string {
		h := homeDir()
		if h == "" {
			return nil
		}
		pattern := filepath.Join(append([]string{h}, parts...)...)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil
		}
		return matches
	}
}
