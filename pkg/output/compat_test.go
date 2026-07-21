// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package output

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestParseFormatUnix(t *testing.T) {
	if got := parseFormat("unix"); got != FormatUnix {
		t.Errorf("parseFormat(unix) = %v, want %v", got, FormatUnix)
	}
	if got := parseFormat("UNIX"); got != FormatUnix {
		t.Errorf("parseFormat(UNIX) = %v, want %v (case-insensitive)", got, FormatUnix)
	}
}

func TestFormatUnix(t *testing.T) {
	type row struct {
		Name string
		Size int
	}
	var buf bytes.Buffer
	f := NewFormatter("unix", &buf)
	if err := f.FormatOutput([]row{{"a.txt", 10}, {"b.txt", 20}}, []string{"Name", "Size"}); err != nil {
		t.Fatalf("FormatOutput: %v", err)
	}
	out := buf.String()
	// No header line; tab-delimited rows.
	if strings.Contains(out, "Name") {
		t.Errorf("unix output should have no header, got:\n%s", out)
	}
	if out != "a.txt\t10\nb.txt\t20\n" {
		t.Errorf("unexpected unix output:\n%q", out)
	}
}

func TestJMESPathForcesJSONAndFilters(t *testing.T) {
	type item struct {
		Name string `json:"name"`
	}
	var buf bytes.Buffer
	f := NewFormatterWithJMESPath("text", "[*].name", &buf)
	if f.Format != FormatJSON {
		t.Errorf("a jmespath expression should force JSON, got %v", f.Format)
	}
	if err := f.FormatOutput([]item{{"one"}, {"two"}}, nil); err != nil {
		t.Fatalf("FormatOutput: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "one") || !strings.Contains(out, "two") {
		t.Errorf("jmespath output missing elements:\n%s", out)
	}
	// The projection [*].name yields a JSON array of strings, not objects.
	if strings.Contains(out, "\"name\"") {
		t.Errorf("jmespath [*].name should strip object keys, got:\n%s", out)
	}
}

func TestJMESPathHook(t *testing.T) {
	old := JMESPathHook
	defer func() { JMESPathHook = old }()
	JMESPathHook = func() string { return "foo" }

	f := NewFormatter("text", &bytes.Buffer{})
	if f.JMESPath != "foo" || f.Format != FormatJSON {
		t.Errorf("JMESPathHook should set expression and force JSON; got jmespath=%q format=%v", f.JMESPath, f.Format)
	}
}

func TestParseHTTPStatusMap(t *testing.T) {
	m, err := ParseHTTPStatusMap("404=50,403=51")
	if err != nil {
		t.Fatalf("ParseHTTPStatusMap: %v", err)
	}
	if m[404] != 50 || m[403] != 51 {
		t.Errorf("unexpected map: %v", m)
	}

	if m, err := ParseHTTPStatusMap(""); err != nil || m != nil {
		t.Errorf("empty spec should yield (nil,nil), got (%v,%v)", m, err)
	}

	// Out-of-range exit code (2-49 not allowed).
	if _, err := ParseHTTPStatusMap("404=42"); err == nil {
		t.Error("expected error for out-of-range exit code 42")
	}
	// Malformed entry.
	if _, err := ParseHTTPStatusMap("404"); err == nil {
		t.Error("expected error for malformed entry")
	}
}

type statusErr struct{ status int }

func (e statusErr) Error() string   { return "boom" }
func (e statusErr) HTTPStatus() int { return e.status }

func TestExitCodeForError(t *testing.T) {
	m := map[int]int{404: 50}

	if code, ok := ExitCodeForError(statusErr{404}, m); !ok || code != 50 {
		t.Errorf("mapped 404 should give (50,true), got (%d,%v)", code, ok)
	}
	// Status not in map.
	if _, ok := ExitCodeForError(statusErr{500}, m); ok {
		t.Error("unmapped status should give ok=false")
	}
	// Error without a status.
	if _, ok := ExitCodeForError(errors.New("plain"), m); ok {
		t.Error("statusless error should give ok=false")
	}
	// Nil error / empty map.
	if _, ok := ExitCodeForError(nil, m); ok {
		t.Error("nil error should give ok=false")
	}
}
