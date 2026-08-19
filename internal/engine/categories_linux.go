//go:build linux

package engine

// platformCategories lists Linux's well-documented, redownload-on-demand
// cache locations. Paths follow the XDG base directory convention (and
// each tool's own documented default) rather than a broad ~/.cache
// umbrella — a curated, individually-labeled list is more transparent
// about what's actually being offered up, and avoids double-counting
// bytes that would also show up under a specific category below.
func platformCategories() []categoryDef {
	return []categoryDef{
		{
			ID:          "npm-cache",
			Name:        "npm cache",
			Description: "Downloaded package archives npm keeps around. Redownloaded automatically if needed again.",
			resolve:     at(".npm"),
		},
		{
			ID:          "yarn-cache",
			Name:        "Yarn cache",
			Description: "Downloaded package archives Yarn keeps around. Redownloaded automatically if needed again.",
			resolve:     at(".cache", "yarn"),
		},
		{
			ID:          "pnpm-store",
			Name:        "pnpm store",
			Description: "pnpm's shared package store. Projects you've already installed keep working — pnpm just redownloads what it needs the next time you install somewhere new.",
			resolve:     at(".local", "share", "pnpm", "store"),
		},
		{
			ID:          "pip-cache",
			Name:        "pip cache",
			Description: "Downloaded Python package archives pip keeps around. Redownloaded automatically if needed again.",
			resolve:     at(".cache", "pip"),
		},
		{
			ID:          "go-build-cache",
			Name:        "Go build cache",
			Description: "Compiled-package build artifacts the Go toolchain keeps to speed up future builds. Rebuilt automatically as needed.",
			resolve:     at(".cache", "go-build"),
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
			resolve:     at(".cache", "google-chrome"),
		},
		{
			ID:          "firefox-cache",
			Name:        "Firefox browser cache",
			Description: "Cached web page resources Firefox keeps for faster page loads. Rebuilt automatically as you browse.",
			resolve:     glob(".cache", "mozilla", "firefox", "*", "cache2"),
		},
	}
}
