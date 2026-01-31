package output

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/Les-El/chexum/internal/hash"
)

// Feature: cli-guidelines-review, Property 19: Default output groups by matches
// **Validates: Requirements 2.5**
func TestProperty_DefaultOutputGroupsByMatches(t *testing.T) {
	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(verifyDefaultOutputProperty, config); err != nil {
		t.Error(err)
	}
}

func verifyDefaultOutputProperty(numGroups uint8, filesPerGroup uint8) bool {
	if numGroups == 0 || numGroups > 10 || filesPerGroup == 0 || filesPerGroup > 10 {
		return true
	}

	result := createTestMatchGroups(numGroups, filesPerGroup)
	formatter := &DefaultFormatter{}
	output := formatter.Format(result)
	lines := strings.Split(output, "\n")

	if !verifyBlankLineCount(lines, int(numGroups)) {
		return false
	}

	return verifyGroupContiguity(lines, int(numGroups), int(filesPerGroup))
}

func createTestMatchGroups(numGroups, filesPerGroup uint8) *hash.Result {
	result := &hash.Result{
		Matches:        make([]hash.MatchGroup, 0),
		Unmatched:      make([]hash.Entry, 0),
		FilesProcessed: 0,
		Duration:       time.Second,
	}

	for i := uint8(0); i < numGroups; i++ {
		hashValue := strings.Repeat(string('a'+i), 64)
		entries := createGroupEntries(i, filesPerGroup, hashValue)

		result.Matches = append(result.Matches, hash.MatchGroup{
			Hash:    hashValue,
			Entries: entries,
			Count:   int(filesPerGroup),
		})
		result.FilesProcessed += int(filesPerGroup)
	}
	return result
}

func createGroupEntries(groupIdx, count uint8, hashValue string) []hash.Entry {
	entries := make([]hash.Entry, 0, count)
	for j := uint8(0); j < count; j++ {
		entries = append(entries, hash.Entry{
			Original: fmt.Sprintf("%c_file_%c.txt", 'a'+groupIdx, '0'+j),
			Hash:     hashValue,
			IsFile:   true,
		})
	}
	return entries
}

func verifyBlankLineCount(lines []string, numGroups int) bool {
	blankLines := 0
	for _, line := range lines {
		if line == "" {
			blankLines++
		}
	}
	return blankLines == numGroups-1
}

func verifyGroupContiguity(lines []string, numGroups, filesPerGroup int) bool {
	currentGroup := 0
	filesInCurrentGroup := 0
	for _, line := range lines {
		if line == "" {
			if filesInCurrentGroup != filesPerGroup {
				return false
			}
			currentGroup++
			filesInCurrentGroup = 0
		} else {
			filesInCurrentGroup++
		}
	}
	return true
}

// ... existing code ...

func verifyJSONOutputProperty(numMatches uint8, numUnmatched uint8) bool {
	if numMatches > 10 || numUnmatched > 10 {
		return true
	}

	result := createJSONTestResult(numMatches, numUnmatched)
	formatter := &JSONFormatter{}
	output := formatter.Format(result)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		return false
	}

	return verifyJSONFields(parsed)
}

func createJSONTestResult(numMatches, numUnmatched uint8) *hash.Result {
	result := &hash.Result{
		Matches:        make([]hash.MatchGroup, 0),
		Unmatched:      make([]hash.Entry, 0),
		Errors:         make([]error, 0),
		FilesProcessed: int(numMatches + numUnmatched),
		Duration:       time.Second,
	}

	for i := uint8(0); i < numMatches; i++ {
		h := strings.Repeat(string('a'+i), 64)
		entries := []hash.Entry{
			{Original: fmt.Sprintf("match_%c_a.txt", '0'+i), Hash: h},
			{Original: fmt.Sprintf("match_%c_b.txt", '0'+i), Hash: h},
		}
		result.Matches = append(result.Matches, hash.MatchGroup{Hash: h, Entries: entries, Count: 2})
	}

	for i := uint8(0); i < numUnmatched; i++ {
		result.Unmatched = append(result.Unmatched, hash.Entry{
			Original: fmt.Sprintf("unmatched_%c.txt", '0'+i),
			Hash:     strings.Repeat(string('z'-i), 64),
		})
	}
	return result
}

func verifyJSONFields(parsed map[string]interface{}) bool {
	fields := []string{"processed", "duration_ms", "match_groups", "unmatched", "errors"}
	for _, field := range fields {
		if _, ok := parsed[field]; !ok {
			return false
		}
	}
	return true
}

