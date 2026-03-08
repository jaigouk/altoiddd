package infrastructure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	knowledgedomain "github.com/alty-cli/alty/internal/knowledge/domain"
)

// KnowledgeDriftDetector scans knowledge entries for drift signals.
// Implements DriftDetection port.
type KnowledgeDriftDetector struct {
	root string
}

// NewKnowledgeDriftDetector creates a KnowledgeDriftDetector.
func NewKnowledgeDriftDetector(knowledgeDir string) *KnowledgeDriftDetector {
	return &KnowledgeDriftDetector{root: knowledgeDir}
}

// Detect detects drift across all tool knowledge entries.
func (d *KnowledgeDriftDetector) Detect(_ context.Context) (knowledgedomain.DriftReport, error) {
	var signals []knowledgedomain.DriftSignal
	toolsDir := filepath.Join(d.root, "tools")

	if _, err := os.Stat(toolsDir); os.IsNotExist(err) {
		return knowledgedomain.NewDriftReport(nil), nil
	}

	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		return knowledgedomain.NewDriftReport(nil), nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		toolDir := filepath.Join(toolsDir, entry.Name())
		signals = append(signals, d.scanTool(toolDir)...)
	}

	return knowledgedomain.NewDriftReport(signals), nil
}

func (d *KnowledgeDriftDetector) scanTool(toolDir string) []knowledgedomain.DriftSignal {
	var signals []knowledgedomain.DriftSignal
	metaPath := filepath.Join(toolDir, "_meta.toml")

	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return signals
	}

	meta := readTOMLFile(metaPath)
	versionsSection, ok := meta["versions"].(map[string]any)
	if !ok {
		return signals
	}

	tracked := toStringSlice(versionsSection["tracked"])

	currentDir := filepath.Join(toolDir, "current")
	if _, err := os.Stat(currentDir); os.IsNotExist(err) {
		return signals
	}

	toolName := filepath.Base(toolDir)

	tomlFiles, _ := filepath.Glob(filepath.Join(currentDir, "*.toml"))
	sort.Strings(tomlFiles)

	for _, entryPath := range tomlFiles {
		entryName := strings.TrimSuffix(filepath.Base(entryPath), ".toml")
		rlmPath := fmt.Sprintf("tools/%s/%s", toolName, entryName)
		currentData := readTOMLFile(entryPath)

		// Staleness check
		signals = append(signals, d.checkStaleness(rlmPath, currentData)...)

		// Version comparison
		if len(tracked) == 0 {
			sig, err := knowledgedomain.NewDriftSignal(
				rlmPath,
				knowledgedomain.DriftVersionChange,
				fmt.Sprintf("No version history for %s -- cannot compare for drift", toolName),
				knowledgedomain.SeverityInfo,
			)
			if err == nil {
				signals = append(signals, sig)
			}
		} else {
			signals = append(signals, d.checkVersionDrift(
				toolDir, toolName, entryName, currentData, tracked)...)
		}
	}
	return signals
}

func (d *KnowledgeDriftDetector) checkStaleness(rlmPath string, data map[string]any) []knowledgedomain.DriftSignal {
	metaRaw, ok := data["_meta"].(map[string]any)
	if !ok {
		return nil
	}

	reviewDateStr, ok := metaRaw["next_review_date"].(string)
	if !ok || reviewDateStr == "" {
		return nil
	}

	reviewDate, err := time.Parse("2006-01-02", reviewDateStr)
	if err != nil {
		return nil
	}

	if time.Now().After(reviewDate) {
		lastVerified := "unknown"
		if lv, ok := metaRaw["last_verified"].(string); ok {
			lastVerified = lv
		}
		sig, err := knowledgedomain.NewDriftSignal(
			rlmPath,
			knowledgedomain.DriftStale,
			fmt.Sprintf("Entry past review date (%s), last verified %s", reviewDateStr, lastVerified),
			knowledgedomain.SeverityWarning,
		)
		if err == nil {
			return []knowledgedomain.DriftSignal{sig}
		}
	}
	return nil
}

