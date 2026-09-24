package sbom

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ParsePythonRequirements parses standard requirements.txt lines with pinned versions (==).
func ParsePythonRequirements(r io.Reader, manifestPath string) ([]*Package, error) {
	scanner := bufio.NewScanner(r)
	var packages []*Package

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and flags
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}

		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		// Handle pinned requirement foo==1.2.3
		parts := strings.Split(line, "==")
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			version := strings.TrimSpace(parts[1])
			packages = append(packages, &Package{
				Name:          name,
				Version:       version,
				Type:          TypePyPI,
				PURL:          fmt.Sprintf("pkg:pypi/%s@%s", name, version),
				InstalledPath: manifestPath,
			})
		} else {
			// Unpinned or comparison requirement
			name := strings.FieldsFunc(line, func(r rune) bool {
				return r == '>' || r == '<' || r == '~' || r == '=' || r == '!' || r == ';'
			})[0]
			name = strings.TrimSpace(name)
			if name != "" {
				packages = append(packages, &Package{
					Name:          name,
					Version:       "unspecified",
					Type:          TypePyPI,
					PURL:          fmt.Sprintf("pkg:pypi/%s", name),
					InstalledPath: manifestPath,
				})
			}
		}
	}

	return packages, scanner.Err()
}
