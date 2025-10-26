// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package timer

import (
	"context"
	"fmt"
	"os"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/timers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ListCmd represents the timer list command
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Globus timers",
	Long: `List all timers you have created.

Timers can be for recurring transfers or flow executions.

Examples:
  # List all your timers
  globus timer list

  # List with JSON output
  globus timer list --format=json

Output Formats:
  --format=text    Human-readable table (default)
  --format=json    JSON format
  --format=csv     CSV format`,
	RunE: runListTimers,
}

func runListTimers(cmd *cobra.Command, args []string) error {
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

	// Create timers client
	timersClient, err := timers.NewClient(
		timers.WithAuthorizer(coreAuthorizer),
	)
	if err != nil {
		return fmt.Errorf("failed to create timers client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// List timers
	timerList, err := timersClient.ListTimers(ctx, nil)
	if err != nil {
		return fmt.Errorf("error listing timers: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable table
		if len(timerList.Timers) == 0 {
			fmt.Println("No timers found.")
			return nil
		}

		fmt.Printf("%-36s  %-30s  %-20s  %-10s\n", "Timer ID", "Name", "Type", "Status")
		fmt.Printf("%s  %s  %s  %s\n",
			"------------------------------------",
			"------------------------------",
			"--------------------",
			"----------")

		for _, timer := range timerList.Timers {
			name := timer.Name
			if len(name) > 30 {
				name = name[:27] + "..."
			}

			// Determine timer type from callback
			timerType := "unknown"
			if timer.Callback != nil {
				timerType = timer.Callback.Type
			}

			status := "active"
			if timer.Status != "" {
				status = timer.Status
			}

			fmt.Printf("%-36s  %-30s  %-20s  %-10s\n",
				timer.ID,
				name,
				timerType,
				status)
		}

		fmt.Printf("\nTotal: %d timer(s)\n", len(timerList.Timers))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"ID", "Name", "CallbackType", "Status"}
		if err := formatter.FormatOutput(timerList.Timers, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
