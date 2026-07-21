// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package output

import (
	"fmt"
	"strconv"
	"strings"
)

// HTTPStatusError is implemented by errors that carry an HTTP status code, so
// --map-http-status can map them to process exit codes. The SDK's API errors
// expose their status via a StatusCode field rather than a method, so the CLI
// wraps them in a type implementing this interface before mapping (see
// cmd.httpStatusError). Keeping pkg/output free of an SDK import.
type HTTPStatusError interface {
	error
	HTTPStatus() int
}

// ParseHTTPStatusMap parses a --map-http-status value of the form
// "404=50,403=51" into a map of HTTP status -> exit code. Exit codes are
// restricted to 0, 1, or 50-99 (matching the Python CLI). An empty string
// yields a nil map and no error.
func ParseHTTPStatusMap(spec string) (map[int]int, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, nil
	}
	result := make(map[int]int)
	for _, pair := range strings.Split(spec, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid --map-http-status entry %q (want STATUS=EXITCODE)", pair)
		}
		status, err := strconv.Atoi(strings.TrimSpace(kv[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid HTTP status %q in --map-http-status", kv[0])
		}
		code, err := strconv.Atoi(strings.TrimSpace(kv[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid exit code %q in --map-http-status", kv[1])
		}
		if !validExitCode(code) {
			return nil, fmt.Errorf("exit code %d out of range in --map-http-status (allowed: 0, 1, 50-99)", code)
		}
		result[status] = code
	}
	return result, nil
}

func validExitCode(code int) bool {
	return code == 0 || code == 1 || (code >= 50 && code <= 99)
}

// ExitCodeForError returns the exit code an error should produce given a parsed
// --map-http-status map. If the error carries an HTTP status present in the map,
// the mapped code is returned along with true; otherwise (0, false) meaning the
// caller should fall back to its default exit behavior.
func ExitCodeForError(err error, statusMap map[int]int) (int, bool) {
	if err == nil || len(statusMap) == 0 {
		return 0, false
	}
	status := httpStatusOf(err)
	if status == 0 {
		return 0, false
	}
	if code, ok := statusMap[status]; ok {
		return code, true
	}
	return 0, false
}

// httpStatusOf extracts an HTTP status from an error if it exposes one.
func httpStatusOf(err error) int {
	if e, ok := err.(HTTPStatusError); ok {
		return e.HTTPStatus()
	}
	return 0
}
