// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package testhelpers

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
)

// ExecuteCommand executes a cobra command for testing with the given arguments
// Returns the output as a string and any error that occurred
func ExecuteCommand(t *testing.T, root *cobra.Command, args ...string) (string, error) {
	t.Helper()

	// Create buffer to capture output
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)

	// Set args
	root.SetArgs(args)

	// Capture output from fmt.Print statements
	var output string
	stdout, stderr := CaptureOutput(func() {
		err := root.Execute()
		if err != nil {
			fmt.Fprintf(buf, "Error: %v\n", err)
		}
	})

	// Combine all output sources
	output = buf.String() + stdout + stderr

	// Execute command again for error
	root.SetArgs(args)
	err := root.Execute()

	// Return output and error
	return output, err
}

// ExecuteCommandWithConfig executes a cobra command for testing with a proper environment
// This sets up a complete test environment with config and tokens
func ExecuteCommandWithConfig(t *testing.T, root *cobra.Command, args ...string) (string, error) {
	t.Helper()

	// Setup test config with tokens
	_, cleanup := SetupTestConfig(t)
	defer cleanup()

	return ExecuteCommand(t, root, args...)
}

// ExecuteCommandWithToken executes a cobra command for testing with just a token file
// This is useful for tests that only need authentication but not full config
func ExecuteCommandWithToken(t *testing.T, root *cobra.Command, args ...string) (string, error) {
	t.Helper()

	// Setup test token
	_, cleanup := SetupTokenFile(t)
	defer cleanup()

	return ExecuteCommand(t, root, args...)
}

// ExecuteCommandWithCustomProfile executes a cobra command with a specific profile
func ExecuteCommandWithCustomProfile(t *testing.T, root *cobra.Command, profile string, args ...string) (string, error) {
	t.Helper()

	// Setup test config with tokens
	_, cleanup := SetupTestConfig(t)
	defer cleanup()

	// Add profile flag to args if it's not already there
	hasProfile := false
	for i, arg := range args {
		if arg == "--profile" || arg == "-p" {
			hasProfile = true
			// Make sure the next arg is our profile
			if i+1 < len(args) {
				args[i+1] = profile
			}
			break
		}
	}

	if !hasProfile {
		newArgs := append([]string{"--profile", profile}, args...)
		args = newArgs
	}

	return ExecuteCommand(t, root, args...)
}

// AssertContains checks if output contains expected string
func AssertContains(t *testing.T, output string, expected string) {
	t.Helper()
	if !bytes.Contains([]byte(output), []byte(expected)) {
		t.Errorf("Expected output to contain %q, got: %q", expected, output)
	}
}

// AssertNotContains checks if output does not contain unexpected string
func AssertNotContains(t *testing.T, output string, unexpected string) {
	t.Helper()
	if bytes.Contains([]byte(output), []byte(unexpected)) {
		t.Errorf("Expected output not to contain %q, but it does: %q", unexpected, output)
	}
}