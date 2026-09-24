package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/alexandrmotologa/layerscope/pkg/advisor"
	"github.com/alexandrmotologa/layerscope/pkg/audit"
	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/sbom"
	"github.com/alexandrmotologa/layerscope/pkg/security"
	"github.com/alexandrmotologa/layerscope/pkg/server"
	"github.com/alexandrmotologa/layerscope/pkg/slim"
	"github.com/alexandrmotologa/layerscope/pkg/tui"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

var (
	version = "1.0.0"
	commit  = "main"
	date    = "2026-09-24"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "layerscope",
		Short: "Next-Gen OCI & Docker Container Layer Inspector, Multi-Arch Diffing & Secret Auditor",
		Long: `LayerScope is a single-binary container image analysis studio and security auditor.
It inspects OCI layers, computes wasted space, detects intermediate layer secret leaks,
diffs images or architectures side-by-side, and generates standard CycloneDX/SPDX SBOMs.`,
	}

	rootCmd.AddCommand(newAnalyzeCmd())
	rootCmd.AddCommand(newTUICmd())
	rootCmd.AddCommand(newSlimCmd())
	rootCmd.AddCommand(newDiffCmd())
	rootCmd.AddCommand(newAuditCmd())
	rootCmd.AddCommand(newSBOMCmd())
	rootCmd.AddCommand(newVersionCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newAnalyzeCmd() *cobra.Command {
	var (
		port        int
		forceRemote bool
		demo        bool
		arch        string
		noBrowser   bool
	)

	cmd := &cobra.Command{
		Use:   "analyze [image]",
		Short: "Analyze a container image and launch the visual web studio",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "demo"
			if len(args) > 0 {
				target = args[0]
			} else if !demo {
				target = "demo"
				demo = true
			}

			fmt.Printf("Analyzing image %q...\n", target)

			scanner := security.NewScanner()
			extractor := sbom.NewExtractor()

			var commands []string
			hook := func(layerIndex int, file *oci.LayerFile, reader io.Reader) error {
				var buf []byte
				cmdStr := ""
				if layerIndex < len(commands) {
					cmdStr = commands[layerIndex]
				}

				// If manifest, buffer to pass to extractor
				if sbom.IsManifestPath(file.Path) {
					data, err := io.ReadAll(reader)
					if err != nil {
						return err
					}
					buf = data
					_ = extractor.Hook()(layerIndex, file, strings.NewReader(string(buf)))
					_, _ = scanner.ScanStream(layerIndex, cmdStr, file.Path, strings.NewReader(string(buf)))
					return nil
				}

				_, err := scanner.ScanStream(layerIndex, cmdStr, file.Path, reader)
				return err
			}

			img, err := oci.LoadImage(target, oci.IngestionOptions{
				ForceRemote:  forceRemote,
				ForceDemo:    demo,
				Architecture: arch,
			}, hook)
			if err != nil {
				return fmt.Errorf("load image: %w", err)
			}

			snapshots := vfs.BuildLayerSnapshots(img)
			wasteSummary := vfs.CalculateWastedSpace(snapshots, nil)
			finalTree := snapshots[len(snapshots)-1].Tree
			runsAsRoot, suid := security.AuditPrivileges(&img.Config, finalTree)
			secReport := scanner.FinalizeAudit(finalTree, runsAsRoot, suid)
			sbomReport, _ := extractor.FinalizeReport(img.Reference.Original, finalTree)

			// Query OSV.dev CVE database for SBOM components (5s timeout)
			enrichCtx, enrichCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer enrichCancel()
			extractor.EnrichWithVulnerabilities(enrichCtx, sbomReport)

			advReport := advisor.AnalyzeImageHeuristics(img, wasteSummary, secReport, finalTree)

			staticFS := server.GetEmbeddedUIFileSystem()
			srv := server.NewServer(img, snapshots, wasteSummary, secReport, sbomReport, advReport, staticFS)

			addr := fmt.Sprintf(":%d", port)
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				// Try random port if requested port is occupied
				listener, err = net.Listen("tcp", ":0")
				if err != nil {
					return fmt.Errorf("listen: %w", err)
				}
				port = listener.Addr().(*net.TCPAddr).Port
			}

			studioURL := fmt.Sprintf("http://localhost:%d", port)
			fmt.Printf("\n LayerScope Studio running at: %s\n", studioURL)
			fmt.Printf(" Efficiency Score: %.1f%% (Grade: %s)\n", advReport.EfficiencyScore, advReport.Grade)
			fmt.Printf(" Total Layers: %d | Wasted Bytes: %.2f MB\n", len(snapshots), float64(wasteSummary.TotalWastedBytes)/(1024*1024))
			fmt.Printf(" Credentials Detected: %d (Intermediate Leaks: %d)\n\n", secReport.TotalFindings, secReport.IntermediateLeaks)

			if !noBrowser {
				openBrowser(studioURL)
			}

			return http.Serve(listener, srv.Router())
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 50060, "HTTP server port")
	cmd.Flags().BoolVar(&forceRemote, "remote", false, "Fetch directly from remote registry without local Docker daemon")
	cmd.Flags().BoolVar(&demo, "demo", false, "Run with sample multi-layer demo image")
	cmd.Flags().StringVar(&arch, "arch", "amd64", "Target architecture (amd64, arm64)")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Do not open default web browser automatically")

	return cmd
}

