// Package output provides formatters for chexum output.
//
// DESIGN PRINCIPLE: Human-First, Machine-Ready
// -------------------------------------------
// Chexum believes that output should be immediately scannable by a human eye
// while remaining robust enough for machine consumption.
//
//  1. DEFAULT FORMAT: Prioritizes duplication detection by grouping identical
//     hashes together with blank line separators.
//  2. JSON/JSONL: Provides complete structured data for automated toolchains.
//  3. PLAIN: A tab-separated "grep-friendly" format for Unix veterans.
//
// Mandate: "No Lock-Out"
// We provide --preserve-order to ensure that our smart grouping defaults
// never prevent a user from seeing the raw sequence of their input.
package output

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Les-El/chexum/internal/hash"
	"github.com/Les-El/chexum/internal/security"
)

// Formatter is the interface for output formatters.
type Formatter interface {
	// Format formats the hash result for output.
	Format(result *hash.Result) string
}

// DefaultFormatter groups files by matching hash with blank lines between groups.
type DefaultFormatter struct{}

// Format implements Formatter for DefaultFormatter.
func (f *DefaultFormatter) Format(result *hash.Result) string {
	var sb strings.Builder

	f.writePoolMatches(&sb, result.PoolMatches)
	if len(result.PoolMatches) > 0 && (len(result.Matches) > 0 || len(result.Unmatched) > 0) {
		sb.WriteString("\n")
	}

	f.writeMatchGroups(&sb, result.Matches)
	if len(result.Matches) > 0 && len(result.Unmatched) > 0 {
		sb.WriteString("\n")
	}

	f.writeUnmatched(&sb, result.Unmatched)
	f.writeRefOrphans(&sb, result.RefOrphans, len(result.Matches) > 0 || len(result.Unmatched) > 0)
	f.writeUnknowns(&sb, result.Unknowns, len(result.Matches) > 0 || len(result.Unmatched) > 0 || len(result.RefOrphans) > 0)

	return strings.TrimSuffix(sb.String(), "\n")
}

func (f *DefaultFormatter) writePoolMatches(sb *strings.Builder, matches []hash.PoolMatch) {
	for _, m := range matches {
		sb.WriteString(fmt.Sprintf("Match, %s, %s, %s, %s\n",
			m.Algorithm, m.ProvidedHash, security.SanitizeOutput(m.FilePath), m.ComputedHash))
	}
}

func (f *DefaultFormatter) writeMatchGroups(sb *strings.Builder, groups []hash.MatchGroup) {
	for i, group := range groups {
		if i > 0 {
			sb.WriteString("\n")
		}
		for _, entry := range group.Entries {
			if entry.IsReference {
				sb.WriteString(fmt.Sprintf("REFERENCE:    %s\n", entry.Hash))
			} else {
				sb.WriteString(fmt.Sprintf("%s    %s\n", security.SanitizeOutput(entry.Original), entry.Hash))
			}
		}
	}
}

func (f *DefaultFormatter) writeUnmatched(sb *strings.Builder, unmatched []hash.Entry) {
	for i, entry := range unmatched {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%s    %s\n", security.SanitizeOutput(entry.Original), entry.Hash))
	}
}

func (f *DefaultFormatter) writeRefOrphans(sb *strings.Builder, orphans []hash.Entry, needsNewline bool) {
	if len(orphans) == 0 {
		return
	}
	if needsNewline {
		sb.WriteString("\n")
	}
	for i, entry := range orphans {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("REFERENCE:    %s\n", entry.Hash))
	}
}

func (f *DefaultFormatter) writeUnknowns(sb *strings.Builder, unknowns []string, needsNewline bool) {
	if len(unknowns) == 0 {
		return
	}
	if needsNewline {
		sb.WriteString("\n")
	}
	for i, unknown := range unknowns {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("INVALID:    %s\n", security.SanitizeOutput(unknown)))
	}
}

// PreserveOrderFormatter maintains input order without grouping.
type PreserveOrderFormatter struct{}

