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

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
)

var (
	// Options for rm command
	recursiveDelete    bool
	ignoreDeleteErrors bool
)

// RmCmd returns the rm command
func RmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm ENDPOINT_ID[:PATH]",
		Short: "Delete files or directories on a Globus endpoint",
		Long: `Delete files or directories on a Globus endpoint.

This command deletes files or directories at the specified path on a Globus endpoint.
Use --recursive to delete directories recursively.

Examples:
  globus transfer rm endpoint_id:/path/to/file       # Delete a file
  globus transfer rm endpoint_id:/path/to/dir -r     # Delete a directory recursively`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteFiles(cmd, args[0])
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&recursiveDelete, "recursive", "r", false, "Delete directories recursively")
	cmd.Flags().BoolVar(&ignoreDeleteErrors, "ignore-errors", false, "Ignore errors during deletion")

	return cmd
}

// deleteFiles deletes files or directories on a Globus endpoint
func deleteFiles(cmd *cobra.Command, arg string) error {
	// Parse the endpoint ID and path
	endpointID, path, err := parseEndpointPath(arg)
	if err != nil {
		return fmt.Errorf("invalid endpoint path: %w", err)
	}

	// Ensure the path is provided
	if path == "" || path == "/" {
		return fmt.Errorf("a path must be provided. Refusing to delete root directory")
	}

	// Get the transfer client
	transferClient, err := getTransferClient()
	if err != nil {
		return err
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if the path exists and if it's a directory
	fileList, err := transferClient.ListFiles(ctx, endpointID, filepath.Dir(path), &pkg.ListFileOptions{})
	if err != nil {
		return fmt.Errorf("failed to check if path exists: %w", err)
	}

	// Check if it's a directory
	isDirectory := false
	fileName := filepath.Base(path)
	exists := false

	for _, file := range fileList.Data {
		if file.Name == fileName {
			exists = true
			isDirectory = file.Type == "dir"
			break
		}
	}

	if !exists {
		return fmt.Errorf("path does not exist: %s", path)
	}

	if isDirectory && !recursiveDelete {
		return fmt.Errorf("cannot delete directory without --recursive flag")
	}

	// Confirm with the user
	if isDirectory {
		fmt.Printf("You are about to delete the directory %s:%s and all its contents\n", endpointID, path)
	} else {
		fmt.Printf("You are about to delete the file %s:%s\n", endpointID, path)
	}
	fmt.Print("Are you sure? [y/N] ")
	
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	// Submit the delete task
	fmt.Printf("Submitting delete task for %s:%s\n", endpointID, path)
	
	options := make(map[string]interface{})
	options["label"] = fmt.Sprintf("CLI Delete %s", time.Now().Format("2006-01-02 15:04:05"))
	options["recursive"] = recursiveDelete
	options["ignore_missing"] = true
	options["interpret_globs"] = false
	options["ignore_errors"] = ignoreDeleteErrors

	deleteResult, err := transferClient.SubmitDelete(ctx, endpointID, []string{path}, options)
	if err != nil {
		return fmt.Errorf("failed to submit delete task: %w", err)
	}

	fmt.Printf("Delete task submitted with task ID: %s\n", deleteResult.TaskID)
	
	// Save the task ID to a file for later status checks
	homeDir, err := os.UserHomeDir()
	if err == nil {
		taskFile := filepath.Join(homeDir, ".globus-cli", "last-task-id")
		os.WriteFile(taskFile, []byte(deleteResult.TaskID), 0600)
	}

	fmt.Println("\nYou can check the status of this task with:")
	fmt.Printf("  globus transfer task show %s\n", deleteResult.TaskID)

	return nil
}