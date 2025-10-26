// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"fmt"
	"os"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/flows"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	listLimit  int
	listOffset int
	listFilter string
	listOrderBy string
)

// ListCmd represents the flows list command
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Globus Flows",
	Long: `List flows you own or have access to.

This command lists all flows where you have permissions to view or execute.
You can filter by various criteria and control the number of results.

Examples:
  # List all your flows
  globus flows list

  # Limit results
  globus flows list --limit 20

  # Filter by keyword
  globus flows list --filter "transfer"

  # JSON output for scripting
  globus flows list --format json`,
	RunE: runFlowsList,
}

func init() {
	ListCmd.Flags().IntVar(&listLimit, "limit", 25, "Maximum number of flows to return")
	ListCmd.Flags().IntVar(&listOffset, "offset", 0, "Offset for pagination")
	ListCmd.Flags().StringVar(&listFilter, "filter", "", "Filter flows by text")
	ListCmd.Flags().StringVar(&listOrderBy, "orderby", "created_at", "Order results by field (created_at, updated_at, title)")
}

func runFlowsList(cmd *cobra.Command, args []string) error {
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

	// Create flows client
	flowsClient, err := flows.NewClient(
		flows.WithAuthorizer(coreAuthorizer),
	)
	if err != nil {
		return fmt.Errorf("failed to create flows client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build list options
	options := &flows.ListFlowsOptions{
		Limit:   listLimit,
		Offset:  listOffset,
		OrderBy: listOrderBy,
	}
	if listFilter != "" {
		options.Q = listFilter
	}

	// List flows
	flowList, err := flowsClient.ListFlows(ctx, options)
	if err != nil {
		return fmt.Errorf("error listing flows: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable table
		if len(flowList.Flows) == 0 {
			fmt.Println("No flows found.")
			return nil
		}

		fmt.Printf("%-36s  %-40s  %-8s  %-6s\n", "Flow ID", "Title", "Public", "Runs")
		fmt.Printf("%s  %s  %s  %s\n",
			"------------------------------------",
			"----------------------------------------",
			"--------",
			"------")

		for _, flow := range flowList.Flows {
			title := flow.Title
			if len(title) > 40 {
				title = title[:37] + "..."
			}

			public := "No"
			if flow.Public {
				public = "Yes"
			}

			fmt.Printf("%-36s  %-40s  %-8s  %-6d\n",
				flow.ID,
				title,
				public,
				flow.RunCount)
		}

		fmt.Printf("\nTotal: %d flow(s)\n", len(flowList.Flows))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"ID", "Title", "Description", "Public", "RunCount", "CreatedAt", "UpdatedAt"}
		if err := formatter.FormatOutput(flowList.Flows, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