func newTUICmd() *cobra.Command {
	var (
		forceRemote bool
		demo        bool
		arch        string
		wastedOnly  bool
	)

	cmd := &cobra.Command{
		Use:   "tui [image]",
		Short: "Inspect an image in an interactive terminal user interface",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "demo"
			if len(args) > 0 {
				target = args[0]
			} else if !demo {
				target = "demo"
				demo = true
			}

			scanner := security.NewScanner()
			img, err := oci.LoadImage(target, oci.IngestionOptions{
				ForceRemote:  forceRemote,
				ForceDemo:    demo,
				Architecture: arch,
			}, scanner.Hook(nil))
			if err != nil {
				return fmt.Errorf("load image: %w", err)
			}

			snapshots := vfs.BuildLayerSnapshots(img)
			wasteSummary := vfs.CalculateWastedSpace(snapshots, nil)
			finalTree := snapshots[len(snapshots)-1].Tree
			runsAsRoot, suid := security.AuditPrivileges(&img.Config, finalTree)
			secReport := scanner.FinalizeAudit(finalTree, runsAsRoot, suid)
			advReport := advisor.AnalyzeImageHeuristics(img, wasteSummary, secReport, finalTree)

			model := tui.NewModel(img, snapshots, wasteSummary, secReport, advReport)
			if wastedOnly {
				// Simulate pressing 'w'
				updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
				model = updated.(tui.Model)
			}

			p := tea.NewProgram(model, tea.WithAltScreen())
			_, err = p.Run()
			return err
		},
	}

	cmd.Flags().BoolVar(&forceRemote, "remote", false, "Fetch directly from remote registry")
	cmd.Flags().BoolVar(&demo, "demo", false, "Run with sample demo image")
	cmd.Flags().StringVar(&arch, "arch", "amd64", "Target architecture")
	cmd.Flags().BoolVar(&wastedOnly, "wasted-only", false, "Start with only wasted files filter enabled")

	return cmd
}

