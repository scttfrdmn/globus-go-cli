// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers"
	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	// Test that the root command runs without errors
	stdout, stderr := testhelpers.CaptureOutput(func() {
		// Set args to just the program name to simulate running without args
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"globus"}

		// Execute with a new rootCmd instance for testing
		cmd := getRootCommandForTesting()
		cmd.SetArgs([]string{"--help"})
		cmd.Execute()
	})

	// Check that the output contains expected content
	if stdout == "" {
		t.Error("Expected help output, but stdout was empty")
	}

	// Check for expected content in the help output
	expectedPhrases := []string{
		"Globus CLI",
		"Available Commands",
		"help",
		"Flags:",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(stdout, phrase) {
			t.Errorf("Expected help output to contain '%s', but it didn't", phrase)
		}
	}

	// Stderr should be empty for a successful help command
	if stderr != "" {
		t.Errorf("Expected empty stderr, but got: %s", stderr)
	}
}

// getRootCommandForTesting creates a new instance of the root command for testing
func getRootCommandForTesting() *cobra.Command {
	// Use the actual root command to ensure subcommands are included
	cmd := Execute()

	// Override version for testing
	cmd.Version = "0.1.0-test"

	return cmd
}
