package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexandrmotologa/layerscope/pkg/advisor"
	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/security"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

type activePane int

const (
	paneLayers activePane = iota
	paneTree
)

// Model represents the Bubbletea terminal state.
type Model struct {
	img             *oci.ImageAnalysis
	snapshots       []*vfs.LayerSnapshot
	waste           *vfs.WastedSpaceSummary
	secReport       *security.SecurityAuditReport
	advisorReport   *advisor.AdvisorReport
	pane            activePane
	selectedLayer   int
	treeCursor      int
	wastedOnly      bool
	width           int
	height          int
	quitting        bool
	flatTreeCache   map[int][]*vfs.FileNode
	expandedDirs    map[string]bool
	searchQuery     string
	searching       bool
}

// NewModel creates an interactive TUI model.
func NewModel(
	img *oci.ImageAnalysis,
	snapshots []*vfs.LayerSnapshot,
	waste *vfs.WastedSpaceSummary,
	secReport *security.SecurityAuditReport,
	adv *advisor.AdvisorReport,
) Model {
	expanded := make(map[string]bool)
	expanded["/"] = true

	return Model{
		img:           img,
		snapshots:     snapshots,
		waste:         waste,
		secReport:     secReport,
		advisorReport: adv,
		pane:          paneLayers,
		selectedLayer: 0,
		treeCursor:    0,
		flatTreeCache: make(map[int][]*vfs.FileNode),
		expandedDirs:  expanded,
		width:         120,
		height:        35,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "tab":
			if m.pane == paneLayers {
				m.pane = paneTree
			} else {
				m.pane = paneLayers
			}
			return m, nil

		case "w", "W":
			m.wastedOnly = !m.wastedOnly
			m.treeCursor = 0
			return m, nil

		case "up", "k":
			if m.pane == paneLayers {
				if m.selectedLayer > 0 {
					m.selectedLayer--
					m.treeCursor = 0
				}
			} else {
				if m.treeCursor > 0 {
					m.treeCursor--
				}
			}
			return m, nil

		case "down", "j":
			if m.pane == paneLayers {
				if m.selectedLayer < len(m.snapshots)-1 {
					m.selectedLayer++
					m.treeCursor = 0
				}
			} else {
				visibleFiles := m.getVisibleFiles()
				if m.treeCursor < len(visibleFiles)-1 {
					m.treeCursor++
				}
			}
			return m, nil

		case "enter":
			if m.pane == paneTree {
				files := m.getVisibleFiles()
				if m.treeCursor < len(files) {
					node := files[m.treeCursor]
					if node.IsDir {
						m.expandedDirs[node.Path] = !m.expandedDirs[node.Path]
					}
				}
			}
			return m, nil
		}
	}

	return m, nil
}

func (m Model) getVisibleFiles() []*vfs.FileNode {
	if m.selectedLayer >= len(m.snapshots) {
		return nil
	}

	snap := m.snapshots[m.selectedLayer]
	var result []*vfs.FileNode

	var walk func(n *vfs.FileNode, depth int)
	walk = func(n *vfs.FileNode, depth int) {
		if n.Path != "/" {
			if m.wastedOnly {
				if n.IsWasted {
					result = append(result, n)
				}
			} else {
				result = append(result, n)
			}
		}

		if n.IsDir && (n.Path == "/" || m.expandedDirs[n.Path]) {
			for _, child := range n.SortedChildren() {
				walk(child, depth+1)
			}
		}
	}

	walk(snap.Tree.Root, 0)
	return result
}

