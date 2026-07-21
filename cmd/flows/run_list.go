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
	runListLimit   int
	runListOffset  int
	runListFlowID  string
	runListStatus  string
	runListOrderBy string
)

// RunListCmd represents the flows run list command
var RunListCmd = &cobra.Command{
	Use:   "list",
	Short: "List flow runs",
	Long: `List flow runs with optional filtering.

You can filter by flow ID, status, and other criteria. Results are paginated
and can be ordered by various fields.

Examples:
  # List all your runs
  globus flows run list

  # List runs for a specific flow
  globus flows run list --flow-id FLOW_ID

  # List only active runs
  globus flows run list --status ACTIVE

  # Limit results
  globus flows run list --limit 50

  # JSON output for scripting
  globus flows run list --format json`,
	RunE: runFlowsRunList,
}

func init() {
	// list_runs is marker-paginated; limit/offset are not accepted.
	RunListCmd.Flags().IntVar(&runListLimit, "limit", 0, "Deprecated: list_runs is marker-paginated")
	RunListCmd.Flags().IntVar(&runListOffset, "offset", 0, "Deprecated: list_runs is marker-paginated")
	_ = RunListCmd.Flags().MarkDeprecated("limit", "list_runs is marker-paginated")
	_ = RunListCmd.Flags().MarkDeprecated("offset", "list_runs is marker-paginated")
	RunListCmd.Flags().StringVar(&runListFlowID, "flow-id", "", "Filter by flow ID")
	RunListCmd.Flags().StringVar(&runListStatus, "status", "", "Filter by status (ACTIVE, SUCCEEDED, FAILED, INACTIVE)")
	// Runs are ordered by run fields (e.g. start_time); created_at is a flow
	// field and is rejected here.
	RunListCmd.Flags().StringVar(&runListOrderBy, "orderby", "start_time DESC", "Order results by field (e.g. start_time DESC)")
}

func runFlowsRunList(cmd *cobra.Command, args []string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Flows client authorized for the current profile.
	flowsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Build list options. list_runs is marker-paginated and rejects
	// limit/offset (HTTP 422), so only filters are sent.
	options := &flows.ListRunsOptions{}
	if runListFlowID != "" {
		options.FilterFlowID = []string{runListFlowID}
	}

	// List runs
	runList, err := flowsClient.ListRuns(ctx, options)
	if err != nil {
		return fmt.Errorf("error listing runs: %w", err)
	}

	// The v4 ListRunsOptions has no server-side status filter, so apply the
	// --status filter client-side to preserve the previous behavior.
	if runListStatus != "" {
		filtered := runList.Runs[:0]
		for _, run := range runList.Runs {
			if run.Status == runListStatus {
				filtered = append(filtered, run)
			}
		}
		runList.Runs = filtered
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable table
		if len(runList.Runs) == 0 {
			fmt.Println("No runs found.")
			return nil
		}

		fmt.Printf("%-36s  %-36s  %-12s  %-20s\n", "Run ID", "Flow ID", "Status", "Started")
		fmt.Printf("%s  %s  %s  %s\n",
			"------------------------------------",
			"------------------------------------",
			"------------",
			"--------------------")

		for _, run := range runList.Runs {
			fmt.Printf("%-36s  %-36s  %-12s  %-20s\n",
				run.RunID,
				run.FlowID,
				run.Status,
				run.StartTime.Format("2006-01-02 15:04:05"))
		}

		fmt.Printf("\nTotal: %d run(s)\n", len(runList.Runs))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"RunID", "FlowID", "FlowTitle", "Status", "Label", "RunOwner", "StartTime", "EndTime"}
		if err := formatter.FormatOutput(runList.Runs, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
