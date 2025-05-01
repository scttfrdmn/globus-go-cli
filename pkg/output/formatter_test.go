// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestItem is a test struct used for formatting tests
type TestItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Count   int    `json:"count"`
}

// TestItems is a collection of TestItem
type TestItems struct {
	Items []TestItem `json:"items"`
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name           string
		formatStr      string
		expectedFormat FormatType
	}{
		{"default format", "text", FormatText},
		{"json format", "json", FormatJSON},
		{"csv format", "csv", FormatCSV},
		{"unknown format", "unknown", FormatText},
		{"case insensitive", "JSON", FormatJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := NewFormatter(tt.formatStr, &buf)
			if f.Format != tt.expectedFormat {
				t.Errorf("Expected format %v, got %v", tt.expectedFormat, f.Format)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	testItems := []TestItem{
		{"1", "Item 1", true, 42},
		{"2", "Item 2", false, 15},
	}

	var buf bytes.Buffer
	f := NewFormatter("json", &buf)

	err := f.FormatOutput(testItems, []string{"ID", "Name", "Enabled", "Count"})
	if err != nil {
		t.Fatalf("FormatOutput returned error: %v", err)
	}

	// Verify JSON output
	var output []TestItem
	err = json.Unmarshal(buf.Bytes(), &output)
	if err != nil {
		t.Fatalf("Error unmarshaling JSON output: %v", err)
	}

	if len(output) != len(testItems) {
		t.Errorf("Expected %d items, got %d", len(testItems), len(output))
	}

	for i, item := range output {
		if item.ID != testItems[i].ID || 
		   item.Name != testItems[i].Name || 
		   item.Enabled != testItems[i].Enabled || 
		   item.Count != testItems[i].Count {
			t.Errorf("Item %d doesn't match expected: got %+v, want %+v", 
				i, item, testItems[i])
		}
	}
}

func TestFormatText(t *testing.T) {
	testItems := []TestItem{
		{"1", "Item 1", true, 42},
		{"2", "Item 2", false, 15},
	}

	var buf bytes.Buffer
	f := NewFormatter("text", &buf)

	headers := []string{"ID", "Name", "Enabled", "Count"}
	err := f.FormatOutput(testItems, headers)
	if err != nil {
		t.Fatalf("FormatOutput returned error: %v", err)
	}

	// Verify text output has headers and data
	output := buf.String()
	for _, h := range headers {
		if !strings.Contains(output, h) {
			t.Errorf("Output doesn't contain header '%s': %s", h, output)
		}
	}

	// Check for item data in output
	for _, item := range testItems {
		if !strings.Contains(output, item.ID) || !strings.Contains(output, item.Name) {
			t.Errorf("Output doesn't contain expected item data: %s", output)
		}
	}
}

func TestFormatCSV(t *testing.T) {
	testItems := []TestItem{
		{"1", "Item 1", true, 42},
		{"2", "Item 2", false, 15},
	}

	var buf bytes.Buffer
	f := NewFormatter("csv", &buf)

	headers := []string{"ID", "Name", "Enabled", "Count"}
	err := f.FormatOutput(testItems, headers)
	if err != nil {
		t.Fatalf("FormatOutput returned error: %v", err)
	}

	// Verify CSV output
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	// Check header line
	if len(lines) < 3 { // Header + 2 data lines
		t.Fatalf("Expected at least 3 lines of output, got %d", len(lines))
	}
	
	// Check header contains all expected headers
	headerLine := lines[0]
	for _, h := range headers {
		if !strings.Contains(headerLine, h) {
			t.Errorf("Header line doesn't contain '%s': %s", h, headerLine)
		}
	}

	// Check data lines
	for i, item := range testItems {
		line := lines[i+1]
		if !strings.Contains(line, item.ID) || !strings.Contains(line, item.Name) {
			t.Errorf("Line %d doesn't contain expected item data: %s", i+1, line)
		}
	}
}