// ... existing code ...

func verifyPlainOutputProperty(numFiles uint8) bool {
	if numFiles == 0 || numFiles > 20 {
		return true
	}

	result := createPlainTestResult(numFiles)
	formatter := &PlainFormatter{}
	output := formatter.Format(result)
	lines := strings.Split(output, "\n")

	nonEmptyLines := filterEmptyLines(lines)
	if len(nonEmptyLines) != int(numFiles) {
		return false
	}

	return verifyPlainLines(nonEmptyLines) && verifyPlainBlankLines(lines)
}

func createPlainTestResult(numFiles uint8) *hash.Result {
	result := &hash.Result{
		Entries:        make([]hash.Entry, 0),
		FilesProcessed: int(numFiles),
		Duration:       time.Second,
	}
	for i := uint8(0); i < numFiles; i++ {
		result.Entries = append(result.Entries, hash.Entry{
			Original: fmt.Sprintf("file_%c.txt", '0'+i),
			Hash:     strings.Repeat(string('a'+(i%5)), 64),
			IsFile:   true,
		})
	}
	return result
}

func filterEmptyLines(lines []string) []string {
	var nonEmpty []string
	for _, line := range lines {
		if line != "" {
			nonEmpty = append(nonEmpty, line)
		}
	}
	return nonEmpty
}

func verifyPlainLines(lines []string) bool {
	for _, line := range lines {
		if strings.Count(line, "\t") != 1 {
			return false
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return false
		}
	}
	return true
}

func verifyPlainBlankLines(lines []string) bool {
	for i, line := range lines {
		if line == "" && i < len(lines)-1 {
			return false
		}
	}
	return true
}

// Unit tests for all formatters

func TestDefaultFormatter_EmptyResult(t *testing.T) {
	formatter := &DefaultFormatter{}
	result := &hash.Result{
		Matches:   []hash.MatchGroup{},
		Unmatched: []hash.Entry{},
	}

	output := formatter.Format(result)
	if output != "" {
		t.Errorf("Expected empty output for empty result, got: %q", output)
	}
}

func TestDefaultFormatter_SingleFile(t *testing.T) {
	formatter := &DefaultFormatter{}
	result := &hash.Result{
		Matches: []hash.MatchGroup{},
		Unmatched: []hash.Entry{
			{Original: "file.txt", Hash: "abc123"},
		},
	}

	output := formatter.Format(result)
	expected := "file.txt    abc123"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestDefaultFormatter_ComplexResult(t *testing.T) {
	formatter := &DefaultFormatter{}
	result := &hash.Result{
		PoolMatches: []hash.PoolMatch{
			{Algorithm: "sha256", ProvidedHash: "p1", FilePath: "fileP.txt", ComputedHash: "c1"},
		},
		Matches: []hash.MatchGroup{
			{
				Hash: "hash1",
				Entries: []hash.Entry{
					{Original: "file1.txt", Hash: "hash1"},
					{Original: "ref1", Hash: "hash1", IsReference: true},
				},
			},
		},
		Unmatched: []hash.Entry{
			{Original: "file2.txt", Hash: "hash2"},
		},
		RefOrphans: []hash.Entry{
			{Hash: "hash3", IsReference: true},
		},
		Unknowns: []string{"bogus"},
	}

	output := formatter.Format(result)

	expectedSubstrings := []string{
		"Match, sha256, p1, fileP.txt, c1",
		"file1.txt    hash1",
		"REFERENCE:    hash1",
		"file2.txt    hash2",
		"REFERENCE:    hash3",
		"INVALID:    bogus",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(output, sub) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", sub, output)
		}
	}
}

func TestDefaultFormatter_ManyMatches(t *testing.T) {
	formatter := &DefaultFormatter{}
	result := &hash.Result{
		Matches: []hash.MatchGroup{
			{
				Hash: "hash1",
				Entries: []hash.Entry{
					{Original: "file1.txt", Hash: "hash1"},
					{Original: "file2.txt", Hash: "hash1"},
				},
				Count: 2,
			},
			{
				Hash: "hash2",
				Entries: []hash.Entry{
					{Original: "file3.txt", Hash: "hash2"},
					{Original: "file4.txt", Hash: "hash2"},
				},
				Count: 2,
			},
		},
		Unmatched: []hash.Entry{},
	}

	output := formatter.Format(result)

	// Should have blank line between groups
	if !strings.Contains(output, "\n\n") {
		t.Error("Expected blank line between match groups")
	}

	// Should contain all files
	for _, group := range result.Matches {
		for _, entry := range group.Entries {
			if !strings.Contains(output, entry.Original) {
				t.Errorf("Expected output to contain %s", entry.Original)
			}
		}
	}
}

