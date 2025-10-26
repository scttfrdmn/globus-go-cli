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
	RunListCmd.Flags().IntVar(&runListLimit, "limit", 25, "Maximum number of runs to return")
	RunListCmd.Flags().IntVar(&runListOffset, "offset", 0, "Offset for pagination")
	RunListCmd.Flags().StringVar(&runListFlowID, "flow-id", "", "Filter by flow ID")
	RunListCmd.Flags().StringVar(&runListStatus, "status", "", "Filter by status (ACTIVE, SUCCEEDED, FAILED, INACTIVE)")
	RunListCmd.Flags().StringVar(&runListOrderBy, "orderby", "created_at DESC", "Order results by field")
}

func runFlowsRunList(cmd *cobra.Command, args []string) error {
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
	options := &flows.ListRunsOptions{
		Limit:   runListLimit,
		Offset:  runListOffset,
		OrderBy: runListOrderBy,
	}
	if runListFlowID != "" {
		options.FlowID = runListFlowID
	}
	if runListStatus != "" {
		options.Status = runListStatus
	}

	// List runs
	runList, err := flowsClient.ListRuns(ctx, options)
	if err != nil {
		return fmt.Errorf("error listing runs: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable table
		if len(runList.Runs) == 0 {
			fmt.Println("No runs found.")
			return nil
		}

		fmt.Printf("%-36s  %-36s  %-12s  %-20s\n", "Run ID", "Flow ID", "Status", "Created")
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
				run.CreatedAt.Format("2006-01-02 15:04:05"))
		}

		fmt.Printf("\nTotal: %d run(s)\n", len(runList.Runs))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"RunID", "FlowID", "Status", "Label", "CreatedAt", "StartedAt", "CompletedAt", "UserID"}
		if err := formatter.FormatOutput(runList.Runs, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
