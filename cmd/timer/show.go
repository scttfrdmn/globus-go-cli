// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package timer

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Timers client authorized for the current profile.
	timersClient, err := getClient(ctx)
	if err != nil {
		return err
	}

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