func newDiffCmd() *cobra.Command {
	var (
		archA  string
		archB  string
		format string
	)

	cmd := &cobra.Command{
		Use:   "diff [imageA] [imageB]",
		Short: "Compare two container image tags or architecture variants side by side",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			imgRefA := args[0]
			imgRefB := args[1]

			imgA, err := oci.LoadImage(imgRefA, oci.IngestionOptions{Architecture: archA}, nil)
			if err != nil {
				return fmt.Errorf("load image A (%s): %w", imgRefA, err)
			}
			snapsA := vfs.BuildLayerSnapshots(imgA)
			treeA := snapsA[len(snapsA)-1].Tree

			imgB, err := oci.LoadImage(imgRefB, oci.IngestionOptions{Architecture: archB}, nil)
			if err != nil {
				return fmt.Errorf("load image B (%s): %w", imgRefB, err)
			}
			snapsB := vfs.BuildLayerSnapshots(imgB)
			treeB := snapsB[len(snapsB)-1].Tree

			diffReport := vfs.DiffVFSTrees(imgRefA, treeA, imgA.TotalSizeBytes, imgRefB, treeB, imgB.TotalSizeBytes)

			if format == "json" {
				return printJSON(diffReport)
			}

			fmt.Printf("LayerScope Diff: %s vs %s\n", imgRefA, imgRefB)
			fmt.Printf("Total Size Delta: %+d bytes (%.2f MB)\n", diffReport.TotalSizeDelta, float64(diffReport.TotalSizeDelta)/(1024*1024))
			fmt.Printf("Added Files: %d | Removed Files: %d | Modified Files: %d | Identical Files: %d\n\n",
				diffReport.AddedCount, diffReport.RemovedCount, diffReport.ModifiedCount, diffReport.SameCount)

			fmt.Println("Top Differences:")
			limit := 25
			if len(diffReport.Files) < limit {
				limit = len(diffReport.Files)
			}
			for i := 0; i < limit; i++ {
				f := diffReport.Files[i]
				fmt.Printf("  [%-8s] %-50s (%+d KB)\n", f.Status, f.Path, f.SizeDelta/1024)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&archA, "arch-a", "amd64", "Architecture for Image A")
	cmd.Flags().StringVar(&archB, "arch-b", "amd64", "Architecture for Image B")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text, json")

	return cmd
}

func newAuditCmd() *cobra.Command {
	var (
		minEfficiency float64
		failOnSecrets bool
		sbomOut       string
		format        string
		outFile       string
		forceRemote   bool
		demo          bool
		arch          string
	)

	cmd := &cobra.Command{
		Use:   "audit [image]",
		Short: "Headless audit for CI pipelines with pass/fail exit code",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "demo"
			if len(args) > 0 {
				target = args[0]
			} else if !demo {
				target = "demo"
				demo = true
			}

			scanner := security.NewScanner()
			extractor := sbom.NewExtractor()

			hook := func(layerIndex int, file *oci.LayerFile, reader io.Reader) error {
				if sbom.IsManifestPath(file.Path) {
					data, err := io.ReadAll(reader)
					if err != nil {
						return err
					}
					_ = extractor.Hook()(layerIndex, file, strings.NewReader(string(data)))
					_, _ = scanner.ScanStream(layerIndex, "", file.Path, strings.NewReader(string(data)))
					return nil
				}
				_, err := scanner.ScanStream(layerIndex, "", file.Path, reader)
				return err
			}

			img, err := oci.LoadImage(target, oci.IngestionOptions{
				ForceRemote:  forceRemote,
				ForceDemo:    demo,
				Architecture: arch,
			}, hook)
			if err != nil {
				return fmt.Errorf("load image: %w", err)
			}

			snapshots := vfs.BuildLayerSnapshots(img)
			wasteSummary := vfs.CalculateWastedSpace(snapshots, nil)
			finalTree := snapshots[len(snapshots)-1].Tree
			runsAsRoot, suid := security.AuditPrivileges(&img.Config, finalTree)
			secReport := scanner.FinalizeAudit(finalTree, runsAsRoot, suid)
			sbomReport, _ := extractor.FinalizeReport(img.Reference.Original, finalTree)

			auditResult := audit.GenerateFullAudit(img, snapshots, wasteSummary, secReport, sbomReport)

			// Export SBOM if requested
			if sbomOut != "" && sbomReport != nil {
				cdxBytes, err := sbom.ExportCycloneDXJSON(sbomReport)
				if err == nil {
					_ = os.WriteFile(sbomOut, cdxBytes, 0644)
				}
			}

			var outputContent []byte
			switch strings.ToLower(format) {
			case "json":
				outputContent, _ = auditResult.FormatJSON()
			case "markdown":
				outputContent = []byte(auditResult.FormatMarkdown())
			case "junit":
				outputContent, _ = auditResult.FormatJUnit(minEfficiency, failOnSecrets)
			case "html":
				outputContent = []byte(auditResult.FormatHTML())
			default:
				// Plain text output
				outputContent = []byte(auditResult.FormatMarkdown())
			}

			if outFile != "" {
				if err := os.WriteFile(outFile, outputContent, 0644); err != nil {
					return fmt.Errorf("write output file: %w", err)
				}
			} else {
				fmt.Println(string(outputContent))
			}

			// CI exit criteria
			failed := false
			if auditResult.Advisor.EfficiencyScore < minEfficiency {
				fmt.Fprintf(os.Stderr, "FAILURE: Efficiency %.1f%% is below threshold %.1f%%\n",
					auditResult.Advisor.EfficiencyScore, minEfficiency)
				failed = true
			}
			if failOnSecrets && auditResult.Security.TotalFindings > 0 {
				fmt.Fprintf(os.Stderr, "FAILURE: Found %d credential leaks in image layers\n",
					auditResult.Security.TotalFindings)
				failed = true
			}

			if failed {
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().Float64Var(&minEfficiency, "min-efficiency", 85.0, "Minimum required image efficiency score (0-100)")
	cmd.Flags().BoolVar(&failOnSecrets, "fail-on-secrets", true, "Fail audit if credentials are found in layers")
	cmd.Flags().StringVar(&sbomOut, "sbom-out", "", "File path to save CycloneDX SBOM")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text, json, markdown, junit, html")
	cmd.Flags().StringVar(&outFile, "out", "", "File path to write report output")
	cmd.Flags().BoolVar(&forceRemote, "remote", false, "Fetch directly from remote registry")
	cmd.Flags().BoolVar(&demo, "demo", false, "Run audit on demo image")
	cmd.Flags().StringVar(&arch, "arch", "amd64", "Target architecture")

	return cmd
}

func newSBOMCmd() *cobra.Command {
	var (
		format      string
		output      string
		forceRemote bool
		demo        bool
		arch        string
	)

	cmd := &cobra.Command{
		Use:   "sbom [image]",
		Short: "Extract software dependencies and export CycloneDX or SPDX SBOM",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "demo"
			if len(args) > 0 {
				target = args[0]
			} else if !demo {
				target = "demo"
				demo = true
			}

			extractor := sbom.NewExtractor()
			img, err := oci.LoadImage(target, oci.IngestionOptions{
				ForceRemote:  forceRemote,
				ForceDemo:    demo,
				Architecture: arch,
			}, extractor.Hook())
			if err != nil {
				return fmt.Errorf("load image: %w", err)
			}

			snapshots := vfs.BuildLayerSnapshots(img)
			finalTree := snapshots[len(snapshots)-1].Tree
			report, err := extractor.FinalizeReport(img.Reference.Original, finalTree)
			if err != nil {
				return fmt.Errorf("extract sbom: %w", err)
			}

			var outputBytes []byte
			switch strings.ToLower(format) {
			case "cyclonedx", "cdx":
				outputBytes, err = sbom.ExportCycloneDXJSON(report)
			case "spdx":
				outputBytes, err = sbom.ExportSPDXJSON(report)
			default:
				outputBytes, err = printJSONBytes(report)
			}

			if err != nil {
				return err
			}

			if output != "" {
				return os.WriteFile(output, outputBytes, 0644)
			}

			fmt.Println(string(outputBytes))
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "cyclonedx", "SBOM format: cyclonedx, spdx, json")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default stdout)")
	cmd.Flags().BoolVar(&forceRemote, "remote", false, "Fetch directly from remote registry")
	cmd.Flags().BoolVar(&demo, "demo", false, "Extract SBOM from demo image")
	cmd.Flags().StringVar(&arch, "arch", "amd64", "Target architecture")

	return cmd
}

func newSlimCmd() *cobra.Command {
	var (
		output      string
		tag         string
		purgeCaches bool
		forceRemote bool
		demo        bool
		arch        string
	)

	cmd := &cobra.Command{
		Use:   "slim [image]",
		Short: "Squash container image into a clean single layer and purge intermediate wasted files",
		Long: `Squashes multi-layer images into a minimal, clean single layer Docker/OCI tarball.
Purges whiteout-deleted files, temporary build caches, and intermediate leaked credentials.
The resulting archive can be loaded directly into Docker with 'docker load -i <output.tar>'.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "demo"
			if len(args) > 0 {
				target = args[0]
			} else if !demo {
				target = "demo"
				demo = true
			}

			if output == "" {
				cleanName := strings.ReplaceAll(strings.ReplaceAll(target, "/", "-"), ":", "-")
				output = fmt.Sprintf("slim-%s.tar", cleanName)
				if output == "slim-demo.tar" {
					output = "slim-image.tar"
				}
			}

			fmt.Printf("Analyzing and squashing image %q...\n", target)

			img, err := oci.LoadImage(target, oci.IngestionOptions{
				ForceRemote:  forceRemote,
				ForceDemo:    demo,
				Architecture: arch,
			}, nil)
			if err != nil {
				return fmt.Errorf("load image: %w", err)
			}

			snapshots := vfs.BuildLayerSnapshots(img)
			finalTree := snapshots[len(snapshots)-1].Tree

			result, err := slim.ExportSquashedImage(img, finalTree, output, slim.Options{
				Tag:         tag,
				PurgeCaches: purgeCaches,
			})
			if err != nil {
				return fmt.Errorf("slim image export: %w", err)
			}

			fmt.Printf("\n✓ Squashed container image exported successfully!\n")
			fmt.Printf("  Original Size:   %.2f MB (%d layers)\n", float64(result.OriginalSize)/(1024*1024), len(snapshots))
			fmt.Printf("  Squashed Size:   %.2f MB (1 layer)\n", float64(result.SlimSize)/(1024*1024))
			fmt.Printf("  Saved Space:     %.2f MB (%.1f%% reduction)\n", float64(result.SavedBytes)/(1024*1024), result.SavedPercentage)
			fmt.Printf("  Output Archive:  %s\n\n", result.OutputFile)
			fmt.Printf("To import this squashed image into Docker:\n")
			fmt.Printf("  docker load -i %s\n", result.OutputFile)

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output tar archive path (default: slim-<image>.tar)")
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "Image repo:tag in the output archive")
	cmd.Flags().BoolVar(&purgeCaches, "purge-caches", true, "Purge package manager caches (/var/cache/*, /tmp/*, /var/lib/apt/lists/*)")
	cmd.Flags().BoolVar(&forceRemote, "remote", false, "Fetch directly from remote registry")
	cmd.Flags().BoolVar(&demo, "demo", false, "Use demo image fixture")
	cmd.Flags().StringVar(&arch, "arch", "amd64", "Target architecture")

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display LayerScope version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("LayerScope v%s (%s, build date: %s, %s/%s)\n",
				version, commit, date, runtime.GOOS, runtime.GOARCH)
		},
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func printJSON(data interface{}) error {
	bytes, err := printJSONBytes(data)
	if err != nil {
		return err
	}
	fmt.Println(string(bytes))
	return nil
}

func printJSONBytes(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}
