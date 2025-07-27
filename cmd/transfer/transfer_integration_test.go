// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

//go:build integration
// +build integration

package transfer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
)

func TestTransferIntegration(t *testing.T) {
	// Load test credentials and skip if not available
	creds := testhelpers.SkipIfNoCredentials(t)

	// Set up test environment
	homeDir, err := os.MkdirTemp("", "globus-cli-test-home-")
	if err != nil {
		t.Fatalf("Failed to create temp home directory: %v", err)
	}
	defer os.RemoveAll(homeDir)

	// Set HOME environment variable to the test directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", origHome)

	// Create test token directory
	tokenDir := filepath.Join(homeDir, ".globus-cli", "tokens")
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		t.Fatalf("Failed to create token directory: %v", err)
	}

	// Get authentication token for transfer operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create auth client with test credentials
	authClient, err := auth.NewClient(
		auth.WithClientCredentials(creds.ClientID, creds.ClientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	// Get a token with transfer scope
	tokenResp, err := authClient.GetClientCredentialsToken(
		ctx,
		[]string{
			"urn:globus:auth:scope:transfer.api.globus.org:all",
			"openid",
			"email",
			"profile",
		},
	)
	if err != nil {
		t.Fatalf("Failed to get token for transfer operations: %v", err)
	}

	// Create transfer client using the token
	transferClient, err := transfer.NewClient(
		transfer.WithBearerToken(tokenResp.AccessToken),
	)
	if err != nil {
		t.Fatalf("Failed to create transfer client: %v", err)
	}

	t.Run("TestEndpointList", func(t *testing.T) {
		// Test endpoint listing
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get endpoints with a filter
		listOptions := &transfer.ListEndpointsOptions{
			Limit: 5,
		}
		endpoints, err := transferClient.ListEndpoints(ctx, listOptions)
		if err != nil {
			t.Fatalf("Failed to list endpoints: %v", err)
		}

		// Verify we got results
		if len(endpoints.Data) == 0 {
			t.Log("No endpoints returned, but request was successful")
		} else {
			// Log the endpoints found
			t.Logf("Found %d endpoints", len(endpoints.Data))
			for i, ep := range endpoints.Data {
				if i < 3 { // Limit output to first 3 endpoints
					t.Logf("Endpoint: %s (%s)", ep.DisplayName, ep.ID)
				}
			}
		}
	})

	t.Run("TestGetEndpoint", func(t *testing.T) {
		// Skip if no source endpoint configured
		if creds.SourceEndpoint == "" {
			t.Skip("Source endpoint not configured in .env.test")
		}

		// Test getting a specific endpoint
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get the source endpoint
		endpoint, err := transferClient.GetEndpoint(ctx, creds.SourceEndpoint)
		if err != nil {
			t.Fatalf("Failed to get endpoint %s: %v", creds.SourceEndpoint, err)
		}

		// Verify endpoint data
		if endpoint.ID != creds.SourceEndpoint {
			t.Fatalf("Expected endpoint ID %s, got %s", creds.SourceEndpoint, endpoint.ID)
		}

		t.Logf("Successfully retrieved endpoint: %s (%s)", endpoint.DisplayName, endpoint.ID)
	})

	t.Run("TestListDirectory", func(t *testing.T) {
		// Skip if no source endpoint or path configured
		if creds.SourceEndpoint == "" || creds.SourcePath == "" {
			t.Skip("Source endpoint and path not configured in .env.test")
		}

		// Test listing a directory
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// List the source directory
		listOptions := &transfer.ListDirectoryOptions{
			Path: creds.SourcePath,
		}
		listing, err := transferClient.ListDirectory(ctx, creds.SourceEndpoint, listOptions)
		if err != nil {
			t.Fatalf("Failed to list directory %s on endpoint %s: %v",
				creds.SourcePath, creds.SourceEndpoint, err)
		}

		// Verify listing data
		t.Logf("Listed directory %s on endpoint %s", listing.Path, creds.SourceEndpoint)
		if len(listing.Data) == 0 {
			t.Log("Directory is empty")
		} else {
			t.Logf("Found %d items in directory", len(listing.Data))
			for i, item := range listing.Data {
				if i < 5 { // Limit output to first 5 items
					t.Logf("  %s (%s, %d bytes)", item.Name, item.Type, item.Size)
				}
			}
		}
	})

	t.Run("TestFileTransfer", func(t *testing.T) {
		// Skip if no transfer endpoints configured
		creds.RequireTransferEndpoints(t)

		// Also require source and destination paths
		if creds.SourcePath == "" || creds.DestinationPath == "" {
			t.Skip("Source or destination path not configured in .env.test")
		}

		// Create a test file to transfer
		testDir, cleanup := testhelpers.CreateTemporaryTestFiles(t, 1)
		defer cleanup()
		testFile := filepath.Join(testDir, "test-file-0")

		// Test file transfer
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Longer timeout for transfers
		defer cancel()

		// Submit a transfer task
		options := transfer.SubmitTransferOptions{
			Label:       "CLI Integration Test",
			SyncLevel:   0,
			VerifyWrite: true,
		}

		// Define a single transfer item
		transferItem := transfer.TransferItem{
			Source:      fmt.Sprintf("%s/test-transfer-file.txt", creds.SourcePath),
			Destination: fmt.Sprintf("%s/test-transfer-file-%d.txt", creds.DestinationPath, time.Now().Unix()),
		}

		// Submit the transfer
		task, err := transferClient.SubmitTransfer(
			ctx,
			creds.SourceEndpoint,
			creds.DestinationEndpoint,
			[]transfer.TransferItem{transferItem},
			&options,
		)
		if err != nil {
			t.Fatalf("Failed to submit transfer task: %v", err)
		}

		// Verify we got a task ID
		if task.TaskID == "" {
			t.Fatal("Empty task ID received")
		}

		t.Logf("Successfully submitted transfer task with ID: %s", task.TaskID)
		t.Logf("  From: %s:%s", creds.SourceEndpoint, transferItem.Source)
		t.Logf("  To:   %s:%s", creds.DestinationEndpoint, transferItem.Destination)

		// Note: In a real test, we might want to poll for task completion,
		// but that would make the test too long-running. Instead, we just
		// verify that the task was successfully submitted.
	})
}
