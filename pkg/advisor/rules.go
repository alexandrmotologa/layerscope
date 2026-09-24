package advisor

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/security"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

// AnalyzeImageHeuristics runs best-practice rules against the image layers, tree, and security report.
func AnalyzeImageHeuristics(
	img *oci.ImageAnalysis,
	waste *vfs.WastedSpaceSummary,
	secReport *security.SecurityAuditReport,
	finalTree *vfs.VFSTree,
) *AdvisorReport {
	score, grade := CalculateEfficiency(waste.TotalImageBytes, waste.TotalWastedBytes)

	report := &AdvisorReport{
		EfficiencyScore:       score,
		Grade:                 grade,
		TotalImageBytes:       waste.TotalImageBytes,
		TotalWastedBytes:      waste.TotalWastedBytes,
		PotentialSavingsBytes: waste.TotalWastedBytes,
	}

	// 1. Check Package Manager Cache Leftovers in final tree and wasted files
	var cacheBytes int64
	var cachePaths []string
	checkCachePath := func(p string, size int64) {
		pLower := strings.ToLower(p)
		if strings.Contains(pLower, "/var/lib/apt/lists") ||
			strings.Contains(pLower, "/var/cache/apt") ||
			strings.Contains(pLower, "/root/.npm") ||
			strings.Contains(pLower, "/root/.cache") ||
			strings.Contains(pLower, "/var/cache/apk") {
			cacheBytes += size
			if len(cachePaths) < 3 {
				cachePaths = append(cachePaths, p)
			}
		}
	}

	for _, f := range waste.TopWastedFiles {
		checkCachePath(f.Path, f.Size)
	}
	if finalTree != nil {
		for p, node := range finalTree.Index {
			if !node.IsDir {
				checkCachePath(p, node.Size)
			}
		}
	}

	if cacheBytes > 0 {
		report.Recommendations = append(report.Recommendations, &Recommendation{
			ID:                    "uncleaned-pkg-cache",
			Category:              CategorySizeWaste,
			Title:                 "Clean Package Manager Caches in the Same Layer",
			Severity:              "high",
			EstimatedSavingsBytes: cacheBytes,
			Explanation: fmt.Sprintf("Found %d KB of uncleaned package manager caches in files like %s. Package manager caches stored in intermediate layers inflate download size permanently.",
				cacheBytes/1024, strings.Join(cachePaths, ", ")),
			RemediationSnippet: `# Combine package installation and cache removal in a single RUN instruction:
# Debian/Ubuntu:
RUN apt-get update && apt-get install -y --no-install-recommends \
    package-name \
    && rm -rf /var/lib/apt/lists/*

# Alpine:
RUN apk add --no-cache package-name

# Node / NPM:
RUN npm ci --omit=dev && npm cache clean --force`,
		})
	}

	// 2. Check Build Cache Invalidation Ordering
	var seenCopyAll bool
	var copyAllIndex int
	for i, l := range img.Layers {
		cmd := strings.ToLower(l.Command)
		if strings.Contains(cmd, "copy .") || strings.Contains(cmd, "copy src") {
			seenCopyAll = true
			copyAllIndex = i
		}
		if seenCopyAll && (strings.Contains(cmd, "npm install") || strings.Contains(cmd, "npm ci") ||
			strings.Contains(cmd, "pip install") || strings.Contains(cmd, "go mod download")) {
			report.Recommendations = append(report.Recommendations, &Recommendation{
				ID:               "cache-invalidation-order",
				Category:         CategoryBuildCache,
				Title:            "Optimize Dockerfile Layer Caching Sequence",
				Severity:         "medium",
				TargetLayerIndex: copyAllIndex,
				Explanation:      "Application source code is copied into the image before installing dependencies. Any change to application source files invalidates the layer cache for all subsequent dependency downloads.",
				RemediationSnippet: `# Copy lockfiles first to preserve Docker layer caching:
COPY package*.json ./
RUN npm ci

# Copy remaining application source afterwards:
COPY . .`,
			})
			break
		}
	}

	// 3. Intermediate Layer Leaked Secrets
	if secReport != nil && secReport.IntermediateLeaks > 0 {
		report.Recommendations = append(report.Recommendations, &Recommendation{
			ID:       "intermediate-secret-leak",
			Category: CategorySecurity,
			Title:    "Use BuildKit Secret Mounts for Build Credentials",
			Severity: "high",
			Explanation: fmt.Sprintf("Found %d credential leaks in historical layers that were deleted in subsequent steps. Intermediate layers remain downloadable by anyone with image pull rights.",
				secReport.IntermediateLeaks),
			RemediationSnippet: `# Instead of COPY .env or passing tokens as ARG:
# Mount secrets directly during build without baking them into layers:
RUN --mount=type=secret,id=my_secret \
    export API_KEY=$(cat /run/secrets/my_secret) && \
    make build`,
		})
	}

	// 4. Non-Root Execution Check
	if secReport != nil && secReport.RunsAsRoot {
		report.Recommendations = append(report.Recommendations, &Recommendation{
			ID:          "run-as-root",
			Category:    CategorySecurity,
			Title:       "Execute Container as Non-Root User",
			Severity:    "medium",
			Explanation: "The container image runs as root (UID 0) by default. If a container breakout occurs, the attacker immediately gains root capabilities on the host kernel.",
			RemediationSnippet: `# Create and switch to a dedicated non-privileged user:
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser`,
		})
	}

	// 5. Multi-Stage Build Recommendation for Deleted/Overwritten Files
	if waste.DeletedFiles > 5 || waste.TotalWastedBytes > 5*1024*1024 {
		report.Recommendations = append(report.Recommendations, &Recommendation{
			ID:                    "use-multi-stage",
			Category:              CategorySizeWaste,
			Title:                 "Adopt Multi-Stage Builds for Compilers and Build Tools",
			Severity:              "high",
			EstimatedSavingsBytes: waste.TotalWastedBytes,
			Explanation: fmt.Sprintf("The image wastes %d KB across %d overwritten or deleted files. Building artifacts in an intermediate stage and copying only the final binary into a minimal runtime base eliminates build tool bloat.",
				waste.TotalWastedBytes/1024, waste.DeletedFiles+waste.OverwrittenFiles),
			RemediationSnippet: `# Multi-stage build example:
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
USER node
CMD ["node", "dist/server.js"]`,
		})
	}

	return report
}
