
## PROJECT OVERVIEW
Build a production-grade, cross-platform package manager CLI called GoPack — written entirely in Go (Golang) — that rivals npm, pnpm, and Bun in speed and developer experience. It must install, search, and manage both JavaScript/TypeScript packages and full-stack frameworks (Next.js, NestJS, React, Vue, Svelte, Astro, Remix, SvelteKit, Nuxt, Angular, and more) from a single unified CLI.

## CORE CLI INTERFACE
Design an interactive TUI (Terminal UI) using the Bubble Tea framework with the following keyboard-driven navigation:

- ↑ / ↓ arrow keys → scroll through search results
- ← / → arrow keys → switch between two search tabs: [framework] and [packages]
- Enter → select and install the highlighted package or framework
- Live search as you type — results update instantly with fuzzy matching
- "more — scroll down" hint visible when the list is truncated
- Show package name, version, weekly downloads, and description inline in results
- Each result row has a compact security badge on the right side (see Security section)

## COMMANDS
Implement the following CLI commands:

  gp search <query>          → open interactive TUI search
  gp install <name>          → install latest stable version
  gp install <name@version>   → install specific version
  gp remove <name>           → uninstall package
  gp update                   → update all packages, show changelog diff
  gp list                     → list installed packages with security status
  gp audit                    → run full vulnerability scan on all deps
  gp create <framework>      → scaffold a new project (like create-next-app)
  gp doctor                   → check environment health (Node, Go, runtime versions)
  gp cache clean              → clear local package cache

## FRAMEWORK INSTALLER
Support one-command scaffolding for the following (always fetch latest stable release):

  Next.js, NestJS, React, Vue 3, Svelte, SvelteKit, Astro, Remix, Nuxt 3,
  Angular, Solid.js, Qwik, Expo (React Native), Electron, and Tauri.

Each scaffold should:
- Auto-detect the user's package manager preference (npm / pnpm / bun / yarn)
- Prompt for TypeScript vs JavaScript, CSS framework choice, and Git init
- Run a post-install security audit automatically before handing off

## SECURITY SYSTEM
This is the most visually distinctive feature. Every package in search results, install output, and `gp list` must display a real-time security badge rendered as a colored terminal box on the right side of the row:

  ■ GREEN box   → No known vulnerabilities. Last audited timestamp shown.
  ■ YELLOW box  → Low/moderate advisory exists. CVE ID shown on hover/expand.
  ■ RED box     → High or critical CVE. Blocked install by default with --force flag to override.

Security data sources to query in parallel:
  1. OSV (osv.dev API) — Google's open-source vulnerability database
  2. npm audit registry — for JS/TS packages
  3. Snyk Advisor API — for health score and maintenance status
  4. GitHub Advisory Database — via GraphQL API

Cache audit results locally for 1 hour to avoid redundant API calls.

## PERFORMANCE REQUIREMENTS
- Written in pure Go — single compiled binary, zero runtime dependency
- Parallel downloads using goroutines — install multiple packages simultaneously
- Content-addressable local cache (similar to pnpm store) to avoid re-downloading
- Cold-start search response under 100ms
- Install time for a typical package under 2 seconds on a standard connection
- Progress bar with bytes/sec and ETA during downloads (use mpb or Bubble Tea progress)

## TUI DESIGN SPEC
Use Bubble Tea + Lip Gloss for all terminal rendering:

- Dark background panel with a thin rounded border (use box-drawing characters)
- Two tab pills at the top: [framework] [packages] — active tab underlined
- Search input at the top with a blinking cursor
- Results list below with alternating row shading
- Right-aligned security badge column using ANSI color blocks
- Bottom status bar: keybindings legend (↑↓ navigate · ←→ switch tab · Enter install · q quit)
- Smooth scroll animation when navigating results

## ERROR HANDLING & UX
- Never silently fail. All errors must show a descriptive message with a suggested fix.
- Network timeout: retry 3 times with exponential backoff, then show offline fallback from cache.
- If a RED-rated package is requested: show full CVE details, affected versions, and the patched version before prompting.
- Log all operations to ~/.gopack/logs/ with timestamps.

## CONFIG & LOCKFILE
- gopack.json — project manifest (mirrors package.json structure for compatibility)
- gopack.lock — deterministic lockfile with content hashes (SHA-256)
- Support reading existing package.json for drop-in migration from npm/pnpm/bun

## TECH STACK
  Language:         Go 1.22+
  TUI framework:    Bubble Tea + Lip Gloss + Bubbles
  HTTP client:      Go net/http with connection pooling
  JSON parsing:     encoding/json + gjson for fast path queries
  Config:           Viper
  CLI flags:        Cobra
  Progress bars:    mpb or Bubble Tea progress component
  Testing:          Go testing + testify
  Build/release:    GoReleaser (cross-compile for macOS, Linux, Windows)

## DELIVERABLES
Produce a fully working Go codebase with:
  1. All commands implemented and tested
  2. TUI matching the described spec with keyboard nav
  3. Security badge rendering working in all three states
  4. README with install instructions, command reference, and a GIF demo placeholder
  5. GitHub Actions CI/CD pipeline using GoReleaser
  6. Homebrew formula and install script for one-line install
