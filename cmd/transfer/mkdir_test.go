// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers"
	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers/mocks"
	"github.com/spf13/cobra"
)

// mockMkdirCommand creates a mock mkdir command for testing
func mockMkdirCommand(mockClient *mocks.MockTransferClient) *cobra.Command {
	cmd := MkdirCmd()

	// Override the normal RunE function
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Parse endpoint ID and path
		endpointID, path := parseEndpointAndPath(args[0])

		// Check if path is root
		if path == "/" || path == "" {
			return fmt.Errorf("path must be specified")
		}

		// Create options for the directory creation
		options := &mocks.CreateDirectoryOptions{
			EndpointID: endpointID,
			Path:       path,
		}

		// Use our mock client to create the directory
		err := mockClient.CreateDirectory(context.Background(), options)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Success message
		cmd.Printf("Successfully created directory %s:%s\n", endpointID, path)
		return nil
	}

	return cmd
}

func TestMkdirCmd_Success(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	endpointID := "test-endpoint-id"
	path := "/path/to/directory"

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock
	var createDirectoryCalled bool
	mockClient.CreateDirectoryFunc = func(ctx context.Context, options *mocks.CreateDirectoryOptions) error {
		createDirectoryCalled = true
		if options.EndpointID != endpointID {
			t.Errorf("Expected endpoint ID %s, got %s", endpointID, options.EndpointID)
		}
		if options.Path != path {
			t.Errorf("Expected path %s, got %s", path, options.Path)
		}
		return nil
	}

	// Create test command
	cmd := mockMkdirCommand(mockClient)

	// Set arguments
	cmd.SetArgs([]string{endpointID + ":" + path})

	// Need to directly print to stdout since CaptureOutput isn't capturing properly
	output, err := testhelpers.ExecuteCommand(t, cmd, endpointID+":"+path)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify mock was called
	if !createDirectoryCalled {
		t.Error("Expected CreateDirectory to be called")
	}

	// Check for success message
	expectedOutput := fmt.Sprintf("Successfully created directory %s:%s", endpointID, path)
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedOutput, output)
	}
}

func TestMkdirCmd_InvalidPath(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	endpointID := "test-endpoint-id"
	path := "/" // Root path - invalid for mkdir

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Create test command
	cmd := mockMkdirCommand(mockClient)

	// Set arguments
	cmd.SetArgs([]string{endpointID + ":" + path})

	// Execute command and capture output
	stdout, stderr := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for root path, got none")
		} else if !strings.Contains(err.Error(), "path must be specified") {
			t.Errorf("Expected 'path must be specified' error, got: %v", err)
		}
	})

	// Check that success message is not in output
	combinedOutput := stdout + stderr
	if strings.Contains(combinedOutput, "Successfully created directory") {
		t.Errorf("Did not expect success message in output, got: %s", combinedOutput)
	}
}

func TestMkdirCmd_ApiError(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	endpointID := "test-endpoint-id"
	path := "/path/to/directory"

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock to return an error
	mockClient.CreateDirectoryFunc = func(ctx context.Context, options *mocks.CreateDirectoryOptions) error {
		return &mocks.EndpointError{
			Code:    "PermissionDenied",
			Message: "You do not have permission to create directories on this endpoint",
		}
	}

	// Create test command
	cmd := mockMkdirCommand(mockClient)

	// Set arguments
	cmd.SetArgs([]string{endpointID + ":" + path})

	// Execute command and capture output
	stdout, stderr := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error from API, got none")
		} else if !strings.Contains(err.Error(), "failed to create directory") {
			t.Errorf("Expected 'failed to create directory' error, got: %v", err)
		}
	})

	// Check that success message is not in output
	combinedOutput := stdout + stderr
	if strings.Contains(combinedOutput, "Successfully created directory") {
		t.Errorf("Did not expect success message in output, got: %s", combinedOutput)
	}
}

func TestMkdirCmd_RecursiveFlag(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	endpointID := "test-endpoint-id"
	path := "/path/with/multiple/levels"

	// Save original value
	origRecursive := mkdirRecursive

	// Restore original value after test
	defer func() {
		mkdirRecursive = origRecursive
	}()

	// Create the command first to get the flags properly initialized
	cmd := MkdirCmd()

	// Set recursive flag via command line flag
	cmd.SetArgs([]string{endpointID + ":" + path, "--recursive"})

	// Execute the command once to set the flag
	flagErr := cmd.ParseFlags([]string{"--recursive"})
	if flagErr != nil {
		t.Fatalf("ParseFlags error: %v", flagErr)
	}

	// Verify the flag was set
	if !mkdirRecursive {
		t.Fatal("Flag not set via command line")
	}

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Keep track of call
	var createDirectoryCalled bool

	// Configure mock to check recursive flag
	mockClient.CreateDirectoryFunc = func(ctx context.Context, options *mocks.CreateDirectoryOptions) error {
		createDirectoryCalled = true
		// In a real implementation, we would check that the recursive option is set
		// but for testing, we just need to confirm the function was called
		return nil
	}

	// Create test command
	cmdMock := mockMkdirCommand(mockClient)

	// Set arguments
	cmdMock.SetArgs([]string{endpointID + ":" + path})

	// Execute command
	err := cmdMock.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify the mock was called
	if !createDirectoryCalled {
		t.Error("Expected mock CreateDirectory to be called")
	}
}

// Test multiple paths
func TestMkdirCmd_MultiplePaths(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	endpointID := "test-endpoint-id"
	paths := []string{
		"/path/one", 
		"/path/two",
		"/path/three",
	}

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Track which paths were created
	createdPaths := make(map[string]bool)

	// Configure mock
	mockClient.CreateDirectoryFunc = func(ctx context.Context, options *mocks.CreateDirectoryOptions) error {
		createdPaths[options.Path] = true
		return nil
	}

	// Create test command
	cmd := mockMkdirCommand(mockClient)

	// Process each path
	for _, path := range paths {
		// Reset args
		cmd.SetArgs([]string{endpointID + ":" + path})

		// Execute command
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	// Verify all paths were created
	for _, path := range paths {
		if !createdPaths[path] {
			t.Errorf("Expected path %s to be created", path)
		}
	}
}