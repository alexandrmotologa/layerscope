package vuln

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOSVQueryBatch_WithMockServer(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		resp := osvBatchResponse{
			Results: []struct {
				Vulns []struct {
					ID       string   `json:"id"`
					Summary  string   `json:"summary"`
					Details  string   `json:"details"`
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
			}{
				{
					Vulns: []struct {
						ID       string   `json:"id"`
						Summary  string   `json:"summary"`
						Details  string   `json:"details"`
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
					}{
						{
							ID:      "CVE-2024-12345",
							Summary: "Prototype pollution in lodash",
							Aliases: []string{"GHSA-xxxx"},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewClient()
	client.apiURL = mockServer.URL

	res, err := client.QueryBatch(context.Background(), []string{"pkg:npm/lodash@4.17.20"})
	if err != nil {
		t.Fatalf("QueryBatch failed: %v", err)
	}

	vulns := res["pkg:npm/lodash@4.17.20"]
	if len(vulns) != 1 {
		t.Fatalf("expected 1 vulnerability, got %d", len(vulns))
	}
	if vulns[0].ID != "CVE-2024-12345" {
		t.Errorf("expected CVE-2024-12345, got %s", vulns[0].ID)
	}
}
