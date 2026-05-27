
## PROJECT OVERVIEW
Build a production-grade, cross-platform package manager CLI called GoPack — written entirely in Go (Golang) — that rivals npm, pnpm, and Bun in speed and developer experience. Install, search, and manage both JS/TS packages and full-stack frameworks from a single unified CLI. No AI features, no API key required — 100% deterministic, offline-capable tooling.

## CLI COMMANDS
  gp search <query>              → interactive TUI search (fuzzy, live)
  gp install <name>              → install latest stable version
  gp install <name@version>       → install pinned version
  gp remove <name>               → uninstall + clean lockfile
  gp update [name]                → update one or all deps, show changelog diff
  gp update --interactive          → pick which deps to update via TUI checklist
  gp list                         → show installed deps with security status
  gp audit                        → full vulnerability scan, offline-safe
  gp create <framework>           → scaffold new project
  gp why <name>                   → show which package(s) depend on it
  gp dedupe                        → find and collapse duplicate dep versions
  gp outdated                     → list all packages behind latest with age
  gp doctor                        → check env health (Node, runtime, PATH)
  gp pin <name>                   → lock a package to its current version forever
  gp unpin <name>                 → remove pin
  gp cache clean                  → wipe local content-addressable cache
  gp cache stats                  → show cache size, hit rate, oldest entry
  gp licenses                     → print license of every installed dep (SPDX)
  gp export <format>              → export dep tree as JSON / CSV / DOT graph
  gp config get/set/list          → manage ~/.gopack/config.toml

## TUI DESIGN SPEC
Use Bubble Tea + Lip Gloss for all terminal rendering:
- Dark background panel, thin rounded border (box-drawing characters)
- Two tab pills: [framework] [packages] — ← → to switch, active tab underlined
- Live search input with blinking cursor, fuzzy matching as you type
- Results list: name · version · weekly downloads · description · security badge (right-aligned)
- "more — scroll down" hint when list is truncated
- Bottom status bar: ↑↓ navigate · ←→ switch tab · Enter install · / filter · q quit
- Smooth scroll animation on navigation
- Multi-select mode: press Space to tag multiple packages, Enter installs all at once

## SECURITY BADGES
Colored terminal box on the right of every result row and in `gp list`:
  ■ GREEN   → no known CVEs · last audited timestamp shown
  ■ YELLOW  → low/moderate advisory · CVE ID shown on expand
  ■ RED     → high/critical CVE · blocked by default, --force to override
Sources (queried in parallel): OSV API · npm audit · GitHub Advisory DB · Snyk Advisor
Cache audit results locally for 1 hour.


## DEPENDENCY GRAPH VIEWER  ★ NEW
Command: gp graph [--depth N]
Render a full ASCII/Unicode dependency tree in the terminal showing:
- Direct vs transitive deps (color-coded)
- Circular dependency detection with a clear warning
- Duplicate version conflicts flagged inline
- Optionally export to DOT format for Graphviz: gp graph --dot > deps.dot

## SCRIPT RUNNER  ★ NEW
Command: gp run <script>
- Reads scripts block from gopack.json (mirrors package.json)
- gp run --list shows all available scripts with descriptions
- Supports script hooks: pre<name> and post<name> run automatically
- Parallel script execution: gp run build test --parallel
- Live output streaming with timestamps per line

## WORKSPACE / MONOREPO SUPPORT  ★ NEW
Command: gp ws <subcommand>
- Detect workspaces declared in gopack.json (similar to pnpm workspaces)
- gp ws list → list all workspace packages
- gp ws run <script> → run a script across all workspace packages
- gp ws add <pkg> --filter <name> → add dep to a specific workspace
- Shared dep hoisting: common packages installed once, symlinked into each workspace
- Topological build order respecting inter-package dependencies

## PATCH MANAGEMENT  ★ NEW
Command: gp patch <name>
- Copies the package source into a local patches/ folder
- Opens it in $EDITOR for direct edits
- gp patch commit <name> generates a .patch file and registers it in gopack.json
- Patches auto-applied on every install (like pnpm patch)
- Patches are version-pinned; warn when upstream releases a fix

