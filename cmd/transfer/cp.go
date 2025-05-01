// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
)

var (
	// Options for transfer command
	syncLevelFlag      string
	labelFlag          string
	recursiveFlag      bool
	preserveMtimeFlag  bool
	verifyChecksumFlag bool
	encryptDataFlag    bool
	dryRunFlag         bool
	deadlineHoursFlag  int
	notifyFlag         string
	progressFlag       bool
	waitFlag           bool
)

// CpCmd returns the cp (transfer) command
func CpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cp SOURCE_ENDPOINT[:PATH] DEST_ENDPOINT[:PATH]",
		Short: "Transfer files between Globus endpoints",
		Long: `Transfer files between Globus endpoints.

This command initiates a file transfer between two Globus endpoints.
You can transfer a single file or entire directories (with --recursive).

Examples:
  # Transfer a single file
  globus transfer cp endpoint1:/path/file endpoint2:/path/

  # Transfer a directory recursively
  globus transfer cp endpoint1:/path/dir endpoint2:/path/ --recursive

  # Transfer with synchronization
  globus transfer cp endpoint1:/path/ endpoint2:/path/ --recursive --sync-level exists`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return transferFiles(cmd, args[0], args[1])
		},
	}

	// Add flags
	cmd.Flags().StringVar(&syncLevelFlag, "sync-level", "none", "Sync level (none, exists, size, mtime, checksum)")
	cmd.Flags().StringVar(&labelFlag, "label", "", "Label for the transfer")
	cmd.Flags().BoolVarP(&recursiveFlag, "recursive", "r", false, "Transfer directories recursively")
	cmd.Flags().BoolVar(&preserveMtimeFlag, "preserve-mtime", true, "Preserve file modification times")
	cmd.Flags().BoolVar(&verifyChecksumFlag, "verify-checksum", false, "Verify checksums")
	cmd.Flags().BoolVar(&encryptDataFlag, "encrypt", true, "Encrypt data on transfer")
	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Don't actually submit the transfer task")
	cmd.Flags().IntVar(&deadlineHoursFlag, "deadline-hours", 24, "Hours until the transfer expires")
	cmd.Flags().StringVar(&notifyFlag, "notify", "off", "Notification preference (off, succeeded, failed, inactive)")
	cmd.Flags().BoolVar(&progressFlag, "progress", true, "Show progress during transfer")
	cmd.Flags().BoolVar(&waitFlag, "wait", false, "Wait for the transfer to complete")

	return cmd
}

