// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/flows"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	listLimit      int
	listOffset     int
	listFilter     string
	listFilterRole string
	listOrderBy    string
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
	// list_flows is marker-paginated; limit/offset are not accepted (kept as
	// deprecated no-ops for compatibility).
	ListCmd.Flags().IntVar(&listLimit, "limit", 0, "Deprecated: list_flows is marker-paginated")
	ListCmd.Flags().IntVar(&listOffset, "offset", 0, "Deprecated: list_flows is marker-paginated")
	_ = ListCmd.Flags().MarkDeprecated("limit", "list_flows is marker-paginated")
	_ = ListCmd.Flags().MarkDeprecated("offset", "list_flows is marker-paginated")
	ListCmd.Flags().StringVar(&listFilter, "filter", "", "Filter flows by text")
	ListCmd.Flags().StringVar(&listFilterRole, "filter-role", "", "Filter by the caller's role (flow_viewer, flow_starter, flow_administrator, flow_owner, run_manager, run_monitor)")
	ListCmd.Flags().StringVar(&listOrderBy, "orderby", "created_at", "Order results by field (created_at, updated_at, title)")
}

func runFlowsList(cmd *cobra.Command, args []string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Flows client authorized for the current profile.
	flowsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Build list options. list_flows is marker-paginated and rejects
	// limit/offset (HTTP 422), so only filter/orderby are sent.
	options := &flows.ListFlowsOptions{}
	if listOrderBy != "" {
		options.OrderBy = []string{listOrderBy}
	}
	if listFilter != "" {
		options.FilterFulltext = listFilter
	}
	if listFilterRole != "" {
		options.FilterRoles = []string{listFilterRole}
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

		fmt.Printf("%-36s  %-40s  %-36s\n", "Flow ID", "Title", "Owner")
		fmt.Printf("%s  %s  %s\n",
			"------------------------------------",
			"----------------------------------------",
			"------------------------------------")

		for _, flow := range flowList.Flows {
			title := flow.Title
			if len(title) > 40 {
				title = title[:37] + "..."
			}

			fmt.Printf("%-36s  %-40s  %-36s\n",
				flow.ID,
				title,
				flow.OwnerID)
		}

		fmt.Printf("\nTotal: %d flow(s)\n", len(flowList.Flows))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"ID", "Title", "Description", "OwnerID", "Created", "Updated"}
		if err := formatter.FormatOutput(flowList.Flows, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
