// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package testhelpers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// SkipIfNoCredentials skips tests that require credentials if they're not available
func SkipIfNoCredentials(t *testing.T) *TestCredentials {
	creds := LoadTestCredentials(t)
	if creds.ClientID == "" || creds.ClientSecret == "" {
		t.Skip("Skipping test: No credentials available")
	}
	return creds
}

// RequireEnv checks if an environment variable is set and skips the test if not
func RequireEnv(t *testing.T, envVar string) string {
	value := os.Getenv(envVar)
	if value == "" {
		t.Skipf("Skipping test: %s environment variable not set", envVar)
	}
	return value
}

// CreateTemporaryTestFiles creates temporary test files for transfer tests
func CreateTemporaryTestFiles(t *testing.T, fileCount int) (dir string, cleanup func()) {
	dir, err := os.MkdirTemp("", "globus-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	for i := 0; i < fileCount; i++ {
		filename := filepath.Join(dir, fmt.Sprintf("test-file-%d", i))
		content := []byte(fmt.Sprintf("Test content for file %d", i))
		if err := os.WriteFile(filename, content, 0644); err != nil {
			os.RemoveAll(dir) // Clean up before failing
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	return dir, func() { os.RemoveAll(dir) }
}

// TimeoutContext returns a test context with a timeout
func TimeoutContext(t *testing.T) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// WithTimeout runs a function with a timeout and fails the test if it times out
func WithTimeout(t *testing.T, timeout time.Duration, f func()) {
	done := make(chan bool)
	go func() {
		f()
		done <- true
	}()

	select {
	case <-done:
		// Function completed before timeout
		return
	case <-time.After(timeout):
		t.Fatalf("Test timed out after %v", timeout)
	}
}

// TempTokenFile creates a temporary token file for testing
func TempTokenFile(t *testing.T, content []byte) (path string, cleanup func()) {
	// Create a temporary directory
	dir, err := os.MkdirTemp("", "globus-test-tokens-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create token file path
	filePath := filepath.Join(dir, "test-token.json")

	// Write the token file
	if err := os.WriteFile(filePath, content, 0600); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Failed to write temp token file: %v", err)
	}

	return filePath, func() { os.RemoveAll(dir) }
}
