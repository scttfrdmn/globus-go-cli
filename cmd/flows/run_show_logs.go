// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"encoding/json"
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
	runShowLogsLimit  int
	runShowLogsOffset int
)

// RunShowLogsCmd represents the flows run show-logs command
var RunShowLogsCmd = &cobra.Command{
	Use:   "show-logs RUN_ID",
	Short: "Show logs for a flow run",
	Long: `Display execution logs for a specific flow run.

Logs include detailed information about each step of the flow execution,
including timestamps, status changes, and any error messages.

Examples:
  # Show all logs for a run
  globus flows run show-logs RUN_ID

  # Limit number of log entries
  globus flows run show-logs RUN_ID --limit 50

  # JSON output for scripting
  globus flows run show-logs RUN_ID --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsRunShowLogs,
}

func init() {
	RunShowLogsCmd.Flags().IntVar(&runShowLogsLimit, "limit", 100, "Maximum number of log entries to return")
	RunShowLogsCmd.Flags().IntVar(&runShowLogsOffset, "offset", 0, "Offset for pagination")
}

func runFlowsRunShowLogs(cmd *cobra.Command, args []string) error {
	runID := args[0]

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

	// Get run logs
	logs, err := flowsClient.GetRunLogs(ctx, runID, runShowLogsLimit, runShowLogsOffset)
	if err != nil {
		return fmt.Errorf("error getting run logs: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		if len(logs.Entries) == 0 {
			fmt.Println("No log entries found.")
			return nil
		}

		fmt.Printf("Run Logs for %s\n", runID)
		fmt.Printf("====================\n\n")

		for i, entry := range logs.Entries {
			fmt.Printf("Entry %d:\n", i+1)
			fmt.Printf("  Time:    %s\n", entry.CreatedAt.Format(time.RFC3339))
			fmt.Printf("  Code:    %s\n", entry.Code)
			if entry.Description != "" {
				fmt.Printf("  Desc:    %s\n", entry.Description)
			}

			// Display details if available
			if entry.Details != nil {
				detailsJSON, _ := json.MarshalIndent(entry.Details, "    ", "  ")
				fmt.Printf("  Details:\n    %s\n", string(detailsJSON))
			}
			fmt.Println()
		}

		fmt.Printf("Total: %d log entr(ies)\n", len(logs.Entries))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"Timestamp", "Code", "Description"}
		if err := formatter.FormatOutput(logs.Entries, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