func (d *KnowledgeDriftDetector) checkVersionDrift(
	toolDir, toolName, entryName string,
	currentData map[string]any,
	tracked []string,
) []knowledgedomain.DriftSignal {
	var signals []knowledgedomain.DriftSignal
	rlmPath := fmt.Sprintf("tools/%s/%s", toolName, entryName)

	for _, version := range tracked {
		versionDir := filepath.Join(toolDir, version)
		if _, err := os.Stat(versionDir); os.IsNotExist(err) {
			sig, err := knowledgedomain.NewDriftSignal(
				rlmPath,
				knowledgedomain.DriftVersionChange,
				fmt.Sprintf("Version %s listed in _meta.toml but directory not found", version),
				knowledgedomain.SeverityWarning,
			)
			if err == nil {
				signals = append(signals, sig)
			}
			continue
		}

		versionEntry := filepath.Join(versionDir, entryName+".toml")
		if _, err := os.Stat(versionEntry); os.IsNotExist(err) {
			continue
		}

		versionData := readTOMLFile(versionEntry)
		signals = append(signals, d.diffEntries(rlmPath, version, currentData, versionData)...)
	}
	return signals
}

func (d *KnowledgeDriftDetector) diffEntries(
	rlmPath, version string,
	current, previous map[string]any,
) []knowledgedomain.DriftSignal {
	var signals []knowledgedomain.DriftSignal

	for sectionName, currentVal := range current {
		if sectionName == "_meta" {
			continue
		}
		currentSection, ok := currentVal.(map[string]any)
		if !ok {
			continue
		}

		previousVal, exists := previous[sectionName]
		if !exists {
			sig, err := knowledgedomain.NewDriftSignal(
				rlmPath,
				knowledgedomain.DriftVersionChange,
				fmt.Sprintf("Section [%s] added in current but missing in %s", sectionName, version),
				knowledgedomain.SeverityWarning,
			)
			if err == nil {
				signals = append(signals, sig)
			}
			continue
		}

		previousSection, ok := previousVal.(map[string]any)
		if !ok {
			continue
		}

		// Added keys
		added := sortedDiff(currentSection, previousSection)
		for _, key := range added {
			sig, err := knowledgedomain.NewDriftSignal(
				rlmPath,
				knowledgedomain.DriftVersionChange,
				fmt.Sprintf("Key '%s' in [%s] added in current but missing in %s", key, sectionName, version),
				knowledgedomain.SeverityWarning,
			)
			if err == nil {
				signals = append(signals, sig)
			}
		}

		// Removed keys
		removed := sortedDiff(previousSection, currentSection)
		for _, key := range removed {
			sig, err := knowledgedomain.NewDriftSignal(
				rlmPath,
				knowledgedomain.DriftVersionChange,
				fmt.Sprintf("Key '%s' in [%s] present in %s but removed in current", key, sectionName, version),
				knowledgedomain.SeverityWarning,
			)
			if err == nil {
				signals = append(signals, sig)
			}
		}
	}

	// Detect sections removed from current (only in previous)
	for sectionName, previousVal := range previous {
		if sectionName == "_meta" {
			continue
		}
		if _, ok := previousVal.(map[string]any); !ok {
			continue
		}
		if _, exists := current[sectionName]; !exists {
			sig, err := knowledgedomain.NewDriftSignal(
				rlmPath,
				knowledgedomain.DriftVersionChange,
				fmt.Sprintf("Section [%s] present in %s but removed in current", sectionName, version),
				knowledgedomain.SeverityWarning,
			)
			if err == nil {
				signals = append(signals, sig)
			}
		}
	}

	return signals
}

func readTOMLFile(path string) map[string]any {
	data, err := os.ReadFile(path)
	if err != nil {
		return make(map[string]any)
	}
	var result map[string]any
	if _, err := toml.Decode(string(data), &result); err != nil {
		return make(map[string]any)
	}
	return result
}

func toStringSlice(val any) []string {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			result = append(result, fmt.Sprintf("%v", item))
		}
		return result
	case []string:
		return v
	}
	return nil
}

func sortedDiff(a, b map[string]any) []string {
	var diff []string
	for key := range a {
		if _, exists := b[key]; !exists {
			diff = append(diff, key)
		}
	}
	sort.Strings(diff)
	return diff
}
