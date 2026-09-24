package sbom

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseDebianStatus parses a Debian/Ubuntu /var/lib/dpkg/status file.
func ParseDebianStatus(r io.Reader) ([]*Package, error) {
	scanner := bufio.NewScanner(r)
	var packages []*Package
	current := &Package{Type: TypeDebian}
	isInstalled := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			if current.Name != "" && isInstalled {
				current.PURL = fmt.Sprintf("pkg:deb/debian/%s@%s", current.Name, current.Version)
				packages = append(packages, current)
			}
			current = &Package{Type: TypeDebian}
			isInstalled = false
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch strings.ToLower(key) {
		case "package":
			current.Name = val
		case "version":
			current.Version = val
		case "description":
			current.Description = val
		case "status":
			if strings.Contains(val, "installed") && !strings.Contains(val, "not-installed") {
				isInstalled = true
			}
		case "installed-size":
			if szKb, err := strconv.ParseInt(val, 10, 64); err == nil {
				current.Size = szKb * 1024 // dpkg records in KB
			}
		}
	}

	if current.Name != "" && isInstalled {
		current.PURL = fmt.Sprintf("pkg:deb/debian/%s@%s", current.Name, current.Version)
		packages = append(packages, current)
	}

	return packages, scanner.Err()
}
