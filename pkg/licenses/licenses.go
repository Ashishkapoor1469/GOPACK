package licenses

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type LicenseResult struct {
	PackageName string
	LicenseType string
	Status      string // "ALLOWED", "DENIED", "WARNING", "UNKNOWN"
}

// CheckCompliance scans the node_modules directory for package licenses and evaluates them against the rules
func CheckCompliance(allowed []string, denied []string) ([]LicenseResult, error) {
	var results []LicenseResult

	// Read node_modules directory
	dirs, err := os.ReadDir("node_modules")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no packages installed yet
		}
		return nil, err
	}

	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}

		pkgName := d.Name()

		// Read package.json
		pkgJSONPath := filepath.Join("node_modules", pkgName, "package.json")
		if _, err := os.Stat(pkgJSONPath); err != nil {
			// Try scoped package folder
			if strings.HasPrefix(pkgName, "@") {
				subDirs, err := os.ReadDir(filepath.Join("node_modules", pkgName))
				if err == nil {
					for _, sd := range subDirs {
						if sd.IsDir() {
							subPkgName := pkgName + "/" + sd.Name()
							subPkgJSONPath := filepath.Join("node_modules", subPkgName, "package.json")
							if res, err := evaluatePackage(subPkgJSONPath, subPkgName, allowed, denied); err == nil {
								results = append(results, *res)
							}
						}
					}
				}
			}
			continue
		}

		if res, err := evaluatePackage(pkgJSONPath, pkgName, allowed, denied); err == nil {
			results = append(results, *res)
		}
	}

	return results, nil
}

func evaluatePackage(path, name string, allowed, denied []string) (*LicenseResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Dynamic parse to find license field
	var pkg struct {
		License interface{} `json:"license"` // can be string or struct
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	var licStr string
	switch v := pkg.License.(type) {
	case string:
		licStr = v
	case map[string]interface{}:
		if t, ok := v["type"].(string); ok {
			licStr = t
		}
	}

	if licStr == "" {
		licStr = "UNKNOWN"
	}

	status := "ALLOWED"

	// Match against allowed/denied lists
	licStrUpper := strings.ToUpper(licStr)

	// Check denied first
	for _, d := range denied {
		if strings.ToUpper(d) == licStrUpper {
			status = "DENIED"
			break
		}
	}

	if status != "DENIED" && len(allowed) > 0 {
		found := false
		for _, a := range allowed {
			if strings.ToUpper(a) == licStrUpper {
				found = true
				break
			}
		}
		if !found {
			status = "WARNING" // not explicitly allowed
		}
	}

	return &LicenseResult{
		PackageName: name,
		LicenseType: licStr,
		Status:      status,
	}, nil
}
