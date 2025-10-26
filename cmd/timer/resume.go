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
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/timers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	// Get timer first to display name
	timer, err := timersClient.GetTimer(ctx, timerID)
	if err != nil {
		return fmt.Errorf("error getting timer: %w", err)
	}

	// Resume the timer
	updatedTimer, err := timersClient.ResumeTimer(ctx, timerID)
	if err != nil {
		return fmt.Errorf("error resuming timer: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Timer resumed successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Timer ID:    %s\n", updatedTimer.ID)
	fmt.Fprintf(os.Stdout, "Name:        %s\n", timer.Name)
	fmt.Fprintf(os.Stdout, "Status:      %s\n", updatedTimer.Status)
	if updatedTimer.NextDue != nil {
		fmt.Fprintf(os.Stdout, "Next Run:    %s\n", updatedTimer.NextDue.Format(time.RFC3339))
	}

	return nil
}
