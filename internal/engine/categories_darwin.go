//go:build darwin

package engine

// platformCategories lists macOS's well-documented, redownload-on-demand
// cache locations. Every path here is either verified directly against
// the tool that owns it (`brew --cache`, `npm config get cache`,
// `pip cache dir`, `yarn cache dir`, `pnpm store path`, `go env GOCACHE
// GOMODCACHE`) or Apple's own documented Developer/Library layout — none
// of these are guessed. Deliberately excludes broad umbrellas like the
// whole of ~/Library/Caches: a curated, individually-labeled list is more
// transparent about what's actually being offered up, and avoids
// double-counting bytes that would also show up under a specific
// category below.
func platformCategories() []categoryDef {
	return []categoryDef{
		{
			ID:          "homebrew-cache",
			Name:        "Homebrew download cache",
			Description: "Downloaded package archives Homebrew keeps around after installing. Redownloaded automatically if needed again.",
			resolve:     at("Library", "Caches", "Homebrew"),
		},
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
			resolve:     at("Library", "Caches", "Yarn"),
		},
		{
			ID:          "pnpm-store",
			Name:        "pnpm store",
			Description: "pnpm's shared package store. Projects you've already installed keep working — pnpm just redownloads what it needs the next time you install somewhere new.",
			resolve:     at("Library", "pnpm", "store"),
		},
		{
			ID:          "pip-cache",
			Name:        "pip cache",
			Description: "Downloaded Python package archives pip keeps around. Redownloaded automatically if needed again.",
			resolve:     at("Library", "Caches", "pip"),
		},
		{
			ID:          "go-build-cache",
			Name:        "Go build cache",
			Description: "Compiled-package build artifacts the Go toolchain keeps to speed up future builds. Rebuilt automatically as needed.",
			resolve:     at("Library", "Caches", "go-build"),
		},
		{
			ID:          "go-mod-cache",
			Name:        "Go module cache",
			Description: "Downloaded Go module source code. Redownloaded automatically if needed again — this one can be large if you've worked on a lot of Go projects.",
			resolve:     at("go", "pkg", "mod"),
		},
		{
			ID:          "xcode-derived-data",
			Name:        "Xcode DerivedData",
			Description: "Xcode's build products and indexes. Rebuilt automatically the next time you open or build a project — a very common thing to clear when Xcode's disk usage gets out of hand.",
			resolve:     at("Library", "Developer", "Xcode", "DerivedData"),
		},
		{
			ID:          "xcode-device-support",
			Name:        "Xcode iOS Device Support",
			Description: "Debug symbols Xcode downloads per iOS version for on-device debugging. Redownloaded automatically the next time you connect a device running that iOS version.",
			resolve:     at("Library", "Developer", "Xcode", "iOS DeviceSupport"),
		},
		{
			ID:          "simulator-caches",
			Name:        "iOS Simulator caches",
			Description: "Cache data for the iOS Simulator — not the simulator devices themselves, just their cache. Rebuilt automatically.",
			resolve:     at("Library", "Developer", "CoreSimulator", "Caches"),
		},
		{
			ID:          "chrome-cache",
			Name:        "Chrome browser cache",
			Description: "Cached web page resources Chrome keeps for faster page loads. Rebuilt automatically as you browse.",
			resolve:     at("Library", "Caches", "Google", "Chrome"),
		},
		{
			ID:          "safari-cache",
			Name:        "Safari browser cache",
			Description: "Cached web page resources Safari keeps for faster page loads. Rebuilt automatically as you browse.",
			resolve:     at("Library", "Caches", "com.apple.Safari"),
		},
	}
}
