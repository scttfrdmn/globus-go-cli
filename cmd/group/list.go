// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package group

import (
	"context"
	"fmt"
	"os"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/groups"
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
	ListCmd.Flags().BoolVar(&listMyGroupsOnly, "my-groups", false, "List only groups where you are a member")
	ListCmd.Flags().StringArrayVar(&listIncludeStatuses, "include-status", []string{}, "Include groups with specific statuses (active, pending, etc.)")
}

func runListGroups(cmd *cobra.Command, args []string) error {
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

	// Create a simple static token authorizer
	tokenAuthorizer := authorizers.NewStaticTokenAuthorizer(tokenInfo.AccessToken)

	// Create a core authorizer adapter for compatibility
	coreAuthorizer := authorizers.ToCore(tokenAuthorizer)

	// Create groups client
	groupsClient, err := groups.NewClient(
		groups.WithAuthorizer(coreAuthorizer),
	)
	if err != nil {
		return fmt.Errorf("failed to create groups client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Prepare options for listing groups
	options := &groups.ListGroupsOptions{
		MyGroups: listMyGroupsOnly,
	}

	// SDK v3.65.0-1 supports status filtering via Statuses field
	if len(listIncludeStatuses) > 0 {
		options.Statuses = listIncludeStatuses
	}

	// List groups
	groupList, err := groupsClient.ListGroups(ctx, options)
	if err != nil {
		return fmt.Errorf("error listing groups: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable table
		if len(groupList.Groups) == 0 {
			fmt.Println("No groups found.")
			return nil
		}

		fmt.Printf("%-36s  %-40s  %-8s  %-8s\n", "ID", "Name", "Members", "Admin")
		fmt.Printf("%s  %s  %s  %s\n",
			"------------------------------------",
			"----------------------------------------",
			"--------",
			"--------")

		for _, group := range groupList.Groups {
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

		fmt.Printf("\nTotal: %d group(s)\n", len(groupList.Groups))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"ID", "Name", "Description", "MemberCount", "IsGroupAdmin", "IsMember"}
		if err := formatter.FormatOutput(groupList.Groups, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
