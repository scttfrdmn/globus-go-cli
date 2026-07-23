// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

var (
	transferRecursive        bool
	transferSync             int
	transferSyncLevel        string
	transferPreserveTime     bool
	transferVerify           bool
	transferEncrypt          bool
	transferLabel            string
	transferWait             bool
	transferDryRun           bool
	transferDeadline         string
	transferSubmissionID     string
	transferNotify           []string
	transferSkipSourceErrors bool
	transferFailOnQuota      bool
	transferDeleteDestExtra  bool
	transferInclude          []string
	transferExclude          []string
	transferExternalChecksum string
	transferChecksumAlgo     string
	transferSourceLocalUser  string
	transferDestLocalUser    string
)

// CpCmd returns the cp command
func CpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer SOURCE_ENDPOINT:SOURCE_PATH DEST_ENDPOINT:DEST_PATH",
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
	cmd.Flags().StringVarP(&transferSyncLevel, "sync-level", "s", "", "Sync level: exists, size, mtime, checksum (or 0-3)")
	cmd.Flags().BoolVar(&transferPreserveTime, "preserve-timestamp", false, "Preserve file modification times")
	cmd.Flags().BoolVar(&transferPreserveTime, "preserve-mtime", false, "Preserve file modification times (alias of --preserve-timestamp)")
	cmd.Flags().BoolVar(&transferVerify, "verify-checksum", true, "Verify checksum after transfer")
	cmd.Flags().BoolVar(&transferEncrypt, "encrypt-data", false, "Encrypt data sent through the network")
	cmd.Flags().StringVar(&transferLabel, "label", "", "Label for the transfer task")
	cmd.Flags().StringVar(&transferSubmissionID, "submission-id", "", "Task submission ID for safe resubmission")
	cmd.Flags().StringSliceVar(&transferNotify, "notify", nil, "Comma-separated task events that notify by email (on, off, succeeded, failed, inactive)")
	cmd.Flags().BoolVar(&transferSkipSourceErrors, "skip-source-errors", false, "Skip source paths that hit permission-denied or not-found errors")
	cmd.Flags().BoolVar(&transferFailOnQuota, "fail-on-quota-errors", false, "Fail the task if any quota-exceeded errors are hit")
	cmd.Flags().BoolVar(&transferDeleteDestExtra, "delete-destination-extra", false, "Delete files in the destination not in the source (recursive mirroring)")
	cmd.Flags().StringArrayVar(&transferInclude, "include", nil, "Include files matching the given glob pattern in recursive transfers (repeatable)")
	cmd.Flags().StringArrayVar(&transferExclude, "exclude", nil, "Exclude files matching the given glob pattern in recursive transfers (repeatable)")
	cmd.Flags().StringVar(&transferExternalChecksum, "external-checksum", "", "External checksum to verify source file integrity")
	cmd.Flags().StringVar(&transferChecksumAlgo, "checksum-algorithm", "", "Algorithm for --external-checksum or --verify-checksum")
	cmd.Flags().StringVar(&transferSourceLocalUser, "source-local-user", "", "Local user to map to on the source (GCSv5 mapped collections)")
	cmd.Flags().StringVar(&transferDestLocalUser, "destination-local-user", "", "Local user to map to on the destination (GCSv5 mapped collections)")
	cmd.Flags().BoolVar(&transferWait, "wait", false, "Wait for the transfer to complete")
	cmd.Flags().BoolVar(&transferDryRun, "dry-run", false, "Don't actually perform the transfer (test only)")
	cmd.Flags().StringVar(&transferDeadline, "deadline", "", "Transfer deadline (YYYY-MM-DD)")

	return cmd
}

