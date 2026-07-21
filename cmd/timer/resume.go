// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package timer

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// ResumeCmd represents the timer resume command
var ResumeCmd = &cobra.Command{
	Use:   "resume TIMER_ID",
	Short: "Resume a paused timer",
	Long: `Resume a previously paused timer to restart execution.

The timer will begin running again according to its schedule.

Examples:
  # Resume a paused timer
  globus timer resume TIMER_ID

  # Resume and verify
  globus timer resume TIMER_ID && globus timer show TIMER_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runResumeTimer,
}

func runResumeTimer(cmd *cobra.Command, args []string) error {
	timerID := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Timers client authorized for the current profile.
	timersClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Get timer first to display name
	timer, err := timersClient.GetTimer(ctx, timerID)
	if err != nil {
		return fmt.Errorf("error getting timer: %w", err)
	}

	// Resume the timer. The optional *bool controls whether stored credentials
	// are refreshed on resume; nil leaves them unchanged. Returns no body.
	if err := timersClient.ResumeTimer(ctx, timerID, nil); err != nil {
		return fmt.Errorf("error resuming timer: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Timer resumed successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Timer ID:    %s\n", timerID)
	fmt.Fprintf(os.Stdout, "Name:        %s\n", timer.Name)

	return nil
}
