// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

// BookmarkCmd returns the bookmark command
func BookmarkCmd() *cobra.Command {
	// bookmarkCmd represents the bookmark command
	bookmarkCmd := &cobra.Command{
		Use:   "bookmark",
		Short: "Commands for managing Globus bookmarks",
		Long: `Commands for managing Globus Transfer bookmarks including listing,
creating, renaming, and deleting saved endpoint/path bookmarks.`,
	}

	// Add bookmark subcommands
	bookmarkCmd.AddCommand(
		bookmarkListCmd(),
		bookmarkShowCmd(),
		bookmarkCreateCmd(),
		bookmarkRenameCmd(),
		bookmarkDeleteCmd(),
	)

	return bookmarkCmd
}

// Options for bookmark creation
var (
	bookmarkName       string
	bookmarkCollection string
	bookmarkPath       string
)

// bookmarkListCmd returns the bookmark list command
func bookmarkListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List Globus bookmarks",
		Long: `List Globus Transfer bookmarks belonging to the current user.

This command lists all saved bookmarks, each referencing a collection
and a path within that collection.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listBookmarks(cmd)
		},
	}
}

// bookmarkShowCmd returns the bookmark show command
func bookmarkShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show BOOKMARK_ID",
		Short: "Show bookmark details",
		Long: `Show detailed information about a specific Globus bookmark.

This command displays all available details about the specified bookmark,
including the collection and path it references.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showBookmark(cmd, args[0])
		},
	}
}

// bookmarkCreateCmd returns the bookmark create command
func bookmarkCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Globus bookmark",
		Long: `Create a Globus Transfer bookmark referencing a collection and path.

A bookmark saves an endpoint/collection UUID together with a path so it can
be referenced later by name.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createBookmark(cmd)
		},
	}

	// Add flags for creation
	cmd.Flags().StringVar(&bookmarkName, "name", "", "Name for the bookmark (required)")
	cmd.Flags().StringVar(&bookmarkCollection, "collection", "", "Endpoint/collection UUID (required)")
	cmd.Flags().StringVar(&bookmarkPath, "path", "", "Path within the collection (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("collection")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}

// bookmarkRenameCmd returns the bookmark rename command
func bookmarkRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename BOOKMARK_ID NEW_NAME",
		Short: "Rename a Globus bookmark",
		Long: `Rename a Globus Transfer bookmark.

This command changes the display name of the specified bookmark.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return renameBookmark(cmd, args[0], args[1])
		},
	}
}

// bookmarkDeleteCmd returns the bookmark delete command
func bookmarkDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete BOOKMARK_ID",
		Short: "Delete a Globus bookmark",
		Long: `Delete a Globus Transfer bookmark.

This command permanently removes the specified bookmark.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteBookmark(cmd, args[0])
		},
	}
}

// listBookmarks lists Globus bookmarks
func listBookmarks(cmd *cobra.Command) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Get the bookmarks
	resp, err := transferClient.ListBookmarks(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list bookmarks: %w", err)
	}

	// Route all formats through the shared formatter so -F (text/json/unix) and
	// --jmespath/--jq work uniformly. For JSON/JMESPath, emit the raw bookmark
	// documents; for text/unix, a projected row set.
	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())

	if formatter.Format == output.FormatJSON {
		// Emit the enveloped {"DATA":[...]} shape, matching the Python CLI. (The
		// SDK's BookmarkList type has no DATA json tag, so wrap explicitly.)
		return formatter.FormatOutput(map[string]interface{}{"DATA": resp.Bookmarks}, nil)
	}

	type bookmarkRow struct {
		ID           string
		Name         string
		CollectionID string
		Path         string
	}
	rows := make([]bookmarkRow, 0, len(resp.Bookmarks))
	for _, b := range resp.Bookmarks {
		rows = append(rows, bookmarkRow{
			ID: b.ID, Name: b.Name, CollectionID: b.CollectionID, Path: b.Path,
		})
	}
	return formatter.FormatOutput(rows, []string{"ID", "Name", "CollectionID", "Path"})
}

// showBookmark shows details for a specific bookmark
func showBookmark(cmd *cobra.Command, bookmarkID string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Get the bookmark
	bookmark, err := transferClient.GetBookmark(ctx, bookmarkID)
	if err != nil {
		return fmt.Errorf("failed to get bookmark: %w", err)
	}

	// For json/unix or a --jmespath/--jq expression, route through the shared
	// formatter (emitting the raw bookmark document). Otherwise render the text
	// detail view below.
	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(bookmark, nil)
	}

	{
		// Output as text
		fmt.Println("Bookmark Details:")
		fmt.Printf("  ID:             %s\n", bookmark.ID)
		fmt.Printf("  Name:           %s\n", bookmark.Name)
		fmt.Printf("  CollectionID:   %s\n", bookmark.CollectionID)
		fmt.Printf("  Path:           %s\n", bookmark.Path)
	}

	return nil
}

// createBookmark creates a Globus bookmark
func createBookmark(cmd *cobra.Command) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Create the bookmark
	bookmark, err := transferClient.CreateBookmark(ctx, &transfer.BookmarkCreate{
		Collection: bookmarkCollection,
		Name:       bookmarkName,
		Path:       bookmarkPath,
	})
	if err != nil {
		return fmt.Errorf("failed to create bookmark: %w", err)
	}

	fmt.Printf("Created bookmark %s (%s)\n", bookmark.ID, bookmark.Name)
	return nil
}

// renameBookmark renames a Globus bookmark
func renameBookmark(cmd *cobra.Command, bookmarkID, newName string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Update the bookmark name
	if _, err := transferClient.UpdateBookmark(ctx, bookmarkID, &transfer.BookmarkUpdate{Name: &newName}); err != nil {
		return fmt.Errorf("failed to rename bookmark: %w", err)
	}

	fmt.Printf("Renamed bookmark %s to %s\n", bookmarkID, newName)
	return nil
}

// deleteBookmark deletes a Globus bookmark
func deleteBookmark(cmd *cobra.Command, bookmarkID string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Delete the bookmark
	if err := transferClient.DeleteBookmark(ctx, bookmarkID); err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	fmt.Printf("Deleted bookmark %s\n", bookmarkID)
	return nil
}
