package audit

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/layerscope/pkg/advisor"
	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/sbom"
	"github.com/alexandrmotologa/layerscope/pkg/security"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

// AuditResult wraps all analysis engines for a target image.
type AuditResult struct {
	ImageRef   string                        `json:"imageRef"`
	TotalBytes int64                         `json:"totalBytes"`
	AnalyzedAt time.Time                     `json:"analyzedAt"`
	Waste      *vfs.WastedSpaceSummary       `json:"waste"`
	Security   *security.SecurityAuditReport `json:"security"`
	Advisor    *advisor.AdvisorReport        `json:"advisor"`
	SBOM       *sbom.SBOMReport              `json:"sbom,omitempty"`
}

// GenerateFullAudit executes all analysis modules and aggregates the final result.
func GenerateFullAudit(
	img *oci.ImageAnalysis,
	snapshots []*vfs.LayerSnapshot,
	waste *vfs.WastedSpaceSummary,
	secReport *security.SecurityAuditReport,
	sbomReport *sbom.SBOMReport,
) *AuditResult {
	var finalTree *vfs.VFSTree
	if len(snapshots) > 0 {
		finalTree = snapshots[len(snapshots)-1].Tree
	}
	advisorReport := advisor.AnalyzeImageHeuristics(img, waste, secReport, finalTree)

	return &AuditResult{
		ImageRef:   img.Reference.Original,
		TotalBytes: img.TotalSizeBytes,
		AnalyzedAt: time.Now(),
		Waste:      waste,
		Security:   secReport,
		Advisor:    advisorReport,
		SBOM:       sbomReport,
	}
}

// FormatJSON outputs indented JSON for CI automation.
func (a *AuditResult) FormatJSON() ([]byte, error) {
	return json.MarshalIndent(a, "", "  ")
}

// FormatMarkdown outputs clean GitHub-flavored markdown for CI job summaries.
func (a *AuditResult) FormatMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# LayerScope Audit: %s\n\n", a.ImageRef))
	sb.WriteString(fmt.Sprintf("- **Efficiency Score:** %.1f%% (Grade: %s)\n", a.Advisor.EfficiencyScore, a.Advisor.Grade))
	sb.WriteString(fmt.Sprintf("- **Total Image Size:** %.2f MB\n", float64(a.TotalBytes)/(1024*1024)))
	sb.WriteString(fmt.Sprintf("- **Wasted Bytes:** %.2f MB (%.1f%%)\n",
		float64(a.Waste.TotalWastedBytes)/(1024*1024), a.Waste.WastedPercentage))
	sb.WriteString(fmt.Sprintf("- **Security Findings:** %d (Critical: %d, High: %d, Intermediate Leaks: %d)\n",
		a.Security.TotalFindings, a.Security.CriticalCount, a.Security.HighCount, a.Security.IntermediateLeaks))

	if a.SBOM != nil {
		sb.WriteString(fmt.Sprintf("- **Total Packages:** %d (OS: %d, Application: %d)\n\n",
			a.SBOM.TotalPackages, a.SBOM.OSPackages, a.SBOM.AppPackages))
	}

	if len(a.Security.Findings) > 0 {
		sb.WriteString("## Security Credential Leaks\n\n")
		sb.WriteString("| Severity | Rule | File | Layer | Intermediate Leak | Masked Value |\n")
		sb.WriteString("|---|---|---|---|---|---|\n")
		for _, f := range a.Security.Findings {
			leak := "No"
			if f.IsDeletedInFinal {
				leak = "**YES (Deleted in later layer)**"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | `%s` | %d | %s | `%s` |\n",
				f.Severity, f.RuleName, f.FilePath, f.LayerIndex, leak, f.MatchMasked))
		}
		sb.WriteString("\n")
	}

	if len(a.Advisor.Recommendations) > 0 {
		sb.WriteString("## Actionable Optimization Recommendations\n\n")
		for _, r := range a.Advisor.Recommendations {
			sb.WriteString(fmt.Sprintf("### %s [%s]\n", r.Title, r.Category))
			sb.WriteString(fmt.Sprintf("%s\n\n", r.Explanation))
			if r.RemediationSnippet != "" {
				sb.WriteString("```dockerfile\n" + r.RemediationSnippet + "\n```\n\n")
			}
		}
	}

	return sb.String()
}

// JUnit XML types
type jUnitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []jUnitTestSuite `xml:"testsuite"`
}

type jUnitTestSuite struct {
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	TestCases []jUnitTestCase `xml:"testcase"`
}

type jUnitTestCase struct {
	Classname string        `xml:"classname,attr"`
	Name      string        `xml:"name,attr"`
	Failure   *jUnitFailure `xml:"failure,omitempty"`
}

type jUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// FormatJUnit generates standard JUnit XML for CI test reports.
func (a *AuditResult) FormatJUnit(minEfficiency float64, failOnSecrets bool) ([]byte, error) {
	suite := jUnitTestSuite{
		Name: fmt.Sprintf("LayerScope: %s", a.ImageRef),
	}

	// Test 1: Efficiency Threshold
	tcEff := jUnitTestCase{
		Classname: "layerscope.efficiency",
		Name:      "ImageEfficiencyThreshold",
	}
	suite.Tests++
	if a.Advisor.EfficiencyScore < minEfficiency {
		suite.Failures++
		tcEff.Failure = &jUnitFailure{
			Message: fmt.Sprintf("Efficiency %.1f%% is below required minimum %.1f%%", a.Advisor.EfficiencyScore, minEfficiency),
			Type:    "EfficiencyFailure",
			Content: fmt.Sprintf("Wasted %d bytes out of %d total bytes", a.Waste.TotalWastedBytes, a.TotalBytes),
		}
	}
	suite.TestCases = append(suite.TestCases, tcEff)

	// Test 2: Secrets Check
	tcSec := jUnitTestCase{
		Classname: "layerscope.security",
		Name:      "NoLeakedSecrets",
	}
	suite.Tests++
	if failOnSecrets && a.Security.TotalFindings > 0 {
		suite.Failures++
		tcSec.Failure = &jUnitFailure{
			Message: fmt.Sprintf("Discovered %d sensitive credentials in image layers", a.Security.TotalFindings),
			Type:    "SecurityLeakFailure",
			Content: fmt.Sprintf("Found %d critical and %d intermediate leaks", a.Security.CriticalCount, a.Security.IntermediateLeaks),
		}
	}
	suite.TestCases = append(suite.TestCases, tcSec)

	suites := jUnitTestSuites{Suites: []jUnitTestSuite{suite}}
	return xml.MarshalIndent(suites, "", "  ")
}

