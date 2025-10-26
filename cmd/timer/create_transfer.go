// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

// NOTE: This implementation uses direct HTTP calls to the Globus Timers v2 API
// because the Go SDK v3.65.0-1 does not yet support transfer timers.
// The SDK has FlowTimer helpers but no equivalent TransferTimer support.
//
// See SDK issue: https://github.com/scttfrdmn/globus-go-sdk/issues/16
//
// Once the SDK adds TransferTimer support (similar to Python SDK's TransferTimer
// class), this implementation should be refactored to use the SDK's helper methods.

package timer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	createTransferName              string
	createTransferSource            string
	createTransferDest              string
	createTransferRecursive         bool
	createTransferSyncLevel         int
	createTransferVerifyChecksum    bool
	createTransferPreserveTimestamp bool
	createTransferEncryptData       bool
	createTransferDelete            bool
	createTransferDeadline          string
	createTransferInterval          string
	createTransferStart             string
	createTransferStop              string
	createTransferInclude           []string
	createTransferExclude           []string
)

// CreateTransferCmd represents the timer create transfer command
var CreateTransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Create a recurring transfer timer",
	Long: `Create a timer to schedule recurring transfers between endpoints.

The timer will execute transfers at the specified interval.

Intervals use ISO 8601 duration format:
  P1D    - Every 1 day
  P1W    - Every 1 week
  PT1H   - Every 1 hour
  P1M    - Every 1 month

Examples:
  # Daily transfer
  globus timer create transfer \
    --name "Daily Backup" \
    --source SOURCE_EP:/data \
    --dest DEST_EP:/backup \
    --interval P1D

  # Weekly transfer with filters
  globus timer create transfer \
    --name "Weekly Sync" \
    --source SOURCE_EP:/data \
    --dest DEST_EP:/backup \
    --interval P1W \
    --include "*.txt" \
    --exclude "temp/*"

  # Directory mirroring (delete extra files)
  globus timer create transfer \
    --name "Mirror Directories" \
    --source SOURCE_EP:/data \
    --dest DEST_EP:/mirror \
    --interval P1D \
    --recursive \
    --delete`,
	RunE: runCreateTransferTimer,
}

func init() {
	CreateTransferCmd.Flags().StringVar(&createTransferName, "name", "", "Name for the timer (required)")
	CreateTransferCmd.Flags().StringVar(&createTransferSource, "source", "", "Source endpoint and path (ENDPOINT_ID:/path) (required)")
	CreateTransferCmd.Flags().StringVar(&createTransferDest, "dest", "", "Destination endpoint and path (ENDPOINT_ID:/path) (required)")
	CreateTransferCmd.Flags().StringVar(&createTransferInterval, "interval", "", "ISO 8601 interval (e.g., P1D, P1W, PT1H) (required)")
	CreateTransferCmd.Flags().StringVar(&createTransferStart, "start", "", "Start time (RFC3339 format)")
	CreateTransferCmd.Flags().StringVar(&createTransferStop, "stop", "", "Stop time (RFC3339 format)")
	CreateTransferCmd.Flags().BoolVar(&createTransferRecursive, "recursive", false, "Recursively transfer directories")
	CreateTransferCmd.Flags().IntVar(&createTransferSyncLevel, "sync-level", 0, "Sync level (0=exists, 1=size, 2=mtime, 3=checksum)")
	CreateTransferCmd.Flags().BoolVar(&createTransferVerifyChecksum, "verify-checksum", false, "Verify checksums after transfer")
	CreateTransferCmd.Flags().BoolVar(&createTransferPreserveTimestamp, "preserve-timestamp", false, "Preserve source file timestamps")
	CreateTransferCmd.Flags().BoolVar(&createTransferEncryptData, "encrypt-data", false, "Encrypt data during transfer")
	CreateTransferCmd.Flags().BoolVar(&createTransferDelete, "delete", false, "Delete extra files at destination (directory mirroring)")
	CreateTransferCmd.Flags().StringVar(&createTransferDeadline, "deadline", "", "Task deadline (RFC3339 format)")
	CreateTransferCmd.Flags().StringArrayVar(&createTransferInclude, "include", []string{}, "Include patterns")
	CreateTransferCmd.Flags().StringArrayVar(&createTransferExclude, "exclude", []string{}, "Exclude patterns")

	CreateTransferCmd.MarkFlagRequired("name")
	CreateTransferCmd.MarkFlagRequired("source")
	CreateTransferCmd.MarkFlagRequired("dest")
	CreateTransferCmd.MarkFlagRequired("interval")
}