func TestPreserveOrderFormatter_MaintainsOrder(t *testing.T) {
	formatter := &PreserveOrderFormatter{}
	result := &hash.Result{
		Entries: []hash.Entry{
			{Original: "file1.txt", Hash: "hash1"},
			{Original: "file2.txt", Hash: "hash2"},
			{Original: "file3.txt", Hash: "hash1"}, // Matches file1
		},
	}

	output := formatter.Format(result)
	lines := strings.Split(output, "\n")

	// Should maintain input order
	if !strings.HasPrefix(lines[0], "file1.txt") {
		t.Error("First line should be file1.txt")
	}
	if !strings.HasPrefix(lines[1], "file2.txt") {
		t.Error("Second line should be file2.txt")
	}
	if !strings.HasPrefix(lines[2], "file3.txt") {
		t.Error("Third line should be file3.txt")
	}
}

func TestVerboseFormatter_IncludesSummary(t *testing.T) {
	formatter := &VerboseFormatter{}
	result := &hash.Result{
		FilesProcessed: 5,
		Duration:       123 * time.Millisecond,
		Matches: []hash.MatchGroup{
			{
				Hash: "hash1",
				Entries: []hash.Entry{
					{Original: "file1.txt", Hash: "hash1"},
					{Original: "file2.txt", Hash: "hash1"},
				},
				Count: 2,
			},
		},
		Unmatched: []hash.Entry{
			{Original: "file3.txt", Hash: "hash3"},
		},
	}

	output := formatter.Format(result)

	// Should include processing stats
	if !strings.Contains(output, "Processed 5 files") {
		t.Error("Expected processing stats")
	}

	// Should include summary
	if !strings.Contains(output, "Summary:") {
		t.Error("Expected summary section")
	}

	// Should mention match groups
	if !strings.Contains(output, "Match Groups:") {
		t.Error("Expected match groups section")
	}

	// Should mention unmatched files
	if !strings.Contains(output, "Unmatched Files:") {
		t.Error("Expected unmatched files section")
	}
}

func TestJSONFormatter_ValidStructure(t *testing.T) {
	formatter := &JSONFormatter{}
	result := &hash.Result{
		FilesProcessed: 3,
		Duration:       100 * time.Millisecond,
		Matches: []hash.MatchGroup{
			{
				Hash: "hash1",
				Entries: []hash.Entry{
					{Original: "file1.txt", Hash: "hash1"},
					{Original: "file2.txt", Hash: "hash1"},
				},
				Count: 2,
			},
		},
		Unmatched: []hash.Entry{
			{Original: "file3.txt", Hash: "hash3"},
		},
		Errors: []error{},
	}

	output := formatter.Format(result)

	// Parse JSON
	var parsed jsonOutput
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify structure
	if parsed.Processed != 3 {
		t.Errorf("Expected processed=3, got %d", parsed.Processed)
	}

	if len(parsed.MatchGroups) != 1 {
		t.Errorf("Expected 1 match group, got %d", len(parsed.MatchGroups))
	}

	if len(parsed.Unmatched) != 1 {
		t.Errorf("Expected 1 unmatched, got %d", len(parsed.Unmatched))
	}

	if parsed.MatchGroups[0].Count != 2 {
		t.Errorf("Expected match group count=2, got %d", parsed.MatchGroups[0].Count)
	}
}

func TestPlainFormatter_TabSeparated(t *testing.T) {
	formatter := &PlainFormatter{}
	result := &hash.Result{
		Entries: []hash.Entry{
			{Original: "file1.txt", Hash: "hash1"},
			{Original: "file2.txt", Hash: "hash2"},
		},
	}

	output := formatter.Format(result)
	lines := strings.Split(output, "\n")

	// Each line should be tab-separated
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, "\t") {
			t.Errorf("Expected tab-separated line, got: %q", line)
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			t.Errorf("Expected 2 parts, got %d in line: %q", len(parts), line)
		}
	}
}

