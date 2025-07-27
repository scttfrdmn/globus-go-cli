// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"
)

// FormatType defines the output format
type FormatType string

const (
	// FormatText outputs in a human-readable text format
	FormatText FormatType = "text"
	// FormatJSON outputs in JSON format
	FormatJSON FormatType = "json"
	// FormatCSV outputs in CSV format
	FormatCSV FormatType = "csv"
)

// Formatter handles formatting output in different formats
type Formatter struct {
	Format FormatType
	Writer io.Writer
}

// NewFormatter creates a new formatter
func NewFormatter(format string, writer io.Writer) *Formatter {
	formatType := FormatText
	switch strings.ToLower(format) {
	case "json":
		formatType = FormatJSON
	case "csv":
		formatType = FormatCSV
	}

	return &Formatter{
		Format: formatType,
		Writer: writer,
	}
}

// FormatOutput formats the given data according to the configured format
func (f *Formatter) FormatOutput(data interface{}, headers []string) error {
	switch f.Format {
	case FormatJSON:
		return f.formatJSON(data)
	case FormatCSV:
		return f.formatCSV(data, headers)
	default:
		return f.formatText(data, headers)
	}
}

// formatJSON formats the data as JSON
func (f *Formatter) formatJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	fmt.Fprintln(f.Writer, string(jsonData))
	return nil
}

// formatCSV formats the data as CSV
func (f *Formatter) formatCSV(data interface{}, headers []string) error {
	w := csv.NewWriter(f.Writer)
	defer w.Flush()

	// Write headers
	if err := w.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write data
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			item := value.Index(i)
			if err := writeCSVRow(w, item, headers); err != nil {
				return err
			}
		}
	case reflect.Struct:
		if err := writeCSVRow(w, value, headers); err != nil {
			return err
		}
	case reflect.Map:
		itemsField := value.FieldByName("Items")
		if itemsField.IsValid() && itemsField.Kind() == reflect.Slice {
			for i := 0; i < itemsField.Len(); i++ {
				item := itemsField.Index(i)
				if err := writeCSVRow(w, item, headers); err != nil {
					return err
				}
			}
		} else {
			if err := writeCSVRow(w, value, headers); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported data type for CSV formatting: %v", value.Kind())
	}

	return nil
}

// writeCSVRow writes a single row of CSV data
func writeCSVRow(w *csv.Writer, item reflect.Value, headers []string) error {
	if item.Kind() == reflect.Ptr {
		item = item.Elem()
	}

	row := make([]string, len(headers))
	for i, header := range headers {
		field := item.FieldByName(header)
		if !field.IsValid() {
			// Try case-insensitive match
			for j := 0; j < item.NumField(); j++ {
				if strings.EqualFold(item.Type().Field(j).Name, header) {
					field = item.Field(j)
					break
				}
			}
		}

		if field.IsValid() {
			row[i] = formatValue(field)
		}
	}

	return w.Write(row)
}

// formatText formats the data as human-readable text
func (f *Formatter) formatText(data interface{}, headers []string) error {
	w := tabwriter.NewWriter(f.Writer, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Write headers
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	fmt.Fprintln(w, strings.Repeat("-", len(headers)*10))

	// Write data
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			item := value.Index(i)
			if err := writeTextRow(w, item, headers); err != nil {
				return err
			}
		}
	case reflect.Struct:
		if err := writeTextRow(w, value, headers); err != nil {
			return err
		}
	case reflect.Map:
		itemsField := value.FieldByName("Items")
		if itemsField.IsValid() && itemsField.Kind() == reflect.Slice {
			for i := 0; i < itemsField.Len(); i++ {
				item := itemsField.Index(i)
				if err := writeTextRow(w, item, headers); err != nil {
					return err
				}
			}
		} else {
			if err := writeTextRow(w, value, headers); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported data type for text formatting: %v", value.Kind())
	}

	return nil
}

// writeTextRow writes a single row of tabular text data
func writeTextRow(w *tabwriter.Writer, item reflect.Value, headers []string) error {
	if item.Kind() == reflect.Ptr {
		item = item.Elem()
	}

	rowValues := make([]string, len(headers))
	for i, header := range headers {
		field := item.FieldByName(header)
		if !field.IsValid() {
			// Try case-insensitive match
			for j := 0; j < item.NumField(); j++ {
				if strings.EqualFold(item.Type().Field(j).Name, header) {
					field = item.Field(j)
					break
				}
			}
		}

		if field.IsValid() {
			rowValues[i] = formatValue(field)
		}
	}

	fmt.Fprintln(w, strings.Join(rowValues, "\t"))
	return nil
}

// formatValue formats a reflect.Value as a string
func formatValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', 2, 64)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Slice, reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// Handle []byte as string
			return string(v.Bytes())
		}
		// For other slices, format as comma-separated list
		var result strings.Builder
		result.WriteString("[")
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				result.WriteString(", ")
			}
			result.WriteString(formatValue(v.Index(i)))
		}
		result.WriteString("]")
		return result.String()
	case reflect.Map:
		var result strings.Builder
		result.WriteString("{")
		iter := v.MapRange()
		first := true
		for iter.Next() {
			if !first {
				result.WriteString(", ")
			}
			first = false
			result.WriteString(formatValue(iter.Key()))
			result.WriteString(": ")
			result.WriteString(formatValue(iter.Value()))
		}
		result.WriteString("}")
		return result.String()
	case reflect.Struct:
		// Handle common standard library structs
		if v.Type().String() == "time.Time" {
			// Format as string to avoid import cycle
			return fmt.Sprintf("%v", v.Interface())
		}
		return fmt.Sprintf("%+v", v.Interface())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
