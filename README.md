# GoPack 📦

**GoPack** is a production-grade, cross-platform package manager CLI written in Go (Golang) designed to rival npm, pnpm, and Bun in speed, security, and developer experience. It installs, searches, and manages JS/TS packages and full-stack frameworks (Next.js, Svelte, React, etc.) from a single unified tool.

---

## ⚡ Features

- **No AI / 100% Offline-Safe**: Fully deterministic package installation and security auditing.
- **High-Performance**: Concurrently downloads packages using goroutines.
- **Embedded Security Badges**: Scans every package in parallel using OSV database APIs.
- **Workspace Support**: Built-in monorepo management with topologically sorted script execution.
- **Compliance Engine**: Automatically validates licenses of installed dependencies against policies.

---

## 🚀 One-Line Installation

Anyone can install GoPack instantly using one of the following commands:

### For Go Users (Cross-Platform)
```bash
go install github.com/hp/goPack@latest
```

### For Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/hp/goPack/main/install.ps1 | iex
```

### For macOS / Linux (Bash)
```bash
curl -fsSL https://raw.githubusercontent.com/hp/goPack/main/install.sh | sh
```

---

## 🛠️ CLI Command Reference

| Command | Description |
| :--- | :--- |
| `gp` | Launches the interactive TUI search and selection menu |
| `gp /` | Launches the TUI with the search bar focused instantly |
| `gp <pkg> [pkg2]` | Installs package(s) directly, bypassing the TUI |
| `gp install <pkg>` | Installs latest stable package |
| `gp remove <pkg>` | Uninstalls package and updates gopack.json/gopack.lock |
| `gp list` | Lists all installed dependencies with their security status |
| `gp audit` | Runs vulnerability scan and shows risk summaries |
| `gp audit --sync-db` | Downloads OSV DB snapshot to local SQLite for offline audits |
| `gp graph` | Visualizes the dependency tree with circularity indicators |
| `gp health` | Shows full project health dashboard (licenses, updates, security) |
| `gp licenses` | Inspects license compliance against allowed/denied lists |
| `gp run <script>` | Runs a custom script defined in gopack.json |
| `gp config list` | Lists all configurations inside ~/.gopack/config.toml |

---

## 🎨 Interactive Terminal UI (TUI)

If you run `gp` without any arguments, it opens a Bubble Tea interactive TUI containing:
1. **Side-by-side search panels** for `[FRAMEWORK]` and `[PACKAGES]`.
2. **Keyboard-driven navigation** (`↑`/`↓` to navigate, `←`/`→` to switch tabs).
3. **Space Selection**: Tag multiple packages and press `Enter` to install them concurrently.
4. **Security Badge**: Real-time coloring indicators of package vulnerabilities on the right column.

---

## 📖 How to Use GoPack (Step-by-Step)

### 1. Scaffold a New Project
Create a boilerplate template for your favorite framework (e.g. React):
```bash
gp create react
```
This prompts you for TypeScript, CSS framework, and Git initialization, setting up a template directories list and `gopack.json`.

### 2. Install Dependencies
Install packages directly from the command-line interface:
```bash
# Install a single package
gp install chalk

# Install multiple packages concurrently
gp lodash express axios
```
GoPack resolves nested dependency trees, pulls package assets in parallel using Go routines, caches files inside `~/.gopack/store/`, and creates `gopack.lock`.

### 3. Manage Package Manifests
View all installed packages along with their security badges:
```bash
gp list
```
See the full dependency tree, noting duplicate versions or circularities:
```bash
gp graph
```

### 4. Run Custom Scripts
Execute custom scripts defined in `gopack.json` (runs pre and post hooks automatically):
```bash
gp run dev
```

### 5. Check Project Security & Health
Run audits on dependencies to check for known CVE advisories:
```bash
gp audit
```
Sync vulnerability database definitions for offline security scans:
```bash
gp audit --sync-db
```
View the visual overall project dashboard score:
```bash
gp health
```
Check if all installed packages adhere to your licensing policies:
```bash
gp licenses --check strict
```
