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

	// Pause the timer
	updatedTimer, err := timersClient.PauseTimer(ctx, timerID)
	if err != nil {
		return fmt.Errorf("error pausing timer: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Timer paused successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Timer ID:    %s\n", updatedTimer.ID)
	fmt.Fprintf(os.Stdout, "Name:        %s\n", timer.Name)
	fmt.Fprintf(os.Stdout, "Status:      %s\n", updatedTimer.Status)

	return nil
}
