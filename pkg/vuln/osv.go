package vuln

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Vulnerability represents an identified CVE/GHSA advisory from OSV.dev.
type Vulnerability struct {
	ID        string   `json:"id"`        // e.g. CVE-2024-XXXX, GHSA-XXXX
	Summary   string   `json:"summary"`
	Details   string   `json:"details,omitempty"`
	Severity  string   `json:"severity"` // CRITICAL, HIGH, MEDIUM, LOW
	CVSSScore float64  `json:"cvssScore,omitempty"`
	FixedIn   string   `json:"fixedIn,omitempty"`
	PURL      string   `json:"purl"`
	Aliases   []string `json:"aliases,omitempty"`
}

// OSVQuery represents an individual component query sent to OSV.dev batch API.
type OSVQuery struct {
	Package struct {
		PURL string `json:"purl,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"package"`
	Version string `json:"version,omitempty"`
}

type osvBatchRequest struct {
	Queries []OSVQuery `json:"queries"`
}

type osvBatchResponse struct {
	Results []struct {
		Vulns []struct {
			ID       string `json:"id"`
			Summary  string `json:"summary"`
			Details  string `json:"details"`
			Aliases  []string `json:"aliases"`
			Affected []struct {
				Ranges []struct {
					Type   string `json:"type"`
					Events []struct {
						Introduced string `json:"introduced,omitempty"`
						Fixed      string `json:"fixed,omitempty"`
					} `json:"events"`
				} `json:"ranges"`
			} `json:"affected"`
			Severity []struct {
				Type  string `json:"type"`
				Score string `json:"score"`
			} `json:"severity"`
		} `json:"vulns"`
	} `json:"results"`
}

// Client interacts with OSV.dev vulnerability database.
type Client struct {
	httpClient *http.Client
	apiURL     string
}

// NewClient initializes a client with resilient timeout.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		apiURL:     "https://api.osv.dev/v1/querybatch",
	}
}

// QueryBatch fetches advisories for a list of Package URLs (PURLs).
func (c *Client) QueryBatch(ctx context.Context, purls []string) (map[string][]Vulnerability, error) {
	if len(purls) == 0 {
		return nil, nil
	}

	// Limit to max 100 queries per batch
	limit := 100
	if len(purls) < limit {
		limit = len(purls)
	}

	reqBody := osvBatchRequest{
		Queries: make([]OSVQuery, limit),
	}
	for i := 0; i < limit; i++ {
		reqBody.Queries[i].Package.PURL = purls[i]
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Network errors or offline mode: return gracefully without crashing
		return nil, fmt.Errorf("osv query error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("osv api returned status %d", resp.StatusCode)
	}

	var batchResp osvBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&batchResp); err != nil {
		return nil, err
	}

	results := make(map[string][]Vulnerability)
	for i, res := range batchResp.Results {
		if i >= len(purls) {
			break
		}
		purl := purls[i]

		for _, v := range res.Vulns {
			fixed := ""
			for _, aff := range v.Affected {
				for _, r := range aff.Ranges {
					for _, ev := range r.Events {
						if ev.Fixed != "" {
							fixed = ev.Fixed
							break
						}
					}
					if fixed != "" {
						break
					}
				}
				if fixed != "" {
					break
				}
			}

			// Estimate severity from ID or severity blocks
			sev := "MEDIUM"
			if len(v.Severity) > 0 {
				sev = v.Severity[0].Type
			}

			results[purl] = append(results[purl], Vulnerability{
				ID:       v.ID,
				Summary:  v.Summary,
				Details:  v.Details,
				Severity: sev,
				FixedIn:  fixed,
				PURL:     purl,
				Aliases:  v.Aliases,
			})
		}
	}

	return results, nil
}
