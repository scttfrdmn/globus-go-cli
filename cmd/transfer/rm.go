// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

var (
	rmRecursive     bool
	rmForce         bool
	rmIgnoreMissing bool
	rmLabel         string
	rmDeadline      string
	rmLocalUser     string
	rmEnableGlobs   bool
	rmNotify        []string
)

// RmCmd returns the rm command
func RmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm ENDPOINT_ID:PATH",
		Short: "Remove a file or directory on an endpoint",
		Long: `Remove a file or directory on a Globus endpoint.

This command deletes a file or directory on the specified Globus endpoint.
If --recursive is specified, it will delete directories and their contents.

Examples:
  globus transfer rm ddb59aef-6d04-11e5-ba46-22000b92c6ec:/path/to/file
  globus transfer rm --recursive ddb59aef-6d04-11e5-ba46-22000b92c6ec:/path/to/directory`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse endpoint ID and path
			endpointID, path := parseEndpointAndPath(args[0])

			// Check that path is specified
			if path == "/" {
				return fmt.Errorf("path must be specified for rm command")
			}

			return removeItem(cmd, endpointID, path)
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&rmRecursive, "recursive", "r", false, "Remove directories and their contents recursively")
	cmd.Flags().BoolVarP(&rmForce, "force", "f", false, "Force removal without confirmation")
	cmd.Flags().BoolVar(&rmIgnoreMissing, "ignore-missing", false, "Do not error if the path does not exist")
	cmd.Flags().StringVar(&rmLabel, "label", "", "Set a label for this task")
	cmd.Flags().StringVar(&rmDeadline, "deadline", "", "Deadline for the task (YYYY-MM-DD)")
	cmd.Flags().StringVar(&rmLocalUser, "local-user", "", "Local user to map to (GCSv5 mapped collections)")
	cmd.Flags().BoolVar(&rmEnableGlobs, "enable-globs", false, "Interpret shell-style globs in the path")
	cmd.Flags().StringSliceVar(&rmNotify, "notify", nil, "Notification settings: any of on, off, succeeded, failed, inactive")

	return cmd
}

// removeItem removes a file or directory on an endpoint
func removeItem(cmd *cobra.Command, endpointID, path string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Check if we need to prompt for confirmation
	if !rmForce {
		// Get file/directory info
		options := &transfer.ListDirectoryOptions{}

		listing, err := transferClient.ListDirectory(ctx, endpointID, path, options)
		if err != nil {
			// If we can't get info, still prompt
			prompt := fmt.Sprintf("Are you sure you want to delete %s:%s?", endpointID, path)
			if !confirmAction(prompt) {
				fmt.Println("Operation canceled.")
				return nil
			}
		} else {
			// Check if it's a directory
			isDir := false
			for _, item := range listing.Data {
				if item.Type == "dir" && item.Name == "." {
					isDir = true
					break
				}
			}

			if isDir {
				if !rmRecursive {
					return fmt.Errorf("%s is a directory. Use --recursive to remove directories", path)
				}

				// Count items in the directory
				count := len(listing.Data)
				if count > 2 { // Accounting for "." and ".."
					prompt := fmt.Sprintf("Are you sure you want to delete directory %s:%s and all its contents (%d items)?",
						endpointID, path, count-2)
					if !confirmAction(prompt) {
						fmt.Println("Operation canceled.")
						return nil
					}
				}
			} else {
				// It's a file
				prompt := fmt.Sprintf("Are you sure you want to delete file %s:%s?", endpointID, path)
				if !confirmAction(prompt) {
					fmt.Println("Operation canceled.")
					return nil
				}
			}
		}
	}

	// Submit a delete task. The v4 SDK carries recursion on the Delete request
	// itself, and requires a submission ID minted from the service.
	submissionID, err := transferClient.GetSubmissionID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get submission ID: %w", err)
	}

	if rmDeadline != "" {
		if _, derr := time.Parse("2006-01-02", rmDeadline); derr != nil {
			return fmt.Errorf("invalid deadline format, use YYYY-MM-DD: %w", derr)
		}
	}
	notifySucceeded, notifyFailed, notifyInactive, nerr := parseNotify(rmNotify)
	if nerr != nil {
		return nerr
	}

	deleteRequest := &transfer.Delete{
		DATA_TYPE:         "delete",
		SubmissionID:      submissionID,
		Endpoint:          endpointID,
		Label:             rmLabel,
		Recursive:         rmRecursive,
		IgnoreMissing:     rmIgnoreMissing,
		InterpretGlob:     rmEnableGlobs,
		LocalUser:         rmLocalUser,
		Deadline:          rmDeadline,
		NotifyOnSucceeded: notifySucceeded,
		NotifyOnFailed:    notifyFailed,
		NotifyOnInactive:  notifyInactive,
		Items: []transfer.DeleteItem{
			{DATA_TYPE: "delete_item", Path: path},
		},
	}

	// Create a delete task
	taskResponse, err := transferClient.SubmitDelete(ctx, deleteRequest)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	fmt.Printf("Delete task submitted. Task ID: %s\n", taskResponse.TaskID)

	fmt.Printf("Successfully deleted %s:%s\n", endpointID, path)
	return nil
}

// confirmAction asks the user for confirmation
func confirmAction(prompt string) bool {
	confirm := promptui.Prompt{
		Label:     prompt,
		IsConfirm: true,
	}

	result, err := confirm.Run()
	if err != nil {
		return false
	}

	return strings.ToLower(result) == "y"
}
