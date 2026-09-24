package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexandrmotologa/layerscope/pkg/advisor"
	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/security"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

func TestTUIModel_Navigation(t *testing.T) {
	sampleImg, err := oci.GenerateSampleImage(nil)
	if err != nil {
		t.Fatalf("GenerateSampleImage failed: %v", err)
	}

	snapshots := vfs.BuildLayerSnapshots(sampleImg)
	wasteSummary := vfs.CalculateWastedSpace(snapshots, nil)
	secReport := &security.SecurityAuditReport{}
	finalTree := snapshots[len(snapshots)-1].Tree
	advReport := advisor.AnalyzeImageHeuristics(sampleImg, wasteSummary, secReport, finalTree)

	m := NewModel(sampleImg, snapshots, wasteSummary, secReport, advReport)

	// Test Initial State
	if m.selectedLayer != 0 {
		t.Errorf("expected initial layer 0, got %d", m.selectedLayer)
	}

	// Send Down key
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.selectedLayer != 1 {
		t.Errorf("expected layer 1 after Down, got %d", m.selectedLayer)
	}

	// Send Tab key to switch pane
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.pane != paneTree {
		t.Errorf("expected paneTree after Tab, got %v", m.pane)
	}

	// Send 'w' to toggle wasted
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = updated.(Model)
	if !m.wastedOnly {
		t.Errorf("expected wastedOnly to be true after pressing 'w'")
	}

	// Test View rendering does not crash
	viewStr := m.View()
	if len(viewStr) == 0 {
		t.Errorf("expected non-empty view string")
	}
}