// transferFiles initiates a file transfer between two Globus endpoints
func transferFiles(cmd *cobra.Command, source, destination string) error {
	// Parse the source endpoint ID and path
	sourceEndpointID, sourcePath, err := parseEndpointPath(source)
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	// Parse the destination endpoint ID and path
	destEndpointID, destPath, err := parseEndpointPath(destination)
	if err != nil {
		return fmt.Errorf("invalid destination: %w", err)
	}

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
	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create SDK config
	sdkConfig := pkg.NewConfig().
		WithClientID(clientCfg.ClientID).
		WithClientSecret(clientCfg.ClientSecret)

	// Create transfer client
	transferClient := sdkConfig.NewTransferClient(tokenInfo.AccessToken)

	// Create context with timeout for submission
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Determine the transfer label
	if labelFlag == "" {
		// Generate a default label
		labelFlag = fmt.Sprintf("CLI Transfer %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	// Process the sync level
	syncLevel := getSyncLevelValue(syncLevelFlag)

	// Show info about the transfer
	fmt.Printf("Transfer details:\n")
	fmt.Printf("  Source:      %s:%s\n", sourceEndpointID, sourcePath)
	fmt.Printf("  Destination: %s:%s\n", destEndpointID, destPath)
	fmt.Printf("  Recursive:   %t\n", recursiveFlag)
	fmt.Printf("  Sync Level:  %s\n", syncLevelFlag)
	fmt.Printf("  Label:       %s\n", labelFlag)

	// If dry-run, don't actually submit the transfer
	if dryRunFlag {
		fmt.Println("\nDry run mode: transfer would be submitted with the above parameters.")
		return nil
	}

	// Confirm with the user
	fmt.Print("\nProceed with this transfer? [Y/n] ")
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) == "n" {
		fmt.Println("Transfer cancelled.")
		return nil
	}

	// Prepare the transfer
	var result pkg.TransferResponse

	if recursiveFlag {
		// Setup options for recursive transfer
		options := pkg.DefaultRecursiveTransferOptions()
		options.Label = labelFlag
		options.PreserveMtime = preserveMtimeFlag
		options.VerifyChecksum = verifyChecksumFlag
		options.EncryptData = encryptDataFlag
		options.DeadlineHours = deadlineHoursFlag
		options.NotifyOn = notifyFlag

		// Convert sync level string to the appropriate option
		switch syncLevelFlag {
		case "exists":
			options.Sync = true
		case "size", "mtime", "checksum":
			options.Sync = true
			options.SyncLevel = syncLevelFlag
		}

		// Setup progress callback if requested
		if progressFlag {
			spin := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
			spin.Prefix = "Scanning files... "
			spin.Start()

			options.ProgressCallback = func(current, total int64, message string) {
				if total > 0 {
					spin.Stop()
					fmt.Printf("\rScanning: Found %d files (%s)...", total, formatSize(current))
				}
			}
		}

		// Submit the recursive transfer
		fmt.Println("\nPreparing recursive transfer...")
		recursiveResult, err := transferClient.SubmitRecursiveTransfer(
			ctx,
			sourceEndpointID, sourcePath,
			destEndpointID, destPath,
			options,
		)
		if err != nil {
			return fmt.Errorf("failed to submit transfer: %w", err)
		}

		// Show transfer summary
		fmt.Printf("\nTransfer submitted with task ID: %s\n", recursiveResult.TaskID)
		fmt.Printf("Found %d files (%s) in %d directories\n",
			recursiveResult.TotalFiles,
			formatSize(recursiveResult.TotalSize),
			recursiveResult.Directories+recursiveResult.Subdirectories,
		)

		// Set result for status monitoring
		result.TaskID = recursiveResult.TaskID
	} else {
		// Setup options for regular transfer
		options := make(map[string]interface{})
		options["label"] = labelFlag
		options["sync_level"] = syncLevel
		options["verify_checksum"] = verifyChecksumFlag
		options["preserve_mtime"] = preserveMtimeFlag
		options["encrypt_data"] = encryptDataFlag
		options["deadline"] = time.Now().Add(time.Duration(deadlineHoursFlag) * time.Hour).Format(time.RFC3339)
		options["notify_on_succeeded"] = strings.Contains(notifyFlag, "succeeded")
		options["notify_on_failed"] = strings.Contains(notifyFlag, "failed")
		options["notify_on_inactive"] = strings.Contains(notifyFlag, "inactive")

		// Submit the regular transfer
		fmt.Println("\nSubmitting transfer...")
		transferResult, err := transferClient.SubmitTransfer(
			ctx,
			sourceEndpointID, sourcePath,
			destEndpointID, destPath,
			options,
		)
		if err != nil {
			return fmt.Errorf("failed to submit transfer: %w", err)
		}

		// Show transfer summary
		fmt.Printf("Transfer submitted with task ID: %s\n", transferResult.TaskID)

		// Set result for status monitoring
		result = *transferResult
	}

	// Save the task ID to a file for later status checks
	homeDir, err := os.UserHomeDir()
	if err == nil {
		taskFile := filepath.Join(homeDir, ".globus-cli", "last-task-id")
		os.WriteFile(taskFile, []byte(result.TaskID), 0600)
	}

	// Wait for the transfer to complete if requested
	if waitFlag {
		fmt.Println("\nWaiting for transfer to complete...")
		err := waitForTask(transferClient, result.TaskID, progressFlag)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("\nYou can check the status of this transfer with:")
		fmt.Printf("  globus transfer task show %s\n", result.TaskID)
	}

	return nil
}

// getSyncLevelValue converts a sync level string to the value expected by the API
func getSyncLevelValue(syncLevel string) string {
	switch strings.ToLower(syncLevel) {
	case "none":
		return "0"
	case "exists":
		return "1"
	case "size":
		return "2"
	case "mtime":
		return "3"
	case "checksum":
		return "4"
	default:
		return "0"
	}
}

// waitForTask waits for a transfer task to complete
func waitForTask(client *pkg.TransferClient, taskID string, showProgress bool) error {
	// Create context with a long timeout
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Hour)
	defer cancel()

	// Create a progress bar if requested
	var bar *progressbar.ProgressBar
	if showProgress {
		bar = progressbar.NewOptions(
			100,
			progressbar.OptionSetDescription("Transfer Progress:"),
			progressbar.OptionSetWidth(50),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionShowCount(),
			progressbar.OptionThrottle(100*time.Millisecond),
		)
	}

	// Wait for the task with periodic status checks
	checkInterval := 5 * time.Second
	maxInterval := 30 * time.Second
	
	for {
		// Check the task status
		task, err := client.GetTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("failed to get task status: %w", err)
		}

		// Update progress bar if requested
		if showProgress && bar != nil {
			percent := int(percentComplete(task.BytesTransferred, task.BytesExpected))
			bar.Set(percent)
			bar.Describe(fmt.Sprintf("Transferred: %s / %s",
				formatSize(task.BytesTransferred),
				formatSize(task.BytesExpected),
			))
		}

		// Check if the task is complete
		if task.Status == "SUCCEEDED" {
			if bar != nil {
				bar.Finish()
			}
			fmt.Printf("\nTransfer completed successfully!\n")
			fmt.Printf("Files transferred: %d\n", task.FilesTransferred)
			fmt.Printf("Directories created: %d\n", task.DirectoriesCreated)
			fmt.Printf("Total size: %s\n", formatSize(task.BytesTransferred))
			return nil
		} else if task.Status == "FAILED" {
			if bar != nil {
				bar.Clear()
			}
			return fmt.Errorf("transfer failed: %s", task.NiceStatusShortDescription)
		} else if task.Status != "ACTIVE" {
			if bar != nil {
				bar.Clear()
			}
			return fmt.Errorf("transfer in unexpected state: %s", task.Status)
		}

		// Increase check interval up to the maximum
		if checkInterval < maxInterval {
			checkInterval += time.Second
		}
		time.Sleep(checkInterval)
	}
}

// percentComplete calculates the percentage of a transfer completed
func percentComplete(current, total int64) float64 {
	if total <= 0 {
		return 0.0
	}
	return float64(current) / float64(total) * 100.0
}