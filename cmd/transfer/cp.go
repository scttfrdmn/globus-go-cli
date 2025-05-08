// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
)

var (
	transferRecursive    bool
	transferSync         int
	transferPreserveTime bool
	transferVerify       bool
	transferLabel        string
	transferWait         bool
	transferDryRun       bool
	transferDeadline     string
)

// CpCmd returns the cp command
func CpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cp SOURCE_ENDPOINT:SOURCE_PATH DEST_ENDPOINT:DEST_PATH",
		Short: "Transfer files between Globus endpoints",
		Long: `Transfer files between Globus endpoints.

This command submits a transfer task to copy files from a source endpoint to a
destination endpoint. The transfer runs asynchronously, and the command returns
a task ID that can be used to monitor the transfer.

Examples:
  globus transfer cp ddb59aef-6d04-11e5-ba46-22000b92c6ec:/path/file.txt ddb59af0-6d04-11e5-ba46-22000b92c6ec:/path/
  globus transfer cp --recursive ddb59aef-6d04-11e5-ba46-22000b92c6ec:/path/folder/ ddb59af0-6d04-11e5-ba46-22000b92c6ec:/dest/`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse source endpoint and path
			sourceEndpointID, sourcePath := parseEndpointAndPath(args[0])
			
			// Parse destination endpoint and path
			destEndpointID, destPath := parseEndpointAndPath(args[1])
			
			return transferFiles(cmd, sourceEndpointID, sourcePath, destEndpointID, destPath)
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&transferRecursive, "recursive", "r", false, "Transfer directories recursively")
	cmd.Flags().IntVar(&transferSync, "sync-level", 0, "Synchronization level (0-3)")
	cmd.Flags().BoolVar(&transferPreserveTime, "preserve-timestamp", false, "Preserve file timestamps")
	cmd.Flags().BoolVar(&transferVerify, "verify", false, "Verify file integrity after transfer")
	cmd.Flags().StringVar(&transferLabel, "label", "", "Label for the transfer task")
	cmd.Flags().BoolVar(&transferWait, "wait", false, "Wait for the transfer to complete")
	cmd.Flags().BoolVar(&transferDryRun, "dry-run", false, "Don't actually perform the transfer (test only)")
	cmd.Flags().StringVar(&transferDeadline, "deadline", "", "Transfer deadline (YYYY-MM-DD)")

	return cmd
}

// transferFiles transfers files between endpoints
func transferFiles(cmd *cobra.Command, sourceEndpointID, sourcePath, destEndpointID, destPath string) error {
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

	// Load client configuration - not used with direct client initialization in v0.9.10
	// We still load it for future use cases
	_, err = config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create a simple static token authorizer for v0.9.10
	tokenAuthorizer := authorizers.NewStaticTokenAuthorizer(tokenInfo.AccessToken)
	
	// Create a core authorizer adapter for v0.9.10 compatibility
	coreAuthorizer := authorizers.ToCore(tokenAuthorizer)

	// Create transfer client with v0.9.10 compatible options
	transferOptions := []transfer.Option{
		transfer.WithAuthorizer(coreAuthorizer),
	}
	
	transferClient, err := transfer.NewClient(transferOptions...)
	if err != nil {
		return fmt.Errorf("failed to create transfer client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parse deadline if provided
	var deadline *time.Time
	if transferDeadline != "" {
		parsedDeadline, err := time.Parse("2006-01-02", transferDeadline)
		if err != nil {
			return fmt.Errorf("invalid deadline format, use YYYY-MM-DD: %w", err)
		}
		deadline = &parsedDeadline
	}

	// Create transfer options map
	optionsMap := map[string]interface{}{
		"recursive":      transferRecursive,
		"sync_level":     transferSync,
		"preserve_mtime": transferPreserveTime,
		"verify_checksum": transferVerify,
	}
	
	if deadline != nil {
		optionsMap["deadline"] = deadline
	}

	// Show transfer details and confirm if not in dry run mode
	if !transferDryRun {
		fmt.Println("Transfer Details:")
		fmt.Printf("  Source:      %s:%s\n", sourceEndpointID, sourcePath)
		fmt.Printf("  Destination: %s:%s\n", destEndpointID, destPath)
		fmt.Printf("  Recursive:   %t\n", transferRecursive)
		fmt.Printf("  Sync Level:  %d\n", transferSync)
		
		confirm := promptui.Prompt{
			Label:     "Proceed with transfer",
			IsConfirm: true,
		}

		result, err := confirm.Run()
		if err != nil || strings.ToLower(result) != "y" {
			fmt.Println("Transfer canceled.")
			return nil
		}
	}

	// Start spinner for submission
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Submitting transfer task..."
	s.Start()

	// Submit the transfer
	taskResponse, err := transferClient.SubmitTransfer(
		ctx,
		sourceEndpointID, sourcePath,
		destEndpointID, destPath,
		transferLabel,
		optionsMap,
	)
	s.Stop()
	
	if err != nil {
		return fmt.Errorf("failed to submit transfer: %w", err)
	}

	// Display task information
	fmt.Printf("Task ID: %s\n", taskResponse.TaskID)
	fmt.Printf("Task submitted successfully. Run 'globus transfer task show %s' to check status.\n", taskResponse.TaskID)

	// If wait flag is specified, wait for the task to complete
	if transferWait {
		fmt.Println("Waiting for transfer to complete...")
		return waitForTask(cmd, taskResponse.TaskID, 1800) // 30 minutes default timeout
	}

	return nil
}