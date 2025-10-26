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

// ShowCmd represents the group show command
var ShowCmd = &cobra.Command{
	Use:   "show GROUP_ID",
	Short: "Show details for a specific group",
	Long: `Show detailed information about a specific Globus group.

This displays comprehensive information including group metadata, policies,
membership counts, and administrative settings.

Examples:
  # Show group details
  globus group show 12345678-1234-1234-1234-123456789abc

  # Show with JSON output
  globus group show 12345678-1234-1234-1234-123456789abc --format=json

Output Formats:
  --format=text    Human-readable output (default)
  --format=json    JSON format
  --format=csv     CSV format`,
	Args: cobra.ExactArgs(1),
	RunE: runShowGroup,
}

func runShowGroup(cmd *cobra.Command, args []string) error {
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

	// Get group details
	group, err := groupsClient.GetGroup(ctx, groupID)
	if err != nil {
		return fmt.Errorf("error getting group: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		fmt.Printf("Group Information\n")
		fmt.Printf("=================\n\n")
		fmt.Printf("ID:              %s\n", group.ID)
		fmt.Printf("Name:            %s\n", group.Name)
		fmt.Printf("Description:     %s\n", group.Description)
		fmt.Printf("Identity ID:     %s\n", group.IdentityID)
		fmt.Printf("Member Count:    %d\n", group.MemberCount)
		fmt.Printf("\nPermissions\n")
		fmt.Printf("-----------\n")
		fmt.Printf("Is Group Admin:  %v\n", group.IsGroupAdmin)
		fmt.Printf("Is Member:       %v\n", group.IsMember)
		fmt.Printf("\nSettings\n")
		fmt.Printf("--------\n")
		fmt.Printf("Public Group:    %v\n", group.PublicGroup)
		fmt.Printf("Requires Agreement: %v\n", group.RequiresSignAgreement)
		if group.RequiresSignAgreement && group.SignAgreementMessage != "" {
			fmt.Printf("Agreement Message: %s\n", group.SignAgreementMessage)
		}
		if group.ParentID != "" {
			fmt.Printf("Parent Group ID: %s\n", group.ParentID)
		}
		fmt.Printf("\nTimestamps\n")
		fmt.Printf("----------\n")
		fmt.Printf("Created:         %s\n", group.Created.Format(time.RFC3339))
		fmt.Printf("Last Updated:    %s\n", group.LastUpdated.Format(time.RFC3339))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"ID", "Name", "Description", "IdentityID", "MemberCount", "IsGroupAdmin", "IsMember", "PublicGroup"}
		if err := formatter.FormatOutput(group, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