// transferFiles transfers files between endpoints
func transferFiles(cmd *cobra.Command, sourceEndpointID, sourcePath, destEndpointID, destPath string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Parse deadline if provided (validated here; formatted for the request below).
	var deadline string
	if transferDeadline != "" {
		if _, err := time.Parse("2006-01-02", transferDeadline); err != nil {
			return fmt.Errorf("invalid deadline format, use YYYY-MM-DD: %w", err)
		}
		deadline = transferDeadline
	}

	// Resolve sync level. Python accepts named levels (exists/size/mtime/checksum);
	// we also accept the raw ints 0-3 for backward compatibility. Unset leaves the
	// SDK default (0, omitted).
	if cmd.Flags().Changed("sync-level") {
		lvl, err := parseSyncLevel(transferSyncLevel)
		if err != nil {
			return err
		}
		transferSync = lvl
	}

	// Resolve notify events into the boolean notify_on_* fields.
	notifySucceeded, notifyFailed, notifyInactive, err := parseNotify(transferNotify)
	if err != nil {
		return err
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

	// Build the v4 transfer request. A submission ID minted from the service is
	// required for idempotent submission; honor an explicitly supplied one.
	submissionID := transferSubmissionID
	if submissionID == "" {
		submissionID, err = transferClient.GetSubmissionID(ctx)
		if err != nil {
			s.Stop()
			return fmt.Errorf("failed to get submission ID: %w", err)
		}
	}

	request := &transfer.Transfer{
		DATA_TYPE:              "transfer",
		SubmissionID:           submissionID,
		SourceEndpoint:         sourceEndpointID,
		DestinationEndpoint:    destEndpointID,
		Label:                  transferLabel,
		SyncLevel:              transferSync,
		VerifyChecksum:         transferVerify,
		PreserveTimestamp:      transferPreserveTime,
		EncryptData:            transferEncrypt,
		DeleteDestinationExtra: transferDeleteDestExtra,
		SkipSourceErrors:       transferSkipSourceErrors,
		FailOnQuotaErrors:      transferFailOnQuota,
		SourceLocalUser:        transferSourceLocalUser,
		DestinationLocalUser:   transferDestLocalUser,
		NotifyOnSucceeded:      notifySucceeded,
		NotifyOnFailed:         notifyFailed,
		NotifyOnInactive:       notifyInactive,
		Deadline:               deadline,
		FilterRules:            buildFilterRules(transferInclude, transferExclude),
		Items: []transfer.TransferItem{
			{
				DATA_TYPE:         "transfer_item",
				SourcePath:        sourcePath,
				DestinationPath:   destPath,
				Recursive:         transferRecursive,
				ExternalChecksum:  transferExternalChecksum,
				ChecksumAlgorithm: transferChecksumAlgo,
			},
		},
	}

	// Submit the transfer
	taskResponse, err := transferClient.SubmitTransfer(ctx, request)
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

// parseSyncLevel maps Python's named sync levels to the integer the Transfer
// API expects, while still accepting the raw integers 0-3 for compatibility.
func parseSyncLevel(v string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "exists", "0":
		return 0, nil
	case "size", "1":
		return 1, nil
	case "mtime", "2":
		return 2, nil
	case "checksum", "3":
		return 3, nil
	default:
		return 0, fmt.Errorf("invalid sync level %q (use exists, size, mtime, checksum, or 0-3)", v)
	}
}

// parseNotify converts Python's comma-separated --notify events into the
// notify_on_* booleans. "on" enables all, "off" disables all.
func parseNotify(events []string) (succeeded, failed, inactive bool, err error) {
	if len(events) == 0 {
		// Default: notify on succeeded and failed (Transfer API default).
		return true, true, true, nil
	}
	for _, e := range events {
		switch strings.ToLower(strings.TrimSpace(e)) {
		case "on":
			succeeded, failed, inactive = true, true, true
		case "off":
			succeeded, failed, inactive = false, false, false
		case "succeeded":
			succeeded = true
		case "failed":
			failed = true
		case "inactive":
			inactive = true
		case "":
			// ignore empty tokens
		default:
			return false, false, false, fmt.Errorf("invalid --notify value %q (use on, off, succeeded, failed, inactive)", e)
		}
	}
	return succeeded, failed, inactive, nil
}

// buildFilterRules converts --include/--exclude glob patterns into the Transfer
// API filter_rules list. Includes are emitted before excludes; the common
// idiom `--include "*.txt" --exclude "*"` therefore behaves as expected. Note:
// because cobra collects the two flags into separate slices, the exact
// interleaving of includes and excludes on the command line is not preserved.
func buildFilterRules(include, exclude []string) []transfer.FilterRule {
	if len(include) == 0 && len(exclude) == 0 {
		return nil
	}
	rules := make([]transfer.FilterRule, 0, len(include)+len(exclude))
	for _, p := range include {
		rules = append(rules, transfer.FilterRule{
			DATA_TYPE: "filter_rule", Method: "include", Name: p, Type: "file",
		})
	}
	for _, p := range exclude {
		rules = append(rules, transfer.FilterRule{
			DATA_TYPE: "filter_rule", Method: "exclude", Name: p, Type: "file",
		})
	}
	return rules
}
