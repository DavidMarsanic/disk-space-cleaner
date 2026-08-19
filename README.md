# Disk Space Cleaner

Finds well-known, documented-safe cache directories — package managers,
build tools, browsers — and lets you send them to the Trash. Opens as
its own window.

## What it checks

A curated, individually-labeled list, not a broad sweep of your whole
system: Homebrew/npm/Yarn/pnpm/pip download caches, Go's build and
module caches, Xcode's DerivedData/iOS Device Support/Simulator caches,
and browser caches (Chrome, Safari, Edge, Firefox depending on OS).
Every location is one your own package manager or app already knows how
to rebuild — nothing here is a one-of-a-kind file. A category that
isn't installed or is already empty just doesn't show up.

Nothing outside these specific, hardcoded locations under your own home
directory is ever touched, and nothing here needs admin/root access.

## Trash, not delete

Selected caches go to your OS's real Trash/Recycle Bin — recoverable,
not gone — exactly like dragging them there yourself. The only
permanent action in this whole app is the separate "Empty Trash"
button, which empties the Trash itself; it asks for its own explicit
confirmation before doing anything, since that one step can't be
undone.

## Requirements

**A Chromium-based browser already installed**: Google Chrome, Chromium,
Brave, Microsoft Edge, or Arc — renders the app's own UI window.

## Notes

- Windows can't report the Recycle Bin's size (it's a hidden, per-volume,
  per-user structure — not a simple folder to list the way macOS/Linux
  Trash is), so that number shows as "unknown" there. Emptying it still
  works, via PowerShell's built-in `Clear-RecycleBin`.
- This isn't a general disk-usage browser (à la DaisyDisk/WinDirStat) —
  it only ever looks at the specific cache locations listed above, on
  purpose, so every action stays predictable and reversible.

## License

MIT — see [LICENSE](LICENSE).
