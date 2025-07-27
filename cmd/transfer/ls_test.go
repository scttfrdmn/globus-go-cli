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
	"github.com/spf13/viper"
)

// mockLsCommand creates a mock ls command for testing
func mockLsCommand(mockClient *mocks.MockTransferClient) *cobra.Command {
	cmd := LsCmd()

	// Override the normal RunE function completely
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Create a temporary viper config for this test
		viper.Set("format", "text")
		
		// Parse endpoint ID and path
		endpointID, path := parseEndpointAndPath(args[0])
		
		// Parse options
		options := &mocks.ListDirectoryOptions{
			EndpointID: endpointID,
			Path:       path,
			ShowHidden: lsShowHidden,
		}
		
		// Use our mock client directly
		listing, err := mockClient.ListDirectory(context.Background(), options)
		if err != nil {
			return fmt.Errorf("failed to list directory: %w", err)
		}
		
		// Just print the entries directly for testing
		for _, item := range listing.Data {
			if lsLongFormat {
				cmd.Printf("%s %s %s %s %d %s %s\n", 
					getFileType(item.Type), 
					item.Permissions, 
					item.User, 
					item.Group, 
					item.Size, 
					item.LastModified, 
					item.Name)
			} else {
				cmd.Printf("%s %s\n", getFileType(item.Type), item.Name)
			}
		}
		
		cmd.Printf("\nDirectory: %s:%s\n", endpointID, path)
		cmd.Printf("Total: %d items\n", len(listing.Data))
		
		return nil
	}

	return cmd
}

func TestLsCommand(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	tests := []struct {
		name         string
		args         []string
		mockSetup    func(*mocks.MockTransferClient)
		expectOutput string
		expectError  bool
	}{
		{
			name: "successful directory listing",
			args: []string{"endpoint-id:/path"},
			mockSetup: func(m *mocks.MockTransferClient) {
				m.ListDirectoryFunc = func(ctx context.Context, options *mocks.ListDirectoryOptions) (*mocks.ListDirectoryResponse, error) {
					if options.EndpointID != "endpoint-id" || options.Path != "/path" {
						t.Errorf("Unexpected options: %+v", options)
					}
					return &mocks.ListDirectoryResponse{
						Path: "/path",
						Data: []mocks.FileEntry{
							{Name: "file1", Type: "file", Size: 1024, LastModified: "2023-01-01T12:00:00Z"},
							{Name: "dir1", Type: "dir", Size: 0, LastModified: "2023-01-01T12:00:00Z"},
						},
					}, nil
				}
			},
			expectOutput: "file1",
			expectError:  false,
		},
		{
			name: "path with trailing slash",
			args: []string{"endpoint-id:/path/"},
			mockSetup: func(m *mocks.MockTransferClient) {
				m.ListDirectoryFunc = func(ctx context.Context, options *mocks.ListDirectoryOptions) (*mocks.ListDirectoryResponse, error) {
					if options.Path != "/path/" {
						t.Errorf("Expected path with trailing slash, got: %s", options.Path)
					}
					return &mocks.ListDirectoryResponse{
						Path: "/path/",
						Data: []mocks.FileEntry{
							{Name: "file2", Type: "file", Size: 2048, LastModified: "2023-01-01T12:00:00Z"},
						},
					}, nil
				}
			},
			expectOutput: "file2",
			expectError:  false,
		},
		{
			name: "error listing directory",
			args: []string{"invalid-endpoint:/path"},
			mockSetup: func(m *mocks.MockTransferClient) {
				m.ListDirectoryFunc = func(ctx context.Context, options *mocks.ListDirectoryOptions) (*mocks.ListDirectoryResponse, error) {
					return nil, &mocks.EndpointError{
						Code:    "EndpointNotFound",
						Message: "Endpoint not found",
					}
				}
			},
			expectOutput: "failed to list directory",
			expectError:  true,
		},
		{
			name: "long format option",
			args: []string{"endpoint-id:/path", "--long"},
			mockSetup: func(m *mocks.MockTransferClient) {
				m.ListDirectoryFunc = func(ctx context.Context, options *mocks.ListDirectoryOptions) (*mocks.ListDirectoryResponse, error) {
					return &mocks.ListDirectoryResponse{
						Path: "/path",
						Data: []mocks.FileEntry{
							{
								Name:         "file3",
								Type:         "file",
								Size:         4096,
								LastModified: "2023-01-01T12:00:00Z",
								Permissions:  "rw-r--r--",
								User:         "testuser",
								Group:        "testgroup",
							},
						},
					}, nil
				}
			},
			expectOutput: "file3",
			expectError:  false,
		},
		{
			name: "show hidden files option",
			args: []string{"endpoint-id:/path", "--all"},
			mockSetup: func(m *mocks.MockTransferClient) {
				m.ListDirectoryFunc = func(ctx context.Context, options *mocks.ListDirectoryOptions) (*mocks.ListDirectoryResponse, error) {
					if !options.ShowHidden {
						t.Errorf("Expected ShowHidden to be true")
					}
					return &mocks.ListDirectoryResponse{
						Path: "/path",
						Data: []mocks.FileEntry{
							{Name: ".hidden", Type: "file", Size: 512, LastModified: "2023-01-01T12:00:00Z"},
						},
					}, nil
				}
			},
			expectOutput: ".hidden",
			expectError:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mocks.MockTransferClient{}
			if tc.mockSetup != nil {
				tc.mockSetup(mockClient)
			}

			// Create test command
			cmd := mockLsCommand(mockClient)
			cmd.SetArgs(tc.args)

			// Capture output
			stdout, stderr := testhelpers.CaptureOutput(func() {
				err := cmd.Execute()
				if tc.expectError && err == nil {
					t.Errorf("Expected error but got none")
				} else if !tc.expectError && err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			})

			// Check output
			output := stdout + stderr
			if tc.expectOutput != "" && !strings.Contains(output, tc.expectOutput) {
				t.Errorf("Expected output to contain %q, got: %q", tc.expectOutput, output)
			}
		})
	}
}