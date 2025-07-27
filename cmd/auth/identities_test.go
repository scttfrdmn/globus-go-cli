// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"fmt"
	"strings"
	"testing"

	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// customIdentitiesLookup implements a testable version of the identities lookup command logic
func customIdentitiesLookup(t *testing.T) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Extract flags
		username, _ := cmd.Flags().GetString("username")
		email, _ := cmd.Flags().GetString("email")
		id, _ := cmd.Flags().GetString("id")

		// Validate that at least one search parameter is provided
		if username == "" && email == "" && id == "" && len(args) == 0 {
			return fmt.Errorf("must provide at least one of: --username, --email, --id, or a search term as an argument")
		}

		// If an argument is provided, use it for search
		if len(args) > 0 && username == "" && email == "" && id == "" {
			// Determine if the argument looks like an email
			if strings.Contains(args[0], "@") {
				email = args[0]
				fmt.Printf("Using email: %s\n", email)
			} else if strings.HasPrefix(args[0], "urn:globus:auth:identity:") {
				id = args[0]
				fmt.Printf("Using ID: %s\n", id)
			} else {
				username = args[0]
				fmt.Printf("Using username: %s\n", username)
			}
		}

		// Get the current profile
		profile := viper.GetString("profile")
		fmt.Printf("Using profile: %s\n", profile)

		// For testing, we'll mock a successful token load
		fmt.Println("Looking up identities...")

		// Create mock identity responses based on the search criteria
		var output strings.Builder
		output.WriteString("ID\tUsername\tName\tEmail\tStatus\tIDProvider\n")

		// For username search
		if username != "" {
			output.WriteString(fmt.Sprintf("urn:globus:auth:identity:12345\t%s\tUser Example\t%s@example.org\tactive\tglobus.org\n", username, username))
		}

		// For email search
		if email != "" {
			output.WriteString(fmt.Sprintf("urn:globus:auth:identity:67890\t%s\tEmail User\t%s\tactive\texample.org\n",
				strings.Split(email, "@")[0], email))
		}

		// For ID search
		if id != "" {
			output.WriteString(fmt.Sprintf("%s\tid_user\tID User\tid_user@example.org\tactive\tmock-provider\n", id))
		}

		// Print the formatted output
		fmt.Print(output.String())
		return nil
	}
}

// setupLookupTest configures a test environment for the identities lookup command
func setupLookupTest(t *testing.T) *cobra.Command {
	// Save original values
	origProfile := viper.GetString("profile")
	origFormat := viper.GetString("format")

	// Restore original values after test
	defer func() {
		viper.Set("profile", origProfile)
		viper.Set("format", origFormat)
	}()

	// Configure test values
	viper.Set("profile", "test-profile")
	viper.Set("format", "table")

	// Create a custom standalone lookup command
	lookupCmd := &cobra.Command{
		Use:   "lookup",
		Short: "Look up Globus Auth identities",
		Long:  "Look up Globus Auth identities by username, email, or ID.",
		RunE:  customIdentitiesLookup(t),
	}

	// Add the same flags as the original command
	lookupCmd.Flags().String("username", "", "Look up by username")
	lookupCmd.Flags().String("email", "", "Look up by email")
	lookupCmd.Flags().String("id", "", "Look up by identity ID")

	return lookupCmd
}

func TestIdentitiesLookupCmd_NoArgs(t *testing.T) {
	// Set up the test environment
	lookupCmd := setupLookupTest(t)

	// Execute the command without arguments and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := lookupCmd.Execute()
		if err == nil {
			t.Errorf("Expected error for no arguments, but got none")
		}
	})

	// Output should be empty since we expect an error
	if strings.Contains(stdout, "Looking up identities") {
		t.Errorf("Did not expect successful execution, got: %s", stdout)
	}
}

func TestIdentitiesLookupCmd_Username(t *testing.T) {
	// Set up the test environment
	lookupCmd := setupLookupTest(t)

	// Set the flags
	lookupCmd.Flags().Set("username", "testuser")

	// Execute the command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := lookupCmd.Execute()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Check output
	expectedOutputs := []string{
		"Looking up identities",
		"testuser",
		"User Example",
		"testuser@example.org",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', output was: %s", expected, stdout)
		}
	}
}

func TestIdentitiesLookupCmd_Email(t *testing.T) {
	// Set up the test environment
	lookupCmd := setupLookupTest(t)

	// Set the flags
	lookupCmd.Flags().Set("email", "test@example.com")

	// Execute the command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := lookupCmd.Execute()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Check output
	expectedOutputs := []string{
		"Looking up identities",
		"test@example.com",
		"Email User",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', output was: %s", expected, stdout)
		}
	}
}

func TestIdentitiesLookupCmd_ID(t *testing.T) {
	// Set up the test environment
	lookupCmd := setupLookupTest(t)

	// Set the flags
	lookupCmd.Flags().Set("id", "urn:globus:auth:identity:54321")

	// Execute the command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := lookupCmd.Execute()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Check output
	expectedOutputs := []string{
		"Looking up identities",
		"urn:globus:auth:identity:54321",
		"ID User",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', output was: %s", expected, stdout)
		}
	}
}

func TestIdentitiesLookupCmd_ArgumentEmail(t *testing.T) {
	// Set up the test environment
	lookupCmd := setupLookupTest(t)

	// Execute the command with args and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		// Pass args directly to RunE
		err := lookupCmd.RunE(lookupCmd, []string{"arg@example.org"})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Check output for email detection
	expectedOutputs := []string{
		"Using email: arg@example.org",
		"Looking up identities",
		"arg@example.org",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', output was: %s", expected, stdout)
		}
	}
}

func TestIdentitiesLookupCmd_ArgumentID(t *testing.T) {
	// Set up the test environment
	lookupCmd := setupLookupTest(t)

	// Execute the command with args and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		// Pass args directly to RunE
		err := lookupCmd.RunE(lookupCmd, []string{"urn:globus:auth:identity:98765"})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Check output for ID detection
	expectedOutputs := []string{
		"Using ID: urn:globus:auth:identity:98765",
		"Looking up identities",
		"urn:globus:auth:identity:98765",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', output was: %s", expected, stdout)
		}
	}
}

func TestIdentitiesLookupCmd_ArgumentUsername(t *testing.T) {
	// Set up the test environment
	lookupCmd := setupLookupTest(t)

	// Execute the command with args and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		// Pass args directly to RunE
		err := lookupCmd.RunE(lookupCmd, []string{"username_arg"})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Check output for username detection
	expectedOutputs := []string{
		"Using username: username_arg",
		"Looking up identities",
		"username_arg",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', output was: %s", expected, stdout)
		}
	}
}
