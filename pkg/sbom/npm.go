package sbom

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type npmPackageLock struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Packages     map[string]npmPkgEntry `json:"packages"`
	Dependencies map[string]npmPkgEntry `json:"dependencies"`
}

type npmPkgEntry struct {
	Version      string                 `json:"version"`
	Resolved     string                 `json:"resolved"`
	Dependencies map[string]interface{} `json:"dependencies"`
}

// ParseNPMLockfile extracts packages from package-lock.json.
func ParseNPMLockfile(r io.Reader, manifestPath string) ([]*Package, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var lock npmPackageLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("unmarshal package-lock.json: %w", err)
	}

	var packages []*Package
	seen := make(map[string]bool)

	// v2 / v3 lockfile format
	for pkgKey, entry := range lock.Packages {
		if pkgKey == "" {
			continue // Root package
		}
		cleanName := strings.TrimPrefix(pkgKey, "node_modules/")
		if idx := strings.LastIndex(cleanName, "node_modules/"); idx >= 0 {
			cleanName = cleanName[idx+len("node_modules/"):]
		}

		if seen[cleanName+"@"+entry.Version] || entry.Version == "" {
			continue
		}
		seen[cleanName+"@"+entry.Version] = true

		packages = append(packages, &Package{
			Name:          cleanName,
			Version:       entry.Version,
			Type:          TypeNPM,
			PURL:          fmt.Sprintf("pkg:npm/%s@%s", cleanName, entry.Version),
			InstalledPath: manifestPath,
		})
	}

	// v1 lockfile format fallback if packages map was empty
	if len(packages) == 0 {
		for name, entry := range lock.Dependencies {
			if seen[name+"@"+entry.Version] || entry.Version == "" {
				continue
			}
			seen[name+"@"+entry.Version] = true

			packages = append(packages, &Package{
				Name:          name,
				Version:       entry.Version,
				Type:          TypeNPM,
				PURL:          fmt.Sprintf("pkg:npm/%s@%s", name, entry.Version),
				InstalledPath: manifestPath,
			})
		}
	}

	return packages, nil
}