func runCreateTransferTimer(cmd *cobra.Command, args []string) error {
	// Parse source and dest
	sourceParts := strings.SplitN(createTransferSource, ":", 2)
	if len(sourceParts) != 2 {
		return fmt.Errorf("source must be in format ENDPOINT_ID:/path")
	}
	sourceEndpoint := sourceParts[0]
	sourcePath := sourceParts[1]

	destParts := strings.SplitN(createTransferDest, ":", 2)
	if len(destParts) != 2 {
		return fmt.Errorf("dest must be in format ENDPOINT_ID:/path")
	}
	destEndpoint := destParts[0]
	destPath := destParts[1]

	// Get current profile
	profile := viper.GetString("profile")

	// Load token
	tokenInfo, err := authcmd.LoadToken(profile)
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Check if token is valid
	if !authcmd.IsTokenValid(tokenInfo) {
		return fmt.Errorf("token is expired, please login again")
	}

	// Load client configuration
	_, err = config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Build transfer task document
	transferItem := map[string]interface{}{
		"source_path":      sourcePath,
		"destination_path": destPath,
		"recursive":        createTransferRecursive,
	}

	transferBody := map[string]interface{}{
		"DATA_TYPE":             "transfer",
		"source_endpoint":       sourceEndpoint,
		"destination_endpoint":  destEndpoint,
		"DATA":                  []interface{}{transferItem},
		"sync_level":            createTransferSyncLevel,
		"verify_checksum":       createTransferVerifyChecksum,
		"preserve_timestamp":    createTransferPreserveTimestamp,
		"encrypt_data":          createTransferEncryptData,
		"delete_destination_extra": createTransferDelete,
	}

	// Add deadline if specified
	if createTransferDeadline != "" {
		transferBody["deadline"] = createTransferDeadline
	}

	// Add filter rules if specified
	if len(createTransferInclude) > 0 || len(createTransferExclude) > 0 {
		filterRules := []map[string]string{}
		for _, pattern := range createTransferInclude {
			filterRules = append(filterRules, map[string]string{
				"method": "include",
				"name":   pattern,
			})
		}
		for _, pattern := range createTransferExclude {
			filterRules = append(filterRules, map[string]string{
				"method": "exclude",
				"name":   pattern,
			})
		}
		transferBody["filter_rules"] = filterRules
	}

	// Build schedule
	schedule := map[string]interface{}{
		"type":     "recurring",
		"interval": createTransferInterval,
	}

	// Add start time if specified
	if createTransferStart != "" {
		startTime, err := time.Parse(time.RFC3339, createTransferStart)
		if err != nil {
			return fmt.Errorf("invalid start time format (use RFC3339): %w", err)
		}
		schedule["start_time"] = startTime.Format(time.RFC3339)
	}

	// Add end time if specified
	if createTransferStop != "" {
		endTime, err := time.Parse(time.RFC3339, createTransferStop)
		if err != nil {
			return fmt.Errorf("invalid stop time format (use RFC3339): %w", err)
		}
		schedule["end_time"] = endTime.Format(time.RFC3339)
	}

	// Build v2 transfer timer request
	timerRequest := map[string]interface{}{
		"name":     createTransferName,
		"timer_type": "transfer",
		"schedule": schedule,
		"body":     transferBody,
	}

	// Marshal request to JSON
	requestBody, err := json.Marshal(timerRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal timer request: %w", err)
	}

	// Create HTTP request to v2 timer endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiURL := "https://timer.automate.globus.org/v2/timer"
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+tokenInfo.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create timer: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(responseBody))
	}

	// Parse response
	var timerResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &timerResponse); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Transfer timer created successfully!\n\n")
	if timerID, ok := timerResponse["id"].(string); ok {
		fmt.Fprintf(os.Stdout, "Timer ID:    %s\n", timerID)
	}
	fmt.Fprintf(os.Stdout, "Name:        %s\n", createTransferName)
	fmt.Fprintf(os.Stdout, "Interval:    %s\n", createTransferInterval)
	fmt.Fprintf(os.Stdout, "Source:      %s:%s\n", sourceEndpoint, sourcePath)
	fmt.Fprintf(os.Stdout, "Destination: %s:%s\n", destEndpoint, destPath)

	return nil
}
