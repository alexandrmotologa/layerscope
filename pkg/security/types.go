package security

// Severity indicates the threat level of a detected secret or configuration issue.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
)

// SecretFinding captures a detected credential or sensitive token in an image layer.
type SecretFinding struct {
	ID                string   `json:"id"`
	RuleName          string   `json:"ruleName"`
	Severity          Severity `json:"severity"`
	FilePath          string   `json:"filePath"`
	LayerIndex        int      `json:"layerIndex"`
	Command           string   `json:"command"`
	LineNumber        int      `json:"lineNumber"`
	MatchMasked       string   `json:"matchMasked"`
	Description       string   `json:"description"`
	IsDeletedInFinal  bool     `json:"isDeletedInFinal"`
	SurroundingSecret string   `json:"surroundingSecret,omitempty"`
}

// SecurityAuditReport aggregates all findings and security checks for an image.
type SecurityAuditReport struct {
	TotalFindings      int              `json:"totalFindings"`
	CriticalCount      int              `json:"criticalCount"`
	HighCount          int              `json:"highCount"`
	MediumCount        int              `json:"mediumCount"`
	LowCount           int              `json:"lowCount"`
	IntermediateLeaks  int              `json:"intermediateLeaks"` // secrets deleted in later layers
	RunsAsRoot         bool             `json:"runsAsRoot"`
	SuidBinariesCount  int              `json:"suidBinariesCount"`
	SuidBinaries       []string         `json:"suidBinaries,omitempty"`
	Findings           []*SecretFinding `json:"findings"`
}
