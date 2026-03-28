package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	sharedapp "github.com/alto-cli/alto/internal/shared/application"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// DiscoveryReportHandler generates a discovery report from domain stories,
// glossary, and context map artifacts.
type DiscoveryReportHandler struct {
	storyReader      StoryReader
	glossaryReader   GlossaryReader
	contextMapReader ContextMapReader
	writer           sharedapp.FileWriter
}

// NewDiscoveryReportHandler creates a DiscoveryReportHandler with the given ports.
func NewDiscoveryReportHandler(
	storyReader StoryReader,
	glossaryReader GlossaryReader,
	contextMapReader ContextMapReader,
	writer sharedapp.FileWriter,
) *DiscoveryReportHandler {
	return &DiscoveryReportHandler{
		storyReader:      storyReader,
		glossaryReader:   glossaryReader,
		contextMapReader: contextMapReader,
		writer:           writer,
	}
}

// GenerateReport reads stories, glossary, and context map, then writes a
// markdown discovery report to the project directory.
func (h *DiscoveryReportHandler) GenerateReport(
	ctx context.Context,
	storyPaths []string,
	glossaryPath string,
	contextMapPath string,
	projectDir string,
) error {
	// Read stories.
	var stories []*discoverydomain.DomainStory
	for _, path := range storyPaths {
		story, err := h.storyReader.Read(ctx, path)
		if err != nil {
			return fmt.Errorf("reading story %q: %w", path, err)
		}

		stories = append(stories, story)
	}

	// Read glossary.
	entries, err := h.glossaryReader.Read(ctx, glossaryPath)
	if err != nil {
		return fmt.Errorf("reading glossary: %w", err)
	}

	// Read context map.
	cm, err := h.contextMapReader.Read(ctx, contextMapPath)
	if err != nil {
		return fmt.Errorf("reading context map: %w", err)
	}

	// Build summaries and distribution.
	summaries := make([]discoverydomain.StorySummary, 0, len(stories))
	dist := discoverydomain.NewTrustDistribution()

	for _, story := range stories {
		summaries = append(summaries, discoverydomain.SummarizeStory(story))
		dist = dist.AddStory(story)
	}

	sketches := cm.Contexts()
	for _, sketch := range sketches {
		dist = dist.AddSketch(sketch)
	}

	reportPath := filepath.Join(projectDir, ".alto", "discovery-report.md")

	data := reportData{
		summaries:      summaries,
		sketches:       sketches,
		glossaryCount:  len(entries),
		distribution:   dist,
		storyPaths:     storyPaths,
		glossaryPath:   glossaryPath,
		contextMapPath: contextMapPath,
		reportPath:     reportPath,
	}

	content := renderReport(data)

	if err := h.writer.WriteFile(ctx, reportPath, content); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}

	return nil
}

type reportData struct {
	summaries      []discoverydomain.StorySummary
	sketches       []discoverydomain.BoundedContextSketch
	glossaryCount  int
	distribution   discoverydomain.TrustDistribution
	storyPaths     []string
	glossaryPath   string
	contextMapPath string
	reportPath     string
}

func renderReport(data reportData) string {
	var b strings.Builder

	b.WriteString("# Discovery Report\n")
	fmt.Fprintf(&b, "Generated: %s\n", time.Now().Format(time.RFC3339))

	// Stories table.
	b.WriteString("\n## Stories\n")
	b.WriteString("| Story | Actors | Sentences |")

	for _, level := range vo.AllTrustLevels() {
		fmt.Fprintf(&b, " %s |", level)
	}

	b.WriteString("\n")
	b.WriteString("|-------|--------|-----------|")

	for range vo.AllTrustLevels() {
		b.WriteString("------|")
	}

	b.WriteString("\n")

	for _, s := range data.summaries {
		fmt.Fprintf(&b, "| %s | %d | %d |", s.Title, s.ActorCount, s.SentenceCount)

		for _, level := range vo.AllTrustLevels() {
			fmt.Fprintf(&b, " %d |", s.Distribution.Count(level))
		}

		b.WriteString("\n")
	}

	// Boundary Decisions table.
	b.WriteString("\n## Boundary Decisions\n")
	b.WriteString("| Bounded Context | Classification | Confidence |\n")
	b.WriteString("|-----------------|----------------|------------|\n")

	for _, sketch := range data.sketches {
		fmt.Fprintf(&b, "| %s | %s | %s |\n", sketch.Name(), sketch.Classification(), sketch.ConfidenceLevel())
	}

	// Trust Distribution table.
	b.WriteString("\n## Trust Distribution (all elements)\n")
	b.WriteString("| Trust Level | Count |\n")
	b.WriteString("|-------------|-------|\n")

	for _, level := range vo.AllTrustLevels() {
		fmt.Fprintf(&b, "| %s | %d |\n", level, data.distribution.Count(level))
	}

	fmt.Fprintf(&b, "| **Total** | **%d** |\n", data.distribution.Total())

	// Glossary.
	b.WriteString("\n## Glossary\n")
	fmt.Fprintf(&b, "Terms defined: %d\n", data.glossaryCount)

	// Artifacts.
	b.WriteString("\n## Artifacts\n")

	for _, p := range data.storyPaths {
		fmt.Fprintf(&b, "- %s\n", p)
	}

	fmt.Fprintf(&b, "- %s\n", data.glossaryPath)
	fmt.Fprintf(&b, "- %s\n", data.contextMapPath)
	fmt.Fprintf(&b, "- %s\n", data.reportPath)

	return b.String()
}
