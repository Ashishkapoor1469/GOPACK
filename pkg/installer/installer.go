package installer

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Ashishkapoor1469/GOPACK/pkg/config"
	"github.com/Ashishkapoor1469/GOPACK/pkg/registry"
)

// PackageManifest represents gopack.json structure
type PackageManifest struct {
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	Scripts          map[string]string `json:"scripts,omitempty"`
	Dependencies     map[string]string `json:"dependencies,omitempty"`
	DevDependencies  map[string]string `json:"devDependencies,omitempty"`
	PeerDependencies map[string]string `json:"peerDependencies,omitempty"`
	Workspaces       []string          `json:"workspaces,omitempty"`
	LicensePolicy    struct {
		Allow []string `json:"allow,omitempty"`
		Deny  []string `json:"deny,omitempty"`
	} `json:"licensePolicy,omitempty"`
	Patches map[string]string `json:"patches,omitempty"`
}

// LockfileDependency represents a dependency entry in gopack.lock
type LockfileDependency struct {
	Version      string            `json:"version"`
	Resolved     string            `json:"resolved"`
	Integrity    string            `json:"integrity"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

// Lockfile represents gopack.lock structure
type Lockfile struct {
	LockfileVersion int                           `json:"lockfileVersion"`
	Dependencies    map[string]LockfileDependency `json:"dependencies"`
}

// ProgressInfo tracks the status of a single package download/installation
type ProgressInfo struct {
	PackageName string
	Percent     float64 // 0 to 100
	Status      string  // "pending", "downloading", "extracting", "done", "failed"
	Error       string
}

// InstallManager handles the orchestration of package installation
type InstallManager struct {
	client     *registry.RegistryClient
	config     *config.Config
	manifest   *PackageManifest
	lockfile   *Lockfile
	manifestMu sync.Mutex
	lockfileMu sync.Mutex
}

func NewInstallManager(cfg *config.Config) *InstallManager {
	return &InstallManager{
		client: registry.NewRegistryClient(),
		config: cfg,
		manifest: &PackageManifest{
			Name:            "my-app",
			Version:         "1.0.0",
			Dependencies:    make(map[string]string),
			DevDependencies: make(map[string]string),
		},
		lockfile: &Lockfile{
			LockfileVersion: 1,
			Dependencies:    make(map[string]LockfileDependency),
		},
	}
}

func (m *InstallManager) Dependencies() map[string]LockfileDependency {
	m.lockfileMu.Lock()
	defer m.lockfileMu.Unlock()
	return m.lockfile.Dependencies
}

func (m *InstallManager) ManifestDependencies() map[string]string {
	m.manifestMu.Lock()
	defer m.manifestMu.Unlock()
	return m.manifest.Dependencies
}

func (m *InstallManager) ManifestScripts() map[string]string {
	m.manifestMu.Lock()
	defer m.manifestMu.Unlock()
	return m.manifest.Scripts
}

func (m *InstallManager) LicensePolicy() (allowed []string, denied []string) {
	m.manifestMu.Lock()
	defer m.manifestMu.Unlock()
	return m.manifest.LicensePolicy.Allow, m.manifest.LicensePolicy.Deny
}

// LoadManifest reads gopack.json or falls back to package.json
func (m *InstallManager) LoadManifest() error {
	m.manifestMu.Lock()
	defer m.manifestMu.Unlock()

	paths := []string{"gopack.json", "package.json"}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			var manifest PackageManifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				return err
			}
			if manifest.Dependencies == nil {
				manifest.Dependencies = make(map[string]string)
			}
			if manifest.DevDependencies == nil {
				manifest.DevDependencies = make(map[string]string)
			}
			m.manifest = &manifest
			return nil
		}
	}
	return nil
}

// SaveManifest writes gopack.json
func (m *InstallManager) SaveManifest() error {
	m.manifestMu.Lock()
	defer m.manifestMu.Unlock()

	data, err := json.MarshalIndent(m.manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("gopack.json", data, 0644)
}

// LoadLockfile reads gopack.lock
func (m *InstallManager) LoadLockfile() error {
	m.lockfileMu.Lock()
	defer m.lockfileMu.Unlock()

	if _, err := os.Stat("gopack.lock"); err == nil {
		data, err := os.ReadFile("gopack.lock")
		if err != nil {
			return err
		}
		var lockfile Lockfile
		if err := json.Unmarshal(data, &lockfile); err != nil {
			return err
		}
		if lockfile.Dependencies == nil {
			lockfile.Dependencies = make(map[string]LockfileDependency)
		}
		m.lockfile = &lockfile
	}
	return nil
}

// SaveLockfile writes gopack.lock
func (m *InstallManager) SaveLockfile() error {
	m.lockfileMu.Lock()
	defer m.lockfileMu.Unlock()

	data, err := json.MarshalIndent(m.lockfile, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("gopack.lock", data, 0644)
}

// ResolveDependencyTree builds a list of all packages to be installed (including transitives)
func (m *InstallManager) ResolveDependencyTree(pkgs map[string]string) (map[string]registry.PackageVersionInfo, error) {
	resolved := make(map[string]registry.PackageVersionInfo)
	var resolveQueue []struct {
		name string
		spec string
	}

	for k, v := range pkgs {
		resolveQueue = append(resolveQueue, struct{ name, spec string }{k, v})
	}

	for len(resolveQueue) > 0 {
		current := resolveQueue[0]
		resolveQueue = resolveQueue[1:]

		// Avoid infinite loops and redundant fetching
		if _, ok := resolved[current.name]; ok {
			continue
		}

		meta, err := m.client.GetPackage(current.name)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch metadata for %s: %w", current.name, err)
		}

		// Find suitable version
		targetVer := current.spec
		if targetVer == "latest" || targetVer == "" {
			if lVer, ok := meta.DistTags["latest"]; ok {
				targetVer = lVer
			} else {
				return nil, fmt.Errorf("no latest version found for %s", current.name)
			}
		}

		// Check matching version in versions map
		verInfo, ok := meta.Versions[targetVer]
		if !ok {
			// Basic semver fallback - pick first matching version if spec is simple caret/tilde.
			// For simplicity in the prototype, we fall back to latest or exact version.
			// In production this would use semver parser.
			found := false
			for vStr, info := range meta.Versions {
				if strings.Contains(targetVer, "^") || strings.Contains(targetVer, "~") {
					// Clean prefix for a rough match
					cleanSpec := strings.TrimLeft(targetVer, "^~")
					if strings.HasPrefix(vStr, cleanSpec) {
						verInfo = info
						found = true
						break
					}
				}
			}
			if !found {
				// Final fallback, pick latest version
				if lVer, ok := meta.DistTags["latest"]; ok {
					verInfo = meta.Versions[lVer]
				} else {
					return nil, fmt.Errorf("version %s not found for %s", targetVer, current.name)
				}
			}
		}

		resolved[current.name] = verInfo

		// Add dependencies to resolve queue
		for depName, depVer := range verInfo.Dependencies {
			resolveQueue = append(resolveQueue, struct{ name, spec string }{depName, depVer})
		}
	}

	return resolved, nil
}

// ExtractTarball extracts a tar.gz reader to a destination directory
func ExtractTarball(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// npm tarballs place files inside a "package/" folder. We strip this prefix.
		relPath := header.Name
		parts := strings.Split(relPath, "/")
		if len(parts) > 1 && parts[0] == "package" {
			relPath = filepath.Join(parts[1:]...)
		}

		target := filepath.Join(dest, relPath)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

// InstallResolvedPackages installs the list of resolved packages concurrently and reports progress
func (m *InstallManager) InstallResolvedPackages(resolved map[string]registry.PackageVersionInfo, progressChan chan<- []ProgressInfo, isDev bool) error {
	defer close(progressChan)

	progressMap := make(map[string]ProgressInfo)
	var progressMu sync.Mutex

	sendProgress := func() {
		list := make([]ProgressInfo, 0, len(progressMap))
		for _, v := range progressMap {
			list = append(list, v)
		}
		progressChan <- list
	}

	for name := range resolved {
		progressMap[name] = ProgressInfo{
			PackageName: name,
			Percent:     0,
			Status:      "pending",
		}
	}
	sendProgress()

	// Setup task worker pool
	parallelism := m.config.Parallelism
	if parallelism <= 0 {
		parallelism = 8
	}

	sem := make(chan struct{}, parallelism)
	var wg sync.WaitGroup
	var installErr error
	var errMu sync.Mutex

	for name, verInfo := range resolved {
		wg.Add(1)
		go func(pkgName string, info registry.PackageVersionInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Set status to downloading
			progressMu.Lock()
			progressMap[pkgName] = ProgressInfo{
				PackageName: pkgName,
				Percent:     10,
				Status:      "downloading",
			}
			sendProgress()
			progressMu.Unlock()

			// Check if already in cache store
			cachePkgDir := filepath.Join(m.config.CacheDir, fmt.Sprintf("%s@%s", pkgName, info.Version))
			cacheMarkerFile := filepath.Join(cachePkgDir, ".gopack-ok")

			isCached := false
			if _, err := os.Stat(cacheMarkerFile); err == nil {
				isCached = true
			}

			if !isCached {
				// Download tarball
				body, contentLength, err := m.client.DownloadTarball(info.Dist.Tarball)
				if err != nil {
					errMu.Lock()
					installErr = fmt.Errorf("failed to download %s: %w", pkgName, err)
					errMu.Unlock()

					progressMu.Lock()
					progressMap[pkgName] = ProgressInfo{
						PackageName: pkgName,
						Status:      "failed",
						Error:       err.Error(),
					}
					sendProgress()
					progressMu.Unlock()
					return
				}
				defer body.Close()

				// Setup path and progress tracking reader
				if err := os.MkdirAll(cachePkgDir, 0755); err != nil {
					errMu.Lock()
					installErr = err
					errMu.Unlock()
					return
				}

				// Copy body and compute hash while downloading
				tempTarPath := filepath.Join(cachePkgDir, "package.tgz")
				tempFile, err := os.Create(tempTarPath)
				if err != nil {
					errMu.Lock()
					installErr = err
					errMu.Unlock()
					return
				}

				hasher := sha256.New()
				mw := io.MultiWriter(tempFile, hasher)

				buf := make([]byte, 32*1024)
				var written int64
				for {
					nr, er := body.Read(buf)
					if nr > 0 {
						nw, ew := mw.Write(buf[0:nr])
						if nw > 0 {
							written += int64(nw)
						}
						if ew != nil {
							err = ew
							break
						}
						if nr != nw {
							err = io.ErrShortWrite
							break
						}
					}
					if er != nil {
						if er != io.EOF {
							err = er
						}
						break
					}

					// Update progress percentage
					if contentLength > 0 {
						pct := (float64(written) / float64(contentLength)) * 70.0
						progressMu.Lock()
						progressMap[pkgName] = ProgressInfo{
							PackageName: pkgName,
							Percent:     10 + pct,
							Status:      "downloading",
						}
						sendProgress()
						progressMu.Unlock()
					}
				}
				tempFile.Close()

				if err != nil {
					errMu.Lock()
					installErr = fmt.Errorf("download error for %s: %w", pkgName, err)
					errMu.Unlock()
					return
				}

				// Verify SHA1/SHA256 (from registry)
				computedHash := hex.EncodeToString(hasher.Sum(nil))
				// We can compare info.Dist.Shasum or info.Dist.Integrity. For now, we trust downloaded integrity.

				// Update state to extracting
				progressMu.Lock()
				progressMap[pkgName] = ProgressInfo{
					PackageName: pkgName,
					Percent:     80,
					Status:      "extracting",
				}
				sendProgress()
				progressMu.Unlock()

				// Extract files
				tf, err := os.Open(tempTarPath)
				if err != nil {
					errMu.Lock()
					installErr = err
					errMu.Unlock()
					return
				}
				err = ExtractTarball(tf, cachePkgDir)
				tf.Close()
				os.Remove(tempTarPath) // clean up tar archive

				if err != nil {
					errMu.Lock()
					installErr = fmt.Errorf("extraction error for %s: %w", pkgName, err)
					errMu.Unlock()
					return
				}

				// Write cached package manifest or mark ok
				os.WriteFile(cacheMarkerFile, []byte(computedHash), 0644)
			}

			// Deploy package to local node_modules
			nodeModulesDest := filepath.Join("node_modules", pkgName)
			if err := os.MkdirAll(filepath.Dir(nodeModulesDest), 0755); err != nil {
				errMu.Lock()
				installErr = err
				errMu.Unlock()
				return
				
			}

			// Clean existing
			os.RemoveAll(nodeModulesDest)

			// Simple copy or symlink (we do recursive copy for full platform compatibility, especially on Windows)
			err := CopyDir(cachePkgDir, nodeModulesDest)
			if err != nil {
				errMu.Lock()
				installErr = fmt.Errorf("failed to copy %s to node_modules: %w", pkgName, err)
				errMu.Unlock()
				return
			}

			// Record to lockfile
			m.lockfileMu.Lock()
			m.lockfile.Dependencies[pkgName] = LockfileDependency{
				Version:   info.Version,
				Resolved:  info.Dist.Tarball,
				Integrity: info.Dist.Integrity,
			}
			m.lockfileMu.Unlock()

			// Update state to done
			progressMu.Lock()
			progressMap[pkgName] = ProgressInfo{
				PackageName: pkgName,
				Percent:     100,
				Status:      "done",
			}
			sendProgress()
			progressMu.Unlock()

		}(name, verInfo)
	}

	wg.Wait()

	if installErr != nil {
		return installErr
	}



	return nil
}

// CopyDir recursively copies a directory tree.
func CopyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// CopyFile copies a single file from src to dst.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Sync()
}

// Install method to trigger full install of specific lists
func (m *InstallManager) Install(packages []string, isDev bool) error {
	if err := m.LoadManifest(); err != nil {
		return err
	}
	m.LoadLockfile()

	// Build map of targets
	targets := make(map[string]string)
	for _, p := range packages {
		parts := strings.Split(p, "@")
		if len(parts) == 1 {
			targets[parts[0]] = "latest"
		} else if len(parts) == 2 {
			targets[parts[0]] = parts[1]
		}
	}

	// If no packages specified, install everything from manifest
	if len(packages) == 0 {
		for k, v := range m.manifest.Dependencies {
			targets[k] = v
		}
		if !isDev {
			for k, v := range m.manifest.DevDependencies {
				targets[k] = v
			}
		}
	}

	if len(targets) == 0 {
		return nil
	}

	// Resolve tree
	resolved, err := m.ResolveDependencyTree(targets)
	if err != nil {
		return err
	}

	progressChan := make(chan []ProgressInfo)

	// In a CLI, we can print progress to stdout
	// Let's run the install in background and consume progress
	var errResult error
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		errResult = m.InstallResolvedPackages(resolved, progressChan, isDev)
	}()

	// Monitor progress and write output block
	lastLinesCount := 0
	for pList := range progressChan {
		// Clean previous lines
		for i := 0; i < lastLinesCount; i++ {
			fmt.Print("\033[F\033[K") // Cursor up one line & clear line
		}
		lastLinesCount = 0

		fmt.Println("┌─ installing packages ──────────────────────────────┐")
		lastLinesCount++
		for _, p := range pList {
			statusSymbol := "░"
			if p.Status == "done" {
				statusSymbol = "✔"
			} else if p.Status == "failed" {
				statusSymbol = "✘"
			} else {
				statusSymbol = "●"
			}

			// Render standard bar
			barLength := 20
			filledLength := int(p.Percent / 100 * float64(barLength))
			bar := strings.Repeat("█", filledLength) + strings.Repeat("░", barLength-filledLength)

			fmt.Printf("│  %s %-18s  %s  %3.0f%%  │\n", statusSymbol, p.PackageName, bar, p.Percent)
			lastLinesCount++
		}
		fmt.Println("└────────────────────────────────────────────────────┘")
		lastLinesCount++
	}

	wg.Wait()

	if errResult != nil {
		return errResult
	}

	// Add new packages to manifest
	m.manifestMu.Lock()
	for _, p := range packages {
		parts := strings.Split(p, "@")
		name := parts[0]
		ver := "latest"
		if len(parts) == 2 {
			ver = parts[1]
		}
		// Try to look up what version we resolved it to, so we can save exact version range
		if resolvedInfo, ok := resolved[name]; ok {
			ver = "^" + resolvedInfo.Version
		}

		if isDev {
			m.manifest.DevDependencies[name] = ver
		} else {
			m.manifest.Dependencies[name] = ver
		}
	}
	m.manifestMu.Unlock()

	if err := m.SaveManifest(); err != nil {
		return err
	}

	if err := m.SaveLockfile(); err != nil {
		return err
	}

	return nil
}

// Remove uninstalls package and updates manifest/lockfile
func (m *InstallManager) Remove(packageName string) error {
	if err := m.LoadManifest(); err != nil {
		return err
	}
	m.LoadLockfile()

	m.manifestMu.Lock()
	delete(m.manifest.Dependencies, packageName)
	delete(m.manifest.DevDependencies, packageName)
	m.manifestMu.Unlock()

	m.lockfileMu.Lock()
	delete(m.lockfile.Dependencies, packageName)
	m.lockfileMu.Unlock()

	os.RemoveAll(filepath.Join("node_modules", packageName))

	if err := m.SaveManifest(); err != nil {
		return err
	}
	if err := m.SaveLockfile(); err != nil {
		return err
	}

	return nil
}
