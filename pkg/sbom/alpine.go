package sbom

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseAlpineInstalled parses an Alpine Linux /lib/apk/db/installed file.
func ParseAlpineInstalled(r io.Reader) ([]*Package, error) {
	scanner := bufio.NewScanner(r)
	var packages []*Package
	current := &Package{Type: TypeAlpine}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			if current.Name != "" {
				current.PURL = fmt.Sprintf("pkg:apk/alpine/%s@%s", current.Name, current.Version)
				packages = append(packages, current)
			}
			current = &Package{Type: TypeAlpine}
			continue
		}

		if len(line) < 3 || line[1] != ':' {
			continue
		}

		key := line[0]
		val := strings.TrimSpace(line[2:])

		switch key {
		case 'P':
			current.Name = val
		case 'V':
			current.Version = val
		case 'T':
			current.Description = val
		case 'L':
			current.License = val
		case 'I':
			if sz, err := strconv.ParseInt(val, 10, 64); err == nil {
				current.Size = sz
			}
		}
	}

	if current.Name != "" {
		current.PURL = fmt.Sprintf("pkg:apk/alpine/%s@%s", current.Name, current.Version)
		packages = append(packages, current)
	}

	return packages, scanner.Err()
}
