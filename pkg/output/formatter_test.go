// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package output

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestItem is a test struct used for formatting tests
type TestItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
	Count     int       `json:"count"`
	CreatedAt time.Time `json:"created_at"`
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
	currentTime := time.Now()
	testItems := []TestItem{
		{"1", "Item 1", true, 42, currentTime},
		{"2", "Item 2", false, 15, currentTime.Add(-24 * time.Hour)},
	}

	var buf bytes.Buffer
	f := NewFormatter("json", &buf)

	err := f.FormatOutput(testItems, []string{"ID", "Name", "Enabled", "Count", "CreatedAt"})
	if err != nil {
		t.Fatalf("FormatOutput returned error: %v", err)
	}

	// Verify JSON output
	var output []map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &output)
	if err != nil {
		t.Fatalf("Error unmarshaling JSON output: %v", err)
	}

	if len(output) != len(testItems) {
		t.Errorf("Expected %d items, got %d", len(testItems), len(output))
	}

	// Check if JSON contains the expected fields - keys may be lowercase due to JSON marshaling
	if (output[0]["ID"] != "1" && output[0]["id"] != "1") ||
		(output[0]["Name"] != "Item 1" && output[0]["name"] != "Item 1") {
		t.Errorf("JSON output doesn't contain expected values: %v", output[0])
	}
}

func TestFormatText(t *testing.T) {
	currentTime := time.Now()
	testItems := []TestItem{
		{"1", "Item 1", true, 42, currentTime},
		{"2", "Item 2", false, 15, currentTime.Add(-24 * time.Hour)},
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
	currentTime := time.Now()
	testItems := []TestItem{
		{"1", "Item 1", true, 42, currentTime},
		{"2", "Item 2", false, 15, currentTime.Add(-24 * time.Hour)},
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

func TestFormatStructWithTime(t *testing.T) {
	now := time.Now()
	item := TestItem{
		ID:        "test-id",
		Name:      "Test Item",
		Enabled:   true,
		Count:     123,
		CreatedAt: now,
	}

	var buf bytes.Buffer
	f := NewFormatter("text", &buf)

	headers := []string{"ID", "Name", "CreatedAt"}
	err := f.FormatOutput(item, headers)
	if err != nil {
		t.Fatalf("FormatOutput returned error: %v", err)
	}

	// Verify time field is properly formatted
	output := buf.String()
	if !strings.Contains(output, now.Format(time.RFC3339)[:10]) { // At least date part should be present
		t.Errorf("Output doesn't contain properly formatted time: %s", output)
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string", "test", "test"},
		{"int", 42, "42"},
		{"bool", true, "true"},
		{"float", 3.14, "3.14"},
		{"byte slice", []byte("hello"), "hello"},
		{"string slice", []string{"a", "b"}, "[a, b]"},
		{"map", map[string]int{"a": 1, "b": 2}, "a: 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(reflect.ValueOf(tt.input))
			if !strings.Contains(result, tt.expected) {
				t.Errorf("formatValue(%v) = %v, expected to contain %v", tt.input, result, tt.expected)
			}
		})
	}
}
