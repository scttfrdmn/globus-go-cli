// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

var (
	deleteRecursive     bool
	deleteIgnoreMissing bool
)

// DeleteCmd returns the delete command
func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete ENDPOINT_ID:PATH",
		Short: "Submit a delete task for a path on an endpoint",
		Long: `Submit a delete task for a file or directory on a Globus endpoint.

This is the task-based delete operation. If --recursive is specified, it will
delete directories and their contents. If --ignore-missing is specified, the
task will not error when the path does not exist.

Examples:
  globus transfer delete ddb59aef-6d04-11e5-ba46-22000b92c6ec:/path/to/file
  globus transfer delete --recursive ddb59aef-6d04-11e5-ba46-22000b92c6ec:/path/to/directory`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse endpoint ID and path
			endpointID, path := parseEndpointAndPath(args[0])

			if path == "/" {
				return fmt.Errorf("path must be specified for delete command")
			}

			return submitDeleteTask(cmd, endpointID, path)
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&deleteRecursive, "recursive", false, "Delete directories and their contents recursively")
	cmd.Flags().BoolVar(&deleteIgnoreMissing, "ignore-missing", false, "Do not error if the path does not exist")

	return cmd
}

// submitDeleteTask submits a delete task for a path on an endpoint
func submitDeleteTask(cmd *cobra.Command, endpointID, path string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// The v4 SDK requires a submission ID minted from the service and carries
	// recursion/ignore-missing on the Delete request itself.
	submissionID, err := transferClient.GetSubmissionID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get submission ID: %w", err)
	}

	deleteRequest := &transfer.Delete{
		DATA_TYPE:     "delete",
		SubmissionID:  submissionID,
		Endpoint:      endpointID,
		Recursive:     deleteRecursive,
		IgnoreMissing: deleteIgnoreMissing,
		Items: []transfer.DeleteItem{
			{DATA_TYPE: "delete_item", Path: path},
		},
	}

	taskResponse, err := transferClient.SubmitDelete(ctx, deleteRequest)
	if err != nil {
		return fmt.Errorf("failed to submit delete task: %w", err)
	}

	fmt.Printf("Delete task submitted. Task ID: %s\n", taskResponse.TaskID)
	return nil
}