// View renders the dual-pane terminal layout.
func (m Model) View() string {
	if m.quitting {
		return "Exiting LayerScope...\n"
	}

	// Lipgloss styles
	borderColor := lipgloss.Color("240")
	activeBorderColor := lipgloss.Color("39") // vibrant blue

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	boxWidth := (m.width - 6) / 2
	if boxWidth < 30 {
		boxWidth = 30
	}
	contentHeight := m.height - 7
	if contentHeight < 10 {
		contentHeight = 10
	}

	layerBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(boxWidth).
		Height(contentHeight)

	treeBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(boxWidth).
		Height(contentHeight)

	if m.pane == paneLayers {
		layerBoxStyle = layerBoxStyle.BorderForeground(activeBorderColor)
	} else {
		treeBoxStyle = treeBoxStyle.BorderForeground(activeBorderColor)
	}

	// Render Left: Layers List
	var leftSb strings.Builder
	leftSb.WriteString(headerStyle.Render("LAYERS (Tab to switch)"))
	leftSb.WriteString("\n")

	for i, snap := range m.snapshots {
		prefix := "  "
		itemStyle := lipgloss.NewStyle()
		if i == m.selectedLayer {
			prefix = "> "
			itemStyle = itemStyle.Bold(true).Foreground(lipgloss.Color("39"))
		}

		cmdSnippet := snap.Command
		if len(cmdSnippet) > boxWidth-18 {
			cmdSnippet = cmdSnippet[:boxWidth-21] + "..."
		}
		if cmdSnippet == "" {
			cmdSnippet = fmt.Sprintf("Layer %d", i)
		}

		sizeMB := float64(snap.Size) / (1024 * 1024)
		wasteTag := ""
		if snap.WastedBytes > 0 {
			wasteTag = " [Wasted]"
		}

		line := fmt.Sprintf("%s[%d] %-30s %.1fMB%s\n", prefix, i, cmdSnippet, sizeMB, wasteTag)
		leftSb.WriteString(itemStyle.Render(line))
	}

	// Render Right: Virtual File System Tree
	var rightSb strings.Builder
	modeTitle := "ALL FILES"
	if m.wastedOnly {
		modeTitle = "ONLY WASTED FILES (W key to toggle)"
	}
	rightSb.WriteString(headerStyle.Render("FILE TREE: " + modeTitle))
	rightSb.WriteString("\n")

	visibleFiles := m.getVisibleFiles()
	startIdx := 0
	if m.treeCursor >= contentHeight-2 {
		startIdx = m.treeCursor - (contentHeight - 3)
	}
	endIdx := startIdx + contentHeight - 2
	if endIdx > len(visibleFiles) {
		endIdx = len(visibleFiles)
	}

	for idx := startIdx; idx < endIdx; idx++ {
		node := visibleFiles[idx]
		cursorPrefix := "  "
		lineStyle := lipgloss.NewStyle()

		if idx == m.treeCursor && m.pane == paneTree {
			cursorPrefix = "> "
			lineStyle = lineStyle.Bold(true).Foreground(lipgloss.Color("39"))
		}

		changeIndicator := " "
		switch node.ChangeType {
		case vfs.ChangeAdded:
			changeIndicator = "+"
			lineStyle = lineStyle.Foreground(lipgloss.Color("42")) // Green
		case vfs.ChangeModified:
			changeIndicator = "~"
			lineStyle = lineStyle.Foreground(lipgloss.Color("220")) // Yellow
		case vfs.ChangeDeleted:
			changeIndicator = "-"
			lineStyle = lineStyle.Foreground(lipgloss.Color("196")) // Red
		}

		icon := " "
		if node.IsDir {
			if m.expandedDirs[node.Path] {
				icon = "v "
			} else {
				icon = "> "
			}
		}

		displayName := node.Path
		if len(displayName) > boxWidth-16 {
			displayName = "..." + displayName[len(displayName)-(boxWidth-19):]
		}

		sizeStr := ""
		if !node.IsDir {
			sizeStr = fmt.Sprintf(" %dB", node.Size)
			if node.Size > 1024*1024 {
				sizeStr = fmt.Sprintf(" %.1fMB", float64(node.Size)/(1024*1024))
			}
		}

		row := fmt.Sprintf("%s%s %s%-35s%s\n", cursorPrefix, changeIndicator, icon, displayName, sizeStr)
		rightSb.WriteString(lineStyle.Render(row))
	}

	// Assemble panes side-by-side
	panes := lipgloss.JoinHorizontal(lipgloss.Top,
		layerBoxStyle.Render(leftSb.String()),
		treeBoxStyle.Render(rightSb.String()),
	)

	// Top Title and Bottom Status Bar
	topBar := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("63")).
		Padding(0, 1).
		Render(fmt.Sprintf(" LayerScope v1.0.0 | Image: %s | Efficiency: %.1f%% (%s) ",
			m.img.Reference.Original, m.advisorReport.EfficiencyScore, m.advisorReport.Grade))

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render(" [Tab] Switch Pane | [↑/↓] Navigate | [Enter] Expand/Collapse | [W] Toggle Wasted | [Q] Quit")

	return fmt.Sprintf("%s\n\n%s\n%s\n", topBar, panes, footer)
}
