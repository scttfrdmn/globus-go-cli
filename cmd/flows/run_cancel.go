// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"fmt"
	"os"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/flows"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RunCancelCmd represents the flows run cancel command
var RunCancelCmd = &cobra.Command{
	Use:   "cancel RUN_ID",
	Short: "Cancel a flow run",
	Long: `Cancel an active flow run.

This attempts to gracefully cancel a running flow execution. The flow's
cancellation logic will be invoked if defined.

Examples:
  # Cancel a run
  globus flows run cancel RUN_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsRunCancel,
}

func runFlowsRunCancel(cmd *cobra.Command, args []string) error {
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

	// Cancel run
	err = flowsClient.CancelRun(ctx, runID)
	if err != nil {
		return fmt.Errorf("error canceling run: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Run %s canceled successfully.\n", runID)
	fmt.Fprintf(os.Stdout, "\nCheck status with: globus flows run show %s\n", runID)

	return nil
}