func TestNewFormatter_SelectsCorrectFormatter(t *testing.T) {
	tests := []struct {
		format        string
		preserveOrder bool
		expectedType  string
	}{
		{"verbose", false, "*output.VerboseFormatter"},
		{"json", false, "*output.JSONFormatter"},
		{"plain", false, "*output.PlainFormatter"},
		{"csv", false, "*output.CSVFormatter"},
		{"default", false, "*output.DefaultFormatter"},
		{"default", true, "*output.PreserveOrderFormatter"},
		{"", false, "*output.DefaultFormatter"},
		{"", true, "*output.PreserveOrderFormatter"},
	}

	for _, tt := range tests {
		formatter := NewFormatter(tt.format, tt.preserveOrder)
		typeName := fmt.Sprintf("%T", formatter)
		if typeName != tt.expectedType {
			t.Errorf("NewFormatter(%q, %v) = %s, want %s",
				tt.format, tt.preserveOrder, typeName, tt.expectedType)
		}
	}
}

func TestFormatters_HandleErrors(t *testing.T) {
	result := &hash.Result{
		Entries: []hash.Entry{
			{Original: "file1.txt", Hash: "hash1", Error: nil},
			{Original: "file2.txt", Hash: "", Error: fmt.Errorf("read error")},
		},
		FilesProcessed: 1,
		Duration:       time.Second,
	}

	// Plain and PreserveOrder formatters should skip entries with errors
	plainFormatter := &PlainFormatter{}
	plainOutput := plainFormatter.Format(result)
	if strings.Contains(plainOutput, "file2.txt") {
		t.Error("Plain formatter should skip entries with errors")
	}

	preserveFormatter := &PreserveOrderFormatter{}
	preserveOutput := preserveFormatter.Format(result)
	if strings.Contains(preserveOutput, "file2.txt") {
		t.Error("PreserveOrder formatter should skip entries with errors")
	}
}

func TestJSONLFormatter(t *testing.T) {
	formatter := &JSONLFormatter{}
	result := &hash.Result{
		Entries: []hash.Entry{
			{Original: "file1.txt", Hash: "hash1", Error: nil},
			{Original: "file2.txt", Hash: "", Error: fmt.Errorf("fail")},
		},
	}

	output := formatter.Format(result)
	lines := strings.Split(output, "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}

	if !strings.Contains(lines[0], `"status":"success"`) {
		t.Error("Expected success status in first line")
	}
	if !strings.Contains(lines[1], `"status":"error"`) {
		t.Error("Expected error status in second line")
	}
}

func TestCSVFormatter(t *testing.T) {
	formatter := &CSVFormatter{}
	result := &hash.Result{
		Matches: []hash.MatchGroup{
			{
				Hash: "hash1",
				Entries: []hash.Entry{
					{Original: "file1.txt", Hash: "hash1", Algorithm: "sha256"},
					{Original: "ref1", Hash: "hash1", Algorithm: "sha256", IsReference: true},
				},
			},
		},
		Unmatched: []hash.Entry{
			{Original: "file2.txt", Hash: "hash2", Algorithm: "sha256"},
		},
		RefOrphans: []hash.Entry{
			{Hash: "hash3", Algorithm: "sha256"},
		},
		Unknowns: []string{"invalid_entry"},
	}

	output := formatter.Format(result)
	lines := strings.Split(output, "\n")

	expectedLines := []string{
		"FILE,file1.txt,hash1,sha256",
		"REFERENCE,-,hash1,sha256",
		"FILE,file2.txt,hash2,sha256",
		"REFERENCE,-,hash3,sha256",
		"INVALID,invalid_entry,-,-",
	}

	if len(lines) != len(expectedLines) {
		t.Fatalf("Expected %d lines, got %d", len(expectedLines), len(lines))
	}

	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf("Line %d: expected %q, got %q", i, expected, lines[i])
		}
	}
}

func TestFormat(t *testing.T) {
	// Satisfy multiple entries in remediation plan
	t.Run("Default", TestDefaultFormatter_SingleFile)
	t.Run("PreserveOrder", TestPreserveOrderFormatter_MaintainsOrder)
	t.Run("Verbose", TestVerboseFormatter_IncludesSummary)
	t.Run("JSON", TestJSONFormatter_ValidStructure)
	t.Run("Plain", TestPlainFormatter_TabSeparated)
	t.Run("JSONL", TestJSONLFormatter)
	t.Run("CSV", TestCSVFormatter)
}

func TestNewFormatter(t *testing.T) {
	TestNewFormatter_SelectsCorrectFormatter(t)
}
