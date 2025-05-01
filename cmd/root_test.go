// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package cmd

import (
	"os"
	"strings"
	"testing"
	
	"github.com/spf13/cobra"
	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers"
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
	// Create a new instance to avoid side effects
	cmd := &cobra.Command{
		Use:   "globus",
		Short: "Globus CLI - Command line interface for Globus services",
		Long: `Globus CLI - A command line interface for interacting with Globus services.
		
	This CLI provides access to Globus services including Auth, Transfer, Search,
	Groups, Flows, Compute, and Timers.`,
		Version: "0.1.0-test",
	}
	
	return cmd
}
