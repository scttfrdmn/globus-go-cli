// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
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

// MemberListCmd represents the member list command
var MemberListCmd = &cobra.Command{
	Use:   "list GROUP_ID",
	Short: "List members of a group",
	Long: `List all members of a Globus group, including their roles and status.

Examples:
  # List all members
  globus group member list GROUP_ID

  # List with JSON output
  globus group member list GROUP_ID --format=json

Output Formats:
  --format=text    Human-readable table (default)
  --format=json    JSON format
  --format=csv     CSV format`,
	Args: cobra.ExactArgs(1),
	RunE: runMemberList,
}

func runMemberList(cmd *cobra.Command, args []string) error {
	groupID := args[0]

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

	// Create authorizer
	tokenAuthorizer := authorizers.NewStaticTokenAuthorizer(tokenInfo.AccessToken)
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

	// Members are returned by fetching the group with the memberships
	// representation (the API has no standalone list-members route).
	group, err := groupsClient.GetGroup(ctx, groupID, &groups.GetGroupOptions{
		Include: []string{"memberships"},
	})
	if err != nil {
		return fmt.Errorf("error listing members: %w", err)
	}
	members := group.Memberships

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable table
		if len(members) == 0 {
			fmt.Println("No members found.")
			return nil
		}

		fmt.Printf("%-36s  %-30s  %-20s  %-10s\n", "Identity ID", "Username", "Role", "Status")
		fmt.Printf("%s  %s  %s  %s\n",
			"------------------------------------",
			"------------------------------",
			"--------------------",
			"----------")

		for _, member := range members {
			username := member.Username
			if username == "" {
				username = "(unknown)"
			}

			role := member.Role
			if role == "" {
				role = "member"
			}

			status := member.Status
			if status == "" {
				status = "active"
			}

			fmt.Printf("%-36s  %-30s  %-20s  %-10s\n",
				member.IdentityID,
				username,
				role,
				status)
		}

		fmt.Printf("\nTotal: %d member(s)\n", len(members))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"IdentityID", "Username", "Role", "Status"}
		if err := formatter.FormatOutput(members, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
