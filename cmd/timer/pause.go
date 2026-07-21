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

// PauseCmd represents the timer pause command
var PauseCmd = &cobra.Command{
	Use:   "pause TIMER_ID",
	Short: "Pause a timer",
	Long: `Pause a running timer to stop it from executing.

The timer will not run until it is resumed. This is useful for temporarily
disabling a timer without deleting it.

Examples:
  # Pause a timer
  globus timer pause TIMER_ID

  # Pause and verify
  globus timer pause TIMER_ID && globus timer show TIMER_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runPauseTimer,
}

func runPauseTimer(cmd *cobra.Command, args []string) error {
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

	// Pause the timer (returns no body; success is signaled by a nil error).
	if err := timersClient.PauseTimer(ctx, timerID); err != nil {
		return fmt.Errorf("error pausing timer: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Timer paused successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Timer ID:    %s\n", timerID)
	fmt.Fprintf(os.Stdout, "Name:        %s\n", timer.Name)

	return nil
}