// Format implements Formatter for PreserveOrderFormatter.
func (f *PreserveOrderFormatter) Format(result *hash.Result) string {
	var sb strings.Builder

	for _, entry := range result.Entries {
		if entry.Error == nil {
			sb.WriteString(fmt.Sprintf("%s    %s\n", security.SanitizeOutput(entry.Original), entry.Hash))
		}
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// VerboseFormatter provides detailed output with summaries.
type VerboseFormatter struct{}

// Format implements Formatter for VerboseFormatter.
func (f *VerboseFormatter) Format(result *hash.Result) string {
	var sb strings.Builder

	// Header with processing stats
	sb.WriteString(fmt.Sprintf("Processed %d files in %s\n\n",
		result.FilesProcessed, result.Duration.Round(time.Millisecond)))

	// Match groups
	if len(result.Matches) > 0 {
		sb.WriteString("Match Groups:\n")
		for i, group := range result.Matches {
			sb.WriteString(fmt.Sprintf("  Group %d (%d files):\n", i+1, group.Count))
			for _, entry := range group.Entries {
				sb.WriteString(fmt.Sprintf("    %s    %s\n", security.SanitizeOutput(entry.Original), entry.Hash))
			}
			sb.WriteString("\n")
		}
	}

	// Unmatched files
	if len(result.Unmatched) > 0 {
		sb.WriteString("Unmatched Files:\n")
		for _, entry := range result.Unmatched {
			sb.WriteString(fmt.Sprintf("  %s    %s\n", security.SanitizeOutput(entry.Original), entry.Hash))
		}
		sb.WriteString("\n")
	}

	// Summary
	sb.WriteString(fmt.Sprintf("Summary: %d match groups, %d unmatched files",
		len(result.Matches), len(result.Unmatched)))

	return sb.String()
}

// JSONFormatter outputs results in machine-readable JSON format.
type JSONFormatter struct{}

// JSONLFormatter outputs results in line-delimited JSON format.
type JSONLFormatter struct{}

// jsonOutput is the structure for JSON output.
type jsonOutput struct {
	Processed   int              `json:"processed"`
	DurationMS  int64            `json:"duration_ms"`
	MatchGroups []jsonMatchGroup `json:"match_groups"`
	Unmatched   []jsonEntry      `json:"unmatched"`
	Errors      []string         `json:"errors"`
}

type jsonMatchGroup struct {
	Hash  string   `json:"hash"`
	Count int      `json:"count"`
	Files []string `json:"files"`
}

type jsonEntry struct {
	File string `json:"file"`
	Hash string `json:"hash"`
}

type jsonlEntry struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Format implements Formatter for JSONFormatter.
func (f *JSONFormatter) Format(result *hash.Result) string {
	output := jsonOutput{
		Processed:   result.FilesProcessed,
		DurationMS:  result.Duration.Milliseconds(),
		MatchGroups: make([]jsonMatchGroup, 0, len(result.Matches)),
		Unmatched:   make([]jsonEntry, 0, len(result.Unmatched)),
		Errors:      make([]string, 0, len(result.Errors)),
	}

	// Convert match groups
	for _, group := range result.Matches {
		files := make([]string, 0, len(group.Entries))
		for _, entry := range group.Entries {
			// We don't sanitize here because json.Marshal handles escapes
			files = append(files, entry.Original)
		}
		output.MatchGroups = append(output.MatchGroups, jsonMatchGroup{
			Hash:  group.Hash,
			Count: group.Count,
			Files: files,
		})
	}

	// Convert unmatched entries
	for _, entry := range result.Unmatched {
		output.Unmatched = append(output.Unmatched, jsonEntry{
			File: entry.Original,
			Hash: entry.Hash,
		})
	}

	// Convert errors
	for _, err := range result.Errors {
		output.Errors = append(output.Errors, err.Error())
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal JSON: %s"}`, err.Error())
	}

	return string(data)
}

// Format implements Formatter for JSONLFormatter.
func (f *JSONLFormatter) Format(result *hash.Result) string {
	var sb strings.Builder
	now := time.Now().Format(time.RFC3339)

	for _, entry := range result.Entries {
		status := "success"
		if entry.Error != nil {
			status = "error"
		}

		item := jsonlEntry{
			Type:      "file",
			Name:      entry.Original,
			Hash:      entry.Hash,
			Status:    status,
			Timestamp: now,
		}

		data, err := json.Marshal(item)
		if err == nil {
			sb.Write(data)
			sb.WriteString("\n")
		}
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// PlainFormatter outputs tab-separated results for scripting.
type PlainFormatter struct{}

// Format implements Formatter for PlainFormatter.
func (f *PlainFormatter) Format(result *hash.Result) string {
	var sb strings.Builder

	// Output all entries in input order, tab-separated
	for _, entry := range result.Entries {
		if entry.Error == nil {
			sb.WriteString(fmt.Sprintf("%s\t%s\n", security.SanitizeOutput(entry.Original), entry.Hash))
		}
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// CSVFormatter outputs results in CSV format with Type, Name, Hash, Algorithm columns.
type CSVFormatter struct{}

// Format implements Formatter for CSVFormatter.
func (f *CSVFormatter) Format(result *hash.Result) string {
	var sb strings.Builder

	// Output match groups
	for _, group := range result.Matches {
		for _, entry := range group.Entries {
			if entry.IsReference {
				sb.WriteString(fmt.Sprintf("REFERENCE,-,%s,%s\n", entry.Hash, entry.Algorithm))
			} else {
				sb.WriteString(fmt.Sprintf("FILE,%s,%s,%s\n",
					security.SanitizeOutput(entry.Original), entry.Hash, entry.Algorithm))
			}
		}
	}

	// Output unmatched files
	for _, entry := range result.Unmatched {
		sb.WriteString(fmt.Sprintf("FILE,%s,%s,%s\n",
			security.SanitizeOutput(entry.Original), entry.Hash, entry.Algorithm))
	}

	// Output orphaned reference hashes
	for _, entry := range result.RefOrphans {
		sb.WriteString(fmt.Sprintf("REFERENCE,-,%s,%s\n", entry.Hash, entry.Algorithm))
	}

	// Output invalid/unknown strings
	for _, unknown := range result.Unknowns {
		sb.WriteString(fmt.Sprintf("INVALID,%s,-,-\n", security.SanitizeOutput(unknown)))
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// NewFormatter creates a formatter based on the format name.
func NewFormatter(format string, preserveOrder bool) Formatter {
	switch format {
	case "verbose":
		return &VerboseFormatter{}
	case "json":
		return &JSONFormatter{}
	case "jsonl":
		return &JSONLFormatter{}
	case "plain":
		return &PlainFormatter{}
	case "csv":
		return &CSVFormatter{}
	default:
		if preserveOrder {
			return &PreserveOrderFormatter{}
		}
		return &DefaultFormatter{}
	}
}
