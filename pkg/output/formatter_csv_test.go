// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package output

import (
	"bytes"
	"encoding/csv"
	"reflect"
	"testing"
	"time"
)

type CSVTestItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Count     int       `json:"count"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func TestWriteCSVRow(t *testing.T) {
	// Create test item
	now := time.Now()
	item := CSVTestItem{
		ID:        "123",
		Name:      "Test Item",
		Count:     42,
		IsActive:  true,
		CreatedAt: now,
	}

	// Create CSV writer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Define headers matching struct field names
	headers := []string{"ID", "Name", "Count", "IsActive", "CreatedAt"}

	// Call writeCSVRow with reflect.ValueOf(item)
	err := writeCSVRow(writer, reflect.ValueOf(item), headers)
	if err != nil {
		t.Fatalf("writeCSVRow returned error: %v", err)
	}

	// Flush the writer
	writer.Flush()

	// Check for errors during write
	if err := writer.Error(); err != nil {
		t.Fatalf("CSV writer error: %v", err)
	}

	// Parse the output
	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	record, err := reader.Read()
	if err != nil {
		t.Fatalf("Error reading CSV: %v", err)
	}

	// Verify the CSV record
	expectedValues := []string{
		"123",        // ID
		"Test Item",  // Name
		"42",         // Count
		"true",       // IsActive
		now.String(), // CreatedAt (approximate check)
	}

	if len(record) != len(expectedValues) {
		t.Fatalf("Expected %d fields, got %d", len(expectedValues), len(record))
	}

	// Check each field
	for i, expected := range expectedValues {
		// For CreatedAt, just check if the field is non-empty
		if i == 4 {
			if record[i] == "" {
				t.Errorf("Field %d (CreatedAt) is empty", i)
			}
			continue
		}

		if record[i] != expected {
			t.Errorf("Field %d: expected %q, got %q", i, expected, record[i])
		}
	}
}

func TestWriteCSVRow_CaseInsensitiveHeaders(t *testing.T) {
	// Create test item
	item := CSVTestItem{
		ID:       "123",
		Name:     "Test Item",
		Count:    42,
		IsActive: true,
	}

	// Create CSV writer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Define headers with different case than struct fields
	headers := []string{"id", "name", "count", "isactive"}

	// Call writeCSVRow with reflect.ValueOf(item)
	err := writeCSVRow(writer, reflect.ValueOf(item), headers)
	if err != nil {
		t.Fatalf("writeCSVRow returned error: %v", err)
	}

	// Flush the writer
	writer.Flush()

	// Parse the output
	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	record, err := reader.Read()
	if err != nil {
		t.Fatalf("Error reading CSV: %v", err)
	}

	expectedValues := []string{"123", "Test Item", "42", "true"}
	for i, expected := range expectedValues {
		if record[i] != expected {
			t.Errorf("Field %d: expected %q, got %q", i, expected, record[i])
		}
	}
}

func TestWriteCSVRow_NonExistentField(t *testing.T) {
	// Create test item
	item := CSVTestItem{
		ID:   "123",
		Name: "Test Item",
	}

	// Create CSV writer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Define headers with a non-existent field
	headers := []string{"ID", "Name", "NonExistentField"}

	// Call writeCSVRow with reflect.ValueOf(item)
	err := writeCSVRow(writer, reflect.ValueOf(item), headers)
	if err != nil {
		t.Fatalf("writeCSVRow returned error: %v", err)
	}

	// Flush the writer
	writer.Flush()

	// Parse the output
	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	record, err := reader.Read()
	if err != nil {
		t.Fatalf("Error reading CSV: %v", err)
	}

	// The non-existent field should be empty
	expectedValues := []string{"123", "Test Item", ""}
	for i, expected := range expectedValues {
		if record[i] != expected {
			t.Errorf("Field %d: expected %q, got %q", i, expected, record[i])
		}
	}
}
