// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package timer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/timers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deleteConfirm bool

// DeleteCmd represents the timer delete command
var DeleteCmd = &cobra.Command{
	Use:   "delete TIMER_ID",
	Short: "Delete a timer",
	Long: `Delete a timer permanently.

This will stop the timer from running and remove it from your account.
This action cannot be undone.

Examples:
  # Delete with confirmation prompt
  globus timer delete TIMER_ID

  # Delete without confirmation
  globus timer delete TIMER_ID --confirm`,
	Args: cobra.ExactArgs(1),
	RunE: runDeleteTimer,
}

func init() {
	DeleteCmd.Flags().BoolVar(&deleteConfirm, "confirm", false, "Skip confirmation prompt")
}

func runDeleteTimer(cmd *cobra.Command, args []string) error {
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

	// Get timer first to display name and confirm
	timer, err := timersClient.GetTimer(ctx, timerID)
	if err != nil {
		return fmt.Errorf("error getting timer: %w", err)
	}

	// Confirmation prompt unless --confirm flag is set
	if !deleteConfirm {
		fmt.Printf("Are you sure you want to delete timer '%s' (%s)? [y/N]: ", timer.Name, timerID)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading confirmation: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Timer deletion cancelled.")
			return nil
		}
	}

	// Delete the timer
	err = timersClient.DeleteTimer(ctx, timerID)
	if err != nil {
		return fmt.Errorf("error deleting timer: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Timer deleted successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Timer ID:    %s\n", timerID)
	fmt.Fprintf(os.Stdout, "Name:        %s\n", timer.Name)

	return nil
}
