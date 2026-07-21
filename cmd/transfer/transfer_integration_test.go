// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

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
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
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

	// Create a v4 auth client (Basic auth) and obtain a transfer-scoped token
	// via the client_credentials grant.
	authClient, err := auth.NewClient(ctx, &core.Config{
		Authorizer: authorizers.NewBasicAuthAuthorizer(creds.ClientID, creds.ClientSecret),
	})
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	tokenResp, err := authClient.ClientCredentialsTokens(
		ctx,
		creds.ClientID, creds.ClientSecret,
		[]string{
			"urn:globus:auth:scope:transfer.api.globus.org:all",
			"openid", "email", "profile",
		},
	)
	if err != nil {
		t.Fatalf("Failed to get token for transfer operations: %v", err)
	}

	// Create a transfer client authorized with the access token.
	transferClient, err := transfer.NewClient(ctx, &core.Config{
		Authorizer: authorizers.NewAccessTokenAuthorizer(tokenResp.AccessToken),
	})
	if err != nil {
		t.Fatalf("Failed to create transfer client: %v", err)
	}

	t.Run("TestEndpointList", func(t *testing.T) {
		// Test endpoint listing
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Search endpoints with a small page size.
		searchOptions := &transfer.EndpointSearchOptions{
			FilterScope: "my-endpoints",
			Limit:       5,
		}
		endpoints, err := transferClient.EndpointSearch(ctx, searchOptions)
		if err != nil {
			t.Fatalf("Failed to search endpoints: %v", err)
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
		listing, err := transferClient.ListDirectory(ctx, creds.SourceEndpoint, creds.SourcePath, &transfer.ListDirectoryOptions{})
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
		_ = filepath.Join(testDir, "test-file-0") // Test file created but not used in this integration test

		// Test file transfer
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Longer timeout for transfers
		defer cancel()

		// Submit a transfer task
		sourcePath := fmt.Sprintf("%s/test-transfer-file.txt", creds.SourcePath)
		destPath := fmt.Sprintf("%s/test-transfer-file-%d.txt", creds.DestinationPath, time.Now().Unix())

		// A submission ID is required for the v4 transfer submit.
		submissionID, err := transferClient.GetSubmissionID(ctx)
		if err != nil {
			t.Fatalf("Failed to get submission ID: %v", err)
		}

		task, err := transferClient.SubmitTransfer(ctx, &transfer.Transfer{
			DATA_TYPE:           "transfer",
			SubmissionID:        submissionID,
			SourceEndpoint:      creds.SourceEndpoint,
			DestinationEndpoint: creds.DestinationEndpoint,
			Label:               "CLI Integration Test",
			VerifyChecksum:      true,
			Items: []transfer.TransferItem{
				{
					DATA_TYPE:       "transfer_item",
					SourcePath:      sourcePath,
					DestinationPath: destPath,
				},
			},
		})
		if err != nil {
			t.Fatalf("Failed to submit transfer task: %v", err)
		}

		// Verify we got a task ID
		if task.TaskID == "" {
			t.Fatal("Empty task ID received")
		}

		t.Logf("Successfully submitted transfer task with ID: %s", task.TaskID)
		t.Logf("  From: %s:%s", creds.SourceEndpoint, sourcePath)
		t.Logf("  To:   %s:%s", creds.DestinationEndpoint, destPath)

		// Note: In a real test, we might want to poll for task completion,
		// but that would make the test too long-running. Instead, we just
		// verify that the task was successfully submitted.
	})
}