## ENVIRONMENT PROFILES  ★ NEW
Command: gp env <profile>
- Define named install profiles: dev · staging · production · test
- gp install --env production skips devDependencies automatically
- gp env diff dev production shows what packages differ between profiles
- Per-profile .env file loading with automatic variable substitution in scripts

## LICENSE COMPLIANCE ENGINE  ★ NEW
Command: gp licenses [--check <policy>]
- Extract SPDX license identifiers from every installed dep (direct + transitive)
- gp licenses --check strict fails CI if any GPL/AGPL dep is found in a commercial project
- Define allowed/blocked license lists in gopack.json under "licensePolicy"
- Output as table, JSON, or CSV for legal review
- Warn on dual-licensed packages that require attribution

## CHANGELOG VIEWER  ★ NEW
Command: gp changelog <name> [from@version] [to@version]
- Fetches CHANGELOG.md or GitHub Releases for any package
- Renders a clean diff of what changed between two versions
- Run automatically before gp update — show breaking changes in red
- Detect semver major bumps and require explicit --allow-breaking flag


## FULL OFFLINE MODE  ★ NEW
No internet? No problem. GoPack works completely offline once packages are cached.

- Content-addressable local store at ~/.gopack/store/ — packages stored by SHA-256 hash
- gp install --offline resolves entirely from local cache, fails fast if missing
- gp prefetch gopack.json — download all deps declared in a manifest before going offline (e.g. before a flight or conference)
- gp export-bundle <name> — pack a package + all its deps into a single .gpbundle tarball for air-gapped or offline transfer
- gp import-bundle <file.gpbundle> — import a bundle into local store without network
- Audit DB snapshot: download the full OSV vulnerability DB as a local SQLite file via gp audit --sync-db; all future audits run against this local snapshot — no internet needed

## PERFORMANCE TARGETS
- Single compiled binary — zero runtime dependency, instant cold start
- Parallel downloads via goroutines — batch installs simultaneously
- Cache hit install under 50ms · cold install under 2s · search response under 100ms
- HTTP/2 connection pooling for registry requests
- Streaming extraction: decompress tarballs as they download, don't wait for full download
- Delta updates: for patch-version bumps, fetch only the diff (xdelta3 format)

## GLOBAL TOOL MANAGER  ★ NEW
Command: gp global install <name>
- Install CLI tools globally (e.g. typescript, eslint, prettier, tsx)
- gp global list → show all globally installed tools
- gp global update --all → update everything globally
- Isolated per-tool environments: tools never conflict with project deps
- Auto-add ~/.gopack/bin/ to PATH via shell integration script

## SHELL INTEGRATION  ★ NEW
Command: gp shell install [bash|zsh|fish|powershell]
- Injects a thin shell hook that auto-detects gopack.json when cd-ing into a directory
- Activates the correct Node version automatically (reads .nvmrc / .node-version)
- Tab completion for all commands, subcommands, and package names
- gp shell uninstall cleanly removes all hooks

## REPRODUCIBLE BUILDS
- gopack.lock — deterministic lockfile with SHA-256 content hashes for every package
- gp install --frozen — refuse to update lockfile; fail if it would change (CI-safe)
- gp verify — re-hash all installed packages against lockfile; alert on tampering
- Support reading existing package.json for drop-in migration from npm/pnpm/bun
- gp migrate --from npm → import package-lock.json into gopack.lock automatically


## INTERACTIVE UPDATE TUI  ★ NEW
Command: gp update --interactive
- Full TUI checklist: each outdated package on its own row
- Columns: name · installed · latest · age · changelog summary · security status
- Space to toggle selection, Enter to update selected, q to cancel
- Breaking changes (major bumps) shown in red with a confirmation prompt
- "Update all safe" button: one key to select only patch + minor updates

## PROJECT HEALTH DASHBOARD  ★ NEW
Command: gp health
Render a full terminal dashboard showing:
- Total deps: direct / transitive / devDependencies counts
- Security score: X green · Y yellow · Z red packages
- License compliance: pass / warn / fail
- Freshness score: % of deps within 1 major version of latest
- Bundle size estimate per dep (using package.json "size" field + registry data)
- Last audit timestamp
- Suggest gp dedupe, gp outdated, or gp audit based on health signals

