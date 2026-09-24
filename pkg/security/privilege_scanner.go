package security

import (
	"os"
	"strings"

	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

// AuditPrivileges checks container default execution user and locates SUID/SGID binaries.
func AuditPrivileges(cfg *oci.ImageConfig, finalTree *vfs.VFSTree) (bool, []string) {
	runsAsRoot := true
	user := strings.TrimSpace(cfg.Config.User)
	if user != "" && user != "0" && user != "root" {
		runsAsRoot = false
	}

	var suidBinaries []string
	if finalTree != nil {
		for p, node := range finalTree.Index {
			if node.IsDir {
				continue
			}
			// Check SUID or SGID permissions
			if node.Mode&os.ModeSetuid != 0 || node.Mode&os.ModeSetgid != 0 {
				suidBinaries = append(suidBinaries, p)
			}
		}
	}

	return runsAsRoot, suidBinaries
}
