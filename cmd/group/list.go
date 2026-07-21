// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package group

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listMyGroupsOnly bool
var listIncludeStatuses []string

// ListCmd represents the group list command
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Globus groups",
	Long: `List Globus groups you belong to or have access to.

By default, this command lists all groups where you are a member, manager, or admin.

Examples:
  # List all your groups
  globus group list

  # List only groups where you are a member
  globus group list --my-groups

  # List groups with specific statuses (SDK v3.65.0+ feature)
  globus group list --include-status active --include-status pending

Output Formats:
  --format=text    Human-readable table (default)
  --format=json    JSON format
  --format=csv     CSV format`,
	RunE: runListGroups,
}

func init() {
	// The Groups API only returns the caller's own groups, so this flag is a
	// no-op kept for backward compatibility.
	ListCmd.Flags().BoolVar(&listMyGroupsOnly, "my-groups", false, "Deprecated: the API always lists only your groups")
	_ = ListCmd.Flags().MarkDeprecated("my-groups", "the API only lists your groups")
	ListCmd.Flags().StringArrayVar(&listIncludeStatuses, "include-status", []string{}, "Include groups with specific statuses (active, pending, etc.)")
}

func runListGroups(cmd *cobra.Command, args []string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Groups client authorized for the current profile.
	groupsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// The Globus Groups API only lists the caller's own groups
	// (GET /groups/my_groups); optional status filtering is passed through.
	groupList, err := groupsClient.GetMyGroups(ctx, listIncludeStatuses)
	if err != nil {
		return fmt.Errorf("error listing groups: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable table
		if len(groupList) == 0 {
			fmt.Println("No groups found.")
			return nil
		}

		fmt.Printf("%-36s  %-40s  %-8s  %-8s\n", "ID", "Name", "Members", "Admin")
		fmt.Printf("%s  %s  %s  %s\n",
			"------------------------------------",
			"----------------------------------------",
			"--------",
			"--------")

		for _, group := range groupList {
			name := group.Name
			if len(name) > 40 {
				name = name[:37] + "..."
			}

			admin := "No"
			if group.IsGroupAdmin {
				admin = "Yes"
			}

			fmt.Printf("%-36s  %-40s  %-8d  %-8s\n",
				group.ID,
				name,
				group.MemberCount,
				admin)
		}

		fmt.Printf("\nTotal: %d group(s)\n", len(groupList))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"ID", "Name", "Description", "MemberCount", "IsGroupAdmin", "IsMember"}
		if err := formatter.FormatOutput(groupList, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