## NOTIFICATIONS + WATCH MODE  ★ NEW
Command: gp watch
- Run in background: poll registry every 24h for new versions of installed packages
- Send desktop notification (via OS notification API) when a critical CVE is published for any installed package — no dashboard visit needed
- gp watch --webhook <url> → POST a JSON payload to a Slack/Discord webhook on CVE discovery (useful for teams)
- Daemon mode: runs as a lightweight background process via gp watch --daemon

## ALIASING + SHORTCUTS  ★ NEW
- Built-in short aliases: gp i = install · gp rm = remove · gp up = update · gp s = search
- User-defined aliases in config: gp config set alias.dev "install --env dev"
- gp <alias> expands and runs the full command transparently

## CI/CD INTEGRATION  ★ NEW
- gp ci command: installs deps, runs audit, verifies lockfile, all in one — designed for CI pipelines
- Machine-readable output: gp audit --json / gp health --json for piping into dashboards
- Exit codes strictly follow severity: 0 = clean · 1 = warning · 2 = critical CVE or lockfile mismatch
- JUnit XML report output: gp audit --junit > report.xml for Jenkins / GitLab CI artifact ingestion
- GitHub Actions step summary integration: write a markdown summary to $GITHUB_STEP_SUMMARY automatically

## PLUGIN SYSTEM  ★ NEW
- gp plugin install <name> — install a GoPack plugin (itself a Go binary following a defined interface)
- Plugins can add new commands, new registry sources, or new security scanners
- gp plugin list / gp plugin remove <name>
- Plugin interface defined as a simple Go interface; no runtime scripting needed
- Sandboxed: plugins cannot write outside their own data directory


## TECH STACK
  Language:           Go 1.22+
  TUI framework:      Bubble Tea + Lip Gloss + Bubbles
  HTTP client:        Go net/http with HTTP/2 + connection pooling
  JSON parsing:       encoding/json + gjson for fast path queries
  Config:             Viper (TOML + env var override)
  CLI flags:          Cobra
  Progress bars:      mpb or Bubble Tea progress component
  Offline audit DB:   SQLite via mattn/go-sqlite3
  Delta updates:      xdelta3 bindings or pure-Go implementation
  Desktop notify:     go-toast (Windows) + osascript (macOS) + libnotify (Linux)
  Shell completion:   cobra completions (bash/zsh/fish/powershell)
  Archive handling:   archive/tar + compress/gzip (stdlib)
  Testing:            Go testing + testify + golden file tests
  Build/release:      GoReleaser (cross-compile macOS · Linux · Windows · ARM)

## CONFIG FILE SCHEMA
~/.gopack/config.toml:
  defaultEnv = "dev"
  parallelism = 8
  cacheDir = "~/.gopack/store"
  auditDbPath = "~/.gopack/osv.db"
  auditCacheTTL = "1h"
  notifyOnCVE = true
  webhookURL = ""
  licensePolicy = "permissive"    # permissive | strict | custom
  aliases = { dev = "install --env dev" }

## gopack.json SCHEMA
  {
    "name": "my-app",
    "version": "1.0.0",
    "scripts": { "build": "tsc", "test": "jest" },
    "dependencies": {},
    "devDependencies": {},
    "peerDependencies": {},
    "workspaces": ["packages/*"],
    "licensePolicy": { "allow": ["MIT","Apache-2.0"], "deny": ["GPL-3.0"] },
    "patches": { "lodash@4.17.21": "patches/lodash+4.17.21.patch" },
    "gopack": {
      "pinnedVersions": ["react"],
      "hoistPatterns": ["*"]
    }
  }

## DELIVERABLES
  1. Full Go codebase — all commands implemented and tested
  2. TUI matching the spec — keyboard nav, multi-select, security badges
  3. Offline mode — bundle export/import, local OSV DB
  4. Shell integration script for bash/zsh/fish/powershell
  5. README with install instructions, full command reference, ASCII demo
  6. GitHub Actions CI/CD pipeline using GoReleaser
  7. Homebrew formula + one-line curl install script
  8. Plugin interface spec document (PLUGINS.md)
  9. man page generated via cobra-man