package advisor

// Category classifies the type of optimization recommendation.
type Category string

const (
	CategorySizeWaste    Category = "Size Optimization"
	CategorySecurity     Category = "Security Hardening"
	CategoryBuildCache   Category = "Build Cache Efficiency"
	CategoryBestPractice Category = "Dockerfile Best Practice"
)

// Recommendation provides concrete remediation instructions and estimated byte savings.
type Recommendation struct {
	ID                    string   `json:"id"`
	Category              Category `json:"category"`
	Title                 string   `json:"title"`
	Severity              string   `json:"severity"` // "high", "medium", "low", "info"
	EstimatedSavingsBytes int64    `json:"estimatedSavingsBytes"`
	Explanation           string   `json:"explanation"`
	RemediationSnippet    string   `json:"remediationSnippet,omitempty"`
	TargetLayerIndex      int      `json:"targetLayerIndex,omitempty"`
}

// AdvisorReport compiles overall image score and optimization recommendations.
type AdvisorReport struct {
	EfficiencyScore       float64           `json:"efficiencyScore"` // 0.0 - 100.0%
	Grade                 string            `json:"grade"`           // A+, A, B, C, D, F
	TotalImageBytes       int64             `json:"totalImageBytes"`
	TotalWastedBytes      int64             `json:"totalWastedBytes"`
	PotentialSavingsBytes int64             `json:"potentialSavingsBytes"`
	Recommendations       []*Recommendation `json:"recommendations"`
}