// FormatHTML outputs a self-contained, standalone dark-themed HTML report.
func (a *AuditResult) FormatHTML() string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>LayerScope Audit: ` + a.ImageRef + `</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0b0f17; color: #e2e8f0; margin: 0; padding: 2rem; }
  .container { max-width: 1000px; margin: 0 auto; }
  .header { display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #1e293b; padding-bottom: 1.5rem; }
  .badge { background: #3b82f6; color: white; padding: 0.25rem 0.75rem; border-radius: 9999px; font-weight: 600; font-size: 0.875rem; }
  .badge.crit { background: #ef4444; }
  .badge.warn { background: #f59e0b; }
  .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 1rem; margin: 1.5rem 0; }
  .card { background: #131b2e; border: 1px solid #1e293b; border-radius: 0.5rem; padding: 1.25rem; }
  .card-val { font-size: 1.75rem; font-weight: 700; margin-top: 0.25rem; }
  table { width: 100%; border-collapse: collapse; margin: 1rem 0; }
  th, td { text-align: left; padding: 0.75rem; border-bottom: 1px solid #1e293b; }
  th { background: #131b2e; color: #94a3b8; font-size: 0.875rem; }
  pre { background: #050811; padding: 1rem; border-radius: 0.375rem; overflow-x: auto; color: #38bdf8; font-size: 0.875rem; }
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <div>
      <h1 style="margin:0 0 0.5rem 0;">LayerScope Audit Report</h1>
      <div style="color:#94a3b8; font-family: monospace;">` + a.ImageRef + `</div>
    </div>
    <div class="badge">Efficiency Grade: ` + a.Advisor.Grade + ` (` + fmt.Sprintf("%.1f%%", a.Advisor.EfficiencyScore) + `)</div>
  </div>

  <div class="grid">
    <div class="card">
      <div style="color:#94a3b8; font-size:0.875rem;">Total Image Size</div>
      <div class="card-val">` + fmt.Sprintf("%.2f MB", float64(a.TotalBytes)/(1024*1024)) + `</div>
    </div>
    <div class="card">
      <div style="color:#94a3b8; font-size:0.875rem;">Wasted Space</div>
      <div class="card-val" style="color:#f59e0b;">` + fmt.Sprintf("%.2f MB", float64(a.Waste.TotalWastedBytes)/(1024*1024)) + `</div>
    </div>
    <div class="card">
      <div style="color:#94a3b8; font-size:0.875rem;">Security Leaks</div>
      <div class="card-val" style="color:` + (func() string {
		if a.Security.TotalFindings > 0 {
			return "#ef4444"
		}
		return "#10b981"
	})() + `;">` + fmt.Sprintf("%d", a.Security.TotalFindings) + `</div>
    </div>
    <div class="card">
      <div style="color:#94a3b8; font-size:0.875rem;">Intermediate Leaks</div>
      <div class="card-val" style="color:#f87171;">` + fmt.Sprintf("%d", a.Security.IntermediateLeaks) + `</div>
    </div>
  </div>`)

	if len(a.Security.Findings) > 0 {
		sb.WriteString(`<h2>Security & Credential Findings</h2>
<table>
  <thead>
    <tr><th>Severity</th><th>Rule</th><th>File Path</th><th>Layer</th><th>Intermediate Leak</th><th>Masked Value</th></tr>
  </thead>
  <tbody>`)
		for _, f := range a.Security.Findings {
			isInter := "No"
			if f.IsDeletedInFinal {
				isInter = "<strong style='color:#ef4444;'>YES (Deleted in later layer)</strong>"
			}
			sb.WriteString(fmt.Sprintf("<tr><td><span class='badge crit'>%s</span></td><td>%s</td><td><code>%s</code></td><td>%d</td><td>%s</td><td><code>%s</code></td></tr>",
				f.Severity, f.RuleName, f.FilePath, f.LayerIndex, isInter, f.MatchMasked))
		}
		sb.WriteString(`</tbody></table>`)
	}

	if len(a.Advisor.Recommendations) > 0 {
		sb.WriteString(`<h2>Recommendations & Dockerfile Remediations</h2>`)
		for _, r := range a.Advisor.Recommendations {
			sb.WriteString(fmt.Sprintf(`<div class="card" style="margin-bottom:1rem;">
  <div style="font-weight:600; font-size:1.125rem;">%s <span style="font-size:0.875rem; color:#94a3b8;">[%s]</span></div>
  <p style="color:#cbd5e1; margin:0.5rem 0;">%s</p>`, r.Title, r.Category, r.Explanation))
			if r.RemediationSnippet != "" {
				sb.WriteString(fmt.Sprintf("<pre><code>%s</code></pre>", r.RemediationSnippet))
			}
			sb.WriteString(`</div>`)
		}
	}

	sb.WriteString(`</div></body></html>`)
	return sb.String()
}
