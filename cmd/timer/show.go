// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
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

// ShowCmd represents the timer show command
var ShowCmd = &cobra.Command{
	Use:   "show TIMER_ID",
	Short: "Show details for a specific timer",
	Long: `Show detailed information about a specific Globus timer.

This displays comprehensive information including schedule, status,
and configuration details. In v3.39.0+, this includes Activity status.

Examples:
  # Show timer details
  globus timer show TIMER_ID

  # Show with JSON output
  globus timer show TIMER_ID --format=json

Output Formats:
  --format=text    Human-readable output (default)
  --format=json    JSON format
  --format=csv     CSV format`,
	Args: cobra.ExactArgs(1),
	RunE: runShowTimer,
}

func runShowTimer(cmd *cobra.Command, args []string) error {
	timerID := args[0]

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

	// Get timer details
	timer, err := timersClient.GetTimer(ctx, timerID)
	if err != nil {
		return fmt.Errorf("error getting timer: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		fmt.Printf("Timer Information\n")
		fmt.Printf("=================\n\n")
		fmt.Printf("Timer ID:    %s\n", timer.JobID)
		fmt.Printf("Name:        %s\n", timer.Name)
		fmt.Printf("Status:      %s\n", timer.Status)

		fmt.Printf("\nSchedule\n")
		fmt.Printf("--------\n")
		if timer.Schedule != nil {
			if timer.Schedule.Type != "" {
				fmt.Printf("Schedule Type: %s\n", timer.Schedule.Type)
			}
			if timer.Schedule.Datetime != "" {
				fmt.Printf("Run At:        %s\n", timer.Schedule.Datetime)
			}
			if timer.Schedule.IntervalSeconds > 0 {
				fmt.Printf("Interval:      %ds\n", timer.Schedule.IntervalSeconds)
			}
			if timer.Schedule.Start != "" {
				fmt.Printf("Start Time:    %s\n", timer.Schedule.Start)
			}
			if timer.Schedule.End != nil && timer.Schedule.End.Datetime != "" {
				fmt.Printf("End Time:      %s\n", timer.Schedule.End.Datetime)
			}
		}

		fmt.Printf("\nTimestamps\n")
		fmt.Printf("----------\n")
		if !timer.Created.IsZero() {
			fmt.Printf("Created:       %s\n", timer.Created.Format(time.RFC3339))
		}
		if !timer.LastRun.IsZero() {
			fmt.Printf("Last Run:      %s\n", timer.LastRun.Format(time.RFC3339))
		}
		if !timer.NextRun.IsZero() {
			fmt.Printf("Next Run:      %s\n", timer.NextRun.Format(time.RFC3339))
		}
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"JobID", "Name", "Status", "Schedule"}
		if err := formatter.FormatOutput(timer, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
