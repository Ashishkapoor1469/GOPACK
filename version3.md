
## PROJECT — GOPACK
Build a production-grade, cross-platform package manager CLI called GoPack, written entirely in Go (Golang). It rivals npm, pnpm, and Bun in speed and developer experience.

Core capabilities:
  - Install, search, and manage JS/TS packages AND full-stack frameworks from one unified CLI
  - Works completely offline once packages are cached — no internet required
  - Zero AI features, zero API key required — 100% deterministic, rule-based tooling
  - Ships as a single compiled binary, zero runtime dependency

## ENTRY POINT — HOW IT LAUNCHES
When the user types gp in any terminal with no arguments, the program immediately opens the full interactive TUI window. There is NO splash screen, NO loading delay, NO menu — just the TUI, ready to use.

The title bar of the terminal window shows: gp — gopack v2.x.x
The first line inside the TUI shows:
  gopack v2.x.x  ·  registry: registry.npmjs.org  ·  store: ~/.gopack/store  ·  offline-ready

The TUI has TWO modes that coexist:
  1. SEARCH MODE  — the default. Two side-by-side search panels: [FRAMEWORK] and [PACKAGES].
  2. DIRECT MODE  — activated by typing a package name directly OR pressing / to jump to search.

Both modes are always one keypress away from each other. The user never needs to exit and re-run a command.

## FRAMEWORK SCAFFOLDING
Support one-command scaffolding (always fetches latest stable):
  Next.js, NestJS, React, Vue 3, Svelte, SvelteKit, Astro, Remix, Nuxt 3,
  Angular, Solid.js, Qwik, Expo (React Native), Electron, Tauri

Each scaffold:
  - Auto-detects user's preferred package manager (npm / pnpm / bun / yarn)
  - Prompts for TypeScript vs JavaScript, CSS framework, Git init
  - Runs a post-install security audit automatically before handing off



## TUI LAYOUT — EXACT SPECIFICATION
Use Bubble Tea + Lip Gloss + Bubbles for all terminal rendering.
The entire TUI is enclosed in a rounded border using Unicode box-drawing characters.

─── TOP BAR ───────────────────────────────────────────
  Single line: version · registry URL · store path · offline status
  Color: dim/muted. Renders immediately on launch.

