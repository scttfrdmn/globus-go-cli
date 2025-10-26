// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"fmt"
	"os"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/compute"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TaskCancelCmd represents the compute task cancel command
var TaskCancelCmd = &cobra.Command{
	Use:   "cancel TASK_ID",
	Short: "Cancel a running task",
	Long: `Cancel a task that is currently executing or pending.

This attempts to gracefully cancel the task execution. The task may
still complete if cancellation occurs after execution has finished.

Examples:
  # Cancel a task
  globus compute task cancel TASK_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runTaskCancel,
}

func runTaskCancel(cmd *cobra.Command, args []string) error {
	taskID := args[0]

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

	// Create compute client
	computeClient, err := compute.NewClient(
		compute.WithAuthorizer(coreAuthorizer),
	)
	if err != nil {
		return fmt.Errorf("failed to create compute client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Cancel task
	err = computeClient.CancelTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("error canceling task: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Task %s canceled successfully.\n", taskID)
	fmt.Fprintf(os.Stdout, "\nCheck status with: globus compute task show %s\n", taskID)

	return nil
}
