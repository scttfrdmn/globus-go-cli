// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package testhelpers

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CaptureOutput captures stdout and stderr during test execution
func CaptureOutput(f func()) (string, string) {
	// Save original outputs
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipes
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	// Redirect outputs
	os.Stdout = wOut
	os.Stderr = wErr

	// Run the function
	f()

	// Close writers
	wOut.Close()
	wErr.Close()

	// Read captured outputs
	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	// Restore original outputs
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return bufOut.String(), bufErr.String()
}

// CreateTempConfigDir creates a temporary config directory for testing
func CreateTempConfigDir(t *testing.T) string {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "globus-cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create subdirectories
	dirs := []string{"tokens", "profiles"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(tempDir, dir), 0700); err != nil {
			t.Fatalf("Failed to create %s directory: %v", dir, err)
		}
	}

	// Return the temp dir path
	return tempDir
}

// CleanupTempConfigDir removes a temporary config directory
func CleanupTempConfigDir(t *testing.T, dir string) {
	if dir == "" || strings.HasPrefix(filepath.Base(dir), "globus-cli-test-") == false {
		t.Fatalf("Invalid temp directory, refusing to delete: %s", dir)
	}

	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("Failed to remove temp directory: %v", err)
	}
}

// CreateTestTokenFile creates a test token file
func CreateTestTokenFile(t *testing.T, configDir, profile, content string) {
	tokenPath := filepath.Join(configDir, "tokens", profile+".json")
	if err := os.WriteFile(tokenPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test token file: %v", err)
	}
}

// CreateTestConfigFile creates a test config file
func CreateTestConfigFile(t *testing.T, configDir, content string) {
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
}