─── SEARCH ROW ────────────────────────────────────────
  Two side-by-side input panels separated by a vertical divider:

  ┌─────────────────────┐  │  ┌─────────────────────┐
  │ FRAMEWORK           │  │  │ PACKAGES             │
  │ > next█             │  │  │ search packages...   │
  └─────────────────────┘  │  └─────────────────────┘

  - Active tab: border color = purple (#7F77DD), label bright
  - Inactive tab: border dim, label muted, placeholder text visible
  - ← → arrow keys switch active tab instantly
  - Typing filters results in real-time with fuzzy matching
  - Blinking block cursor in the active input field

─── RESULTS LIST ──────────────────────────────────────
  Header row: "RESULTS · N found" on left · "Space=select · multi-install ready" on right

  Each result row contains these columns left-to-right:
    [selector]  [name]        [version]  [description — truncated]  [weekly DLs]  [security badge]

  Selector states:
    ○  = radio dot (cursor is here, not yet selected for queue)
    ☑  = green checkbox (added to install queue via Space)

  Active/cursor row: purple-tinted background (#1a1730), name in lavender
  Hover row: very subtle darker bg
  Security badge (right-aligned colored pill):
    ● safe  — dark green bg, bright green text, green border
    ● low   — dark amber bg, amber text, amber border
    ● CVE   — dark red bg, red text, red border

  "↓  N more — scroll down" hint line at the bottom of results
  ↑ ↓ keys navigate the list. Smooth scroll, no flicker.

─── INSTALL QUEUE ─────────────────────────────────────
  Appears ONLY when ≥1 package is Space-selected.
  Header: green "INSTALL QUEUE · N selected" on left · "Enter to install all" on right
  Body: row of green pill badges, one per queued package, each with a ✕ to dequeue
  Example: [next-auth ✕]  [next-intl ✕]  [tailwindcss ✕]

─── STATUS BAR ────────────────────────────────────────
  Always visible at the very bottom. Fixed one-line keybinding legend:
  [↑↓] navigate  [←→] switch tab  [Space] select  [Enter] install  [/] search  [q] quit
  Right side: pill showing current mode — "multi-select ON" or "direct mode"



## THE / COMMAND — CORE INNOVATION
This is the most important UX feature. Implement it exactly as described.

When the user is OUTSIDE the TUI (normal terminal prompt):
  Typing gp / immediately opens the TUI and focuses the active search panel.
  Cursor is placed in the search input, ready to type.
  This is a single fluid motion — no intermediate step.

When the user is INSIDE the TUI:
  Pressing / at any time clears the current search input and refocuses it.
  The cursor jumps to the search box of the currently active tab (FRAMEWORK or PACKAGES).
  If already in the search box, pressing / selects all text so the user can retype immediately.

## DIRECT INSTALL — TYPE TO INSTALL
Zero friction install. No sub-command needed.

When the user types gp <packagename> with no sub-command:
  GoPack checks if the argument matches a known command (install, search, audit, etc.).
  If it does NOT match any command → treat it as a direct install.
  Example: gp react → installs react@latest immediately, no TUI opened.
  Example: gp react tailwind framer-motion → installs all three in parallel.

Output during direct install:
  ┌─ installing 3 packages ──────────────────────────────┐
  │  ● react@18.3.1          ████████████████████  done  │
  │  ● tailwindcss@3.4.4     ███████████░░░░░░░░░   54%  │
  │  ● framer-motion@11.2.0  ░░░░░░░░░░░░░░░░░░░░    0%  │
  │  downloading 3 files · 2.4 MB/s · ETA 3s            │
  └──────────────────────────────────────────────────────┘

  After install: show security badge for each package inline.
  If any package has a RED badge: show full CVE details before completing.

## MULTI-PACKAGE INSTALL FROM TUI
Inside the TUI results list:
  - Press Space on any result row to toggle it into the install queue
  - Queue pill appears at the bottom showing all selected packages
  - Press Enter to install all queued packages simultaneously using goroutines
  - Progress bars render for each package in parallel (same format as direct install above)
  - Press Escape to clear the queue without installing

## ENTER ON SINGLE RESULT
If no packages are queued (queue is empty) and user presses Enter on a highlighted row:
  - Installs just that one package
  - No confirmation prompt needed — install starts immediately
  - Progress bar renders inline, then returns focus to the TUI on completion



## FULL COMMAND REFERENCE
Launch commands:
  gp                          → open interactive TUI (default, no args)
  gp /                        → open TUI with search focused
  gp <name>                   → direct install (no sub-command needed)
  gp <a> <b> <c>             → install multiple packages in parallel

Install / remove:
  gp install <name>           → install latest stable version
  gp install <name@ver>       → install pinned version
  gp install --env production  → skip devDependencies
  gp install --offline        → resolve from local cache only
  gp install --frozen         → refuse to change lockfile (CI mode)
  gp remove <name>            → uninstall and clean lockfile
  gp pin <name>               → lock to current version forever
  gp unpin <name>             → remove pin

Search + info:
  gp search <query>           → open TUI with query pre-filled
  gp info <name>              → show package details, readme excerpt, badge
  gp why <name>               → which packages depend on it
  gp outdated                 → list all deps behind latest with age
  gp list                     → installed deps with security status

Update:
  gp update [name]            → update one or all, show changelog diff
  gp update --interactive     → TUI checklist — pick what to bump
  gp changelog <n> <v1> <v2> → diff between two versions

Security + health:
  gp audit                    → full CVE scan, works offline
  gp audit --sync-db          → download OSV DB snapshot to local SQLite
  gp audit --json             → machine-readable output for CI
  gp health                   → full project dashboard (score, freshness, licenses)
  gp verify                   → re-hash all packages against lockfile

Workspace / monorepo:
  gp ws list                  → list workspace packages
  gp ws run <script>          → run script across all workspaces
  gp ws add <pkg> --filter X  → add dep to a specific workspace

Scripts:
  gp run <script>             → run script from gopack.json
  gp run --list               → show all scripts with descriptions
  gp run build test --parallel → run multiple scripts simultaneously

Tooling:
  gp create <framework>       → scaffold new project
  gp global install <name>    → install CLI tool globally
  gp doctor                   → check env health
  gp dedupe                   → collapse duplicate dep versions
  gp graph [--depth N]        → ASCII dependency tree
  gp graph --dot              → export as Graphviz DOT format
  gp patch <name>             → edit package source directly
  gp patch commit <name>      → save as .patch, auto-applied on install
  gp licenses [--check X]     → SPDX license report
  gp export <format>          → dep tree as JSON / CSV / DOT
  gp migrate --from npm       → import package-lock.json
  gp cache clean              → wipe cache
  gp cache stats              → cache size, hit rate, oldest entry
  gp config get/set/list      → manage ~/.gopack/config.toml
  gp plugin install/list/remove → manage plugins
  gp shell install [shell]    → install tab-completion + cd hook
  gp watch [--daemon]         → background CVE watcher
  gp watch --webhook <url>    → POST CVE alerts to Slack/Discord
  gp ci                       → install + audit + verify lockfile (CI one-liner)
  gp env diff <a> <b>         → diff two install profiles
  gp prefetch gopack.json     → download all deps before going offline
  gp export-bundle <name>     → pack pkg + all deps into .gpbundle
  gp import-bundle <file>     → import bundle into local store

Short aliases (built-in):
  gp i = install  gp rm = remove  gp up = update  gp s = search



## DEPENDENCY GRAPH VIEWER
Command: gp graph [--depth N]
  - Full ASCII/Unicode dependency tree in terminal
  - Direct deps in white, transitive in dim
  - Circular deps detected and shown with a red ↺ marker
  - Duplicate version conflicts flagged inline with amber ⚠
  - gp graph --dot > deps.dot exports Graphviz DOT for external rendering

## INTERACTIVE UPDATE TUI
Command: gp update --interactive
  - Full TUI checklist: each outdated package on its own row
  - Columns: name · installed · latest · age · changelog summary · security badge
  - Space to toggle, Enter to update selected, q to cancel
  - Major bumps (breaking changes) shown in red, require --allow-breaking override
  - "Update all safe" one-key shortcut: selects only patch + minor updates

## PROJECT HEALTH DASHBOARD
Command: gp health
  - Total deps: direct / transitive / dev counts
  - Security score: X green · Y yellow · Z red
  - License compliance: pass / warn / fail
  - Freshness: % of deps within 1 major of latest
  - Bundle size estimate per dep
  - Last audit timestamp
  - Actionable suggestions: gp dedupe, gp outdated, gp audit

## WATCH DAEMON + NOTIFICATIONS
Command: gp watch [--daemon]
  - Polls registry every 24h for new versions of installed packages
  - Sends OS desktop notification when a critical CVE is published
  - gp watch --webhook <url> → POST JSON to Slack/Discord on CVE discovery
  - Runs as lightweight background process

## PATCH MANAGEMENT
Command: gp patch <name>
  - Copies package source into local patches/ folder
  - Opens in $EDITOR for direct modification
  - gp patch commit generates .patch file registered in gopack.json
  - Patches auto-applied on every install — version-pinned
  - Warns when upstream releases a fix for the patched issue

## ENVIRONMENT PROFILES
  - Named install profiles: dev · staging · production · test
  - gp install --env production skips devDependencies
  - gp env diff dev production shows what differs between profiles
  - Per-profile .env loading with variable substitution in scripts

## LICENSE COMPLIANCE ENGINE
Command: gp licenses [--check <policy>]
  - Extracts SPDX license identifiers from all installed deps
  - --check strict fails CI if any GPL/AGPL dep found in commercial project
  - Define allowed/blocked licenses in gopack.json under "licensePolicy"
  - Output as table, JSON, or CSV
  - Warns on dual-licensed packages requiring attribution

## CI/CD INTEGRATION
  - gp ci: installs deps, audits, verifies lockfile — designed for CI pipelines
  - Machine-readable: gp audit --json / gp health --json
  - Exit codes: 0 = clean · 1 = warning · 2 = critical CVE or lockfile mismatch
  - JUnit XML: gp audit --junit > report.xml
  - GitHub Actions: writes markdown summary to $GITHUB_STEP_SUMMARY automatically

## PLUGIN SYSTEM
  - gp plugin install <name> — install a GoPack plugin (Go binary, defined interface)
  - Plugins can add new commands, registry sources, or security scanners
  - Sandboxed: plugins cannot write outside their own data directory
  - Plugin interface spec document: PLUGINS.md



## FULL OFFLINE MODE
No internet? No problem. GoPack works 100% offline once packages are cached.

Content-addressable store at ~/.gopack/store/ — packages stored by SHA-256 hash.

  gp install --offline         → resolve entirely from local cache, fail fast if missing
  gp prefetch gopack.json      → download all declared deps (before a flight, etc.)
  gp export-bundle <name>      → pack package + all transitive deps → .gpbundle tarball
  gp import-bundle <file>      → import bundle into store without network
  gp audit --sync-db           → download full OSV vulnerability DB as local SQLite
                                   all future audits run against this snapshot — no internet needed

## PERFORMANCE TARGETS
  - Cache-hit install: under 50ms
  - Cold install (network): under 2 seconds for a typical package
  - TUI search response: under 100ms from keystroke to rendered results
  - Binary cold start: under 20ms

Implementation requirements:
  - Parallel downloads via goroutines — all installs simultaneous, not sequential
  - HTTP/2 connection pooling for registry requests
  - Streaming extraction: decompress tarballs as they arrive, don't buffer full download
  - Delta updates: for patch-version bumps, fetch only the binary diff (xdelta3 format)

## PROGRESS BAR RENDERING
During any install (direct, TUI queue, or gp install), show a bordered progress block:

  ┌─ installing N packages ──────────────────────────────┐
  │  ● pkg-a@x.x.x     ████████████████████  done        │
  │  ● pkg-b@x.x.x     ███████████░░░░░░░░░   54%        │
  │  ● pkg-c@x.x.x     ░░░░░░░░░░░░░░░░░░░░    0%        │
  │  downloading N files · X.X MB/s · ETA Xs             │
  └──────────────────────────────────────────────────────┘

After install completes: each package gets its security badge printed inline.
RED badge: block and show full CVE details + patched version before asking to confirm.

## REPRODUCIBLE BUILDS
  - gopack.lock: deterministic lockfile with SHA-256 hashes for every package
  - gp install --frozen: refuse to update lockfile; fail if it would change
  - gp verify: re-hash all installed packages; alert on any tampering
  - gp migrate --from npm: import package-lock.json into gopack.lock automatically

## GLOBAL TOOL MANAGER
  - gp global install <name>: install CLI tools in isolated per-tool environment
  - gp global list / update --all
  - Tools never conflict with project deps
  - Auto-adds ~/.gopack/bin/ to PATH via shell integration script

## SHELL INTEGRATION
Command: gp shell install [bash|zsh|fish|powershell]
  - Shell hook: auto-detects gopack.json on cd, activates correct Node version
  - Full tab completion for all commands, subcommands, and package names
  - gp shell uninstall: cleanly removes all hooks



## SECURITY BADGE SYSTEM
Every package — in search results, install output, and gp list — shows a colored security badge.

Badge states (rendered as a right-aligned colored pill in the TUI):
  ● safe   — No known CVEs. Shows last-audited timestamp on row expand.
  ● low    — Low or moderate advisory exists. CVE ID shown on expand.
  ● CVE    — High or critical CVE. Install is BLOCKED by default.
               Show full CVE details + affected versions + patched version.
               User must pass --force flag to override the block.

Security data sources (queried in parallel, results merged):
  1. OSV API (osv.dev)       — Google's open vulnerability database
  2. npm audit registry      — for JS/TS packages
  3. GitHub Advisory DB      — queried via GraphQL API
  4. Snyk Advisor API        — health score + maintenance status

Caching: audit results cached locally for 1 hour to avoid redundant API calls.
Offline: if network is unavailable, read from local SQLite OSV snapshot (gp audit --sync-db).

## AUDIT OUTPUT FORMAT
  gp audit standard output (terminal):
    ┌─ security audit ────────────────────────────────────────┐
    │  ✓  142 packages  ● safe                                │
    │  ⚠   3 packages   ● low advisory    CVE-2024-XXXXX      │
    │  ✗   1 package    ● critical CVE    CVE-2024-YYYYY      │
    │  last synced: 2 minutes ago                             │
    └─────────────────────────────────────────────────────────┘

  Machine-readable: gp audit --json outputs structured JSON
  JUnit XML: gp audit --junit > report.xml for CI artifact ingestion
  Exit codes: 0 = all clean · 1 = low/moderate advisory · 2 = high/critical CVE

## ERROR HANDLING
  - Never silently fail. Every error shows a descriptive message + suggested fix.
  - Network timeout: retry 3× with exponential backoff, then show offline fallback.
  - RED-rated package requested: show full CVE details, affected versions, patched version.
  - Log all operations to ~/.gopack/logs/ with timestamps.
  - Unknown command that looks like a package name → suggest: "did you mean: gp install X?"



## TECH STACK
  Language:           Go 1.22+
  TUI framework:      Bubble Tea + Lip Gloss + Bubbles
  HTTP client:        Go net/http with HTTP/2 + connection pooling
  JSON parsing:       encoding/json + gjson (fast path queries)
  Config:             Viper (TOML + env var override)
  CLI flags:          Cobra
  Progress bars:      mpb or Bubble Tea progress component
  Fuzzy search:       go-fuzzyfinder or custom Levenshtein impl
  Offline audit DB:   SQLite via mattn/go-sqlite3
  Delta updates:      xdelta3 bindings or pure-Go implementation
  Desktop notify:     go-toast (Windows) · osascript (macOS) · libnotify (Linux)
  Shell completion:   Cobra completions (bash/zsh/fish/powershell)
  Archive handling:   archive/tar + compress/gzip (stdlib)
  Testing:            Go testing + testify + golden file tests
  Build/release:      GoReleaser (cross-compile macOS · Linux · Windows · ARM)

## gopack.json SCHEMA
  {
    "name": "my-app",
    "version": "1.0.0",
    "scripts": {
      "build": "tsc",
      "test": "jest",
      "dev": "next dev"
    },
    "dependencies": {},
    "devDependencies": {},
    "peerDependencies": {},
    "workspaces": ["packages/*"],
    "licensePolicy": {
      "allow": ["MIT", "Apache-2.0", "ISC"],
      "deny":  ["GPL-3.0", "AGPL-3.0"]
    },
    "patches": {
      "lodash@4.17.21": "patches/lodash+4.17.21.patch"
    },
    "gopack": {
      "pinnedVersions": ["react"],
      "hoistPatterns": ["*"],
      "defaultEnv": "dev"
    }
  }

## ~/.gopack/config.toml SCHEMA
  defaultEnv      = "dev"
  parallelism     = 8
  cacheDir        = "~/.gopack/store"
  auditDbPath     = "~/.gopack/osv.db"
  auditCacheTTL   = "1h"
  notifyOnCVE     = true
  webhookURL      = ""
  licensePolicy   = "permissive"     # permissive | strict | custom
  [aliases]
    dev = "install --env dev"
    fresh = "cache clean && install"

## DELIVERABLES CHECKLIST
  1. Full Go codebase — all commands implemented and unit tested
  2. TUI exactly matching the layout spec — keyboard nav, multi-select, queue, badges
  3. / command and direct install (gp <name>) working as described
  4. Offline mode — bundle export/import, local OSV SQLite DB
  5. Shell integration script for bash/zsh/fish/powershell
  6. README with install instructions, full command reference, ASCII demo recording
  7. GitHub Actions CI/CD pipeline using GoReleaser
  8. Homebrew formula + one-line curl install script
  9. Plugin interface spec (PLUGINS.md)
  10. Man page auto-generated via cobra-man
