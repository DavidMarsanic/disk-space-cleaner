//go:build windows

package engine

import (
	"os"
	"path/filepath"
)

// atLocalAppData joins %LocalAppData% with parts — where Windows expects
// per-user cache and temp data to live (as opposed to %AppData%, for
// roaming data that shouldn't be treated as disposable).
func atLocalAppData(parts ...string) func() []string {
	return func() []string {
		base := os.Getenv("LocalAppData")
		if base == "" {
			return nil
		}
		return []string{filepath.Join(append([]string{base}, parts...)...)}
	}
}

// platformCategories lists Windows's well-documented, redownload-on-demand
// cache locations. A curated, individually-labeled list is more
// transparent about what's actually being offered up than one broad
// %LocalAppData%\Temp-and-everything-else umbrella would be.
func platformCategories() []categoryDef {
	return []categoryDef{
		{
			ID:          "windows-temp",
			Name:        "Windows temporary files",
			Description: "Short-lived files apps write to your user temp folder and don't always clean up after themselves. Safe to clear — this is the same folder Windows' own Disk Cleanup targets.",
			resolve:     atLocalAppData("Temp"),
		},
		{
			ID:          "npm-cache",
			Name:        "npm cache",
			Description: "Downloaded package archives npm keeps around. Redownloaded automatically if needed again.",
			resolve:     atLocalAppData("npm-cache"),
		},
		{
			ID:          "yarn-cache",
			Name:        "Yarn cache",
			Description: "Downloaded package archives Yarn keeps around. Redownloaded automatically if needed again.",
			resolve:     atLocalAppData("Yarn", "Cache"),
		},
		{
			ID:          "pnpm-store",
			Name:        "pnpm store",
			Description: "pnpm's shared package store. Projects you've already installed keep working — pnpm just redownloads what it needs the next time you install somewhere new.",
			resolve:     atLocalAppData("pnpm", "store"),
		},
		{
			ID:          "pip-cache",
			Name:        "pip cache",
			Description: "Downloaded Python package archives pip keeps around. Redownloaded automatically if needed again.",
			resolve:     atLocalAppData("pip", "Cache"),
		},
		{
			ID:          "go-build-cache",
			Name:        "Go build cache",
			Description: "Compiled-package build artifacts the Go toolchain keeps to speed up future builds. Rebuilt automatically as needed.",
			resolve:     atLocalAppData("go-build"),
		},
		{
			ID:          "go-mod-cache",
			Name:        "Go module cache",
			Description: "Downloaded Go module source code. Redownloaded automatically if needed again — this one can be large if you've worked on a lot of Go projects.",
			resolve:     at("go", "pkg", "mod"),
		},
		{
			ID:          "chrome-cache",
			Name:        "Chrome browser cache",
			Description: "Cached web page resources Chrome keeps for faster page loads. Rebuilt automatically as you browse.",
			resolve:     atLocalAppData("Google", "Chrome", "User Data", "Default", "Cache"),
		},
		{
			ID:          "edge-cache",
			Name:        "Edge browser cache",
			Description: "Cached web page resources Edge keeps for faster page loads. Rebuilt automatically as you browse.",
			resolve:     atLocalAppData("Microsoft", "Edge", "User Data", "Default", "Cache"),
		},
	}
}
