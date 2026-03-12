package infrastructure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"

	knowledgeapp "github.com/alty-cli/alty/internal/knowledge/application"
	"github.com/alty-cli/alty/internal/knowledge/domain"
)

// Compile-time interface check.
var _ knowledgeapp.DriftDetection = (*DriftDetectionAdapter)(nil)

// DefaultStaleThresholdDays is the default number of days after which a knowledge entry
// is considered stale. Set to 14 days (2 weeks) because AI tools like Claude Code and
// Cursor typically release updates weekly, and knowledge should be verified frequently
// to catch convention changes.
const DefaultStaleThresholdDays = 14

// DriftDetectionAdapter implements DriftDetection by scanning the knowledge base
// for staleness and other drift signals.
type DriftDetectionAdapter struct {
	projectDir         string
	staleThresholdDays int
}

// NewDriftDetectionAdapter creates a new DriftDetectionAdapter with default settings.
func NewDriftDetectionAdapter(projectDir string) *DriftDetectionAdapter {
	return &DriftDetectionAdapter{
		projectDir:         projectDir,
		staleThresholdDays: DefaultStaleThresholdDays,
	}
}

// WithStaleThreshold sets a custom staleness threshold in days.
func (a *DriftDetectionAdapter) WithStaleThreshold(days int) *DriftDetectionAdapter {
	a.staleThresholdDays = days
	return a
}

// Detect scans the knowledge base and returns a drift report.
func (a *DriftDetectionAdapter) Detect(ctx context.Context) (domain.DriftReport, error) {
	select {
	case <-ctx.Done():
		return domain.DriftReport{}, fmt.Errorf("detecting drift: %w", ctx.Err())
	default:
	}

	var signals []domain.DriftSignal

	toolsDir := filepath.Join(a.projectDir, ".alty", "knowledge", "tools")
	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.NewDriftReport(nil), nil
		}
		return domain.DriftReport{}, fmt.Errorf("reading tools directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		toolName := entry.Name()
		metaPath := filepath.Join(toolsDir, toolName, "_meta.toml")

		staleSignals := a.checkStaleness(toolName, metaPath)
		signals = append(signals, staleSignals...)
	}

	return domain.NewDriftReport(signals), nil
}

// toolMeta represents the structure of a _meta.toml file.
type toolMeta struct {
	Tool     toolInfo                   `toml:"tool"`
	Versions map[string]versionMetadata `toml:"versions"`
}

type toolInfo struct {
	Name string `toml:"name"`
}

type versionMetadata struct {
	LastVerified string `toml:"last_verified"`
}

// checkStaleness checks if a tool's knowledge entry is stale.
func (a *DriftDetectionAdapter) checkStaleness(toolName, metaPath string) []domain.DriftSignal {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil // No meta file, skip
	}

	var meta toolMeta
	if err := toml.Unmarshal(data, &meta); err != nil {
		return nil // Malformed TOML, skip gracefully
	}

	var signals []domain.DriftSignal
	threshold := time.Now().AddDate(0, 0, -a.staleThresholdDays)

	// Check each version's last_verified date
	for versionKey, version := range meta.Versions {
		// Skip special keys that aren't version entries
		if versionKey == "current" || versionKey == "tracked" || versionKey == "deprecated" {
			continue
		}

		isStale := false
		var reason string

		if version.LastVerified == "" {
			isStale = true
			reason = "no last_verified date"
		} else {
			lastVerified, err := time.Parse("2006-01-02", version.LastVerified)
			if err != nil {
				isStale = true
				reason = "invalid last_verified date format"
			} else if lastVerified.Before(threshold) {
				isStale = true
				reason = fmt.Sprintf("last verified %s (>%d days ago)", version.LastVerified, a.staleThresholdDays)
			}
		}

		if isStale {
			entryPath := fmt.Sprintf("tools/%s/%s", toolName, versionKey)
			signal, err := domain.NewDriftSignal(
				entryPath,
				domain.DriftStale,
				fmt.Sprintf("Knowledge entry stale: %s", reason),
				domain.SeverityInfo,
			)
			if err == nil {
				signals = append(signals, signal)
			}
		}
	}

	return signals
}
