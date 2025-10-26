// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

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
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/flows"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deleteYes bool

// DeleteCmd represents the flows delete command
var DeleteCmd = &cobra.Command{
	Use:   "delete FLOW_ID",
	Short: "Delete a Globus Flow",
	Long: `Delete a flow that you own.

WARNING: This action cannot be undone. All flow metadata and definition
will be permanently deleted. Existing run history will be preserved.

Examples:
  # Delete a flow (with confirmation)
  globus flows delete FLOW_ID

  # Delete without confirmation prompt
  globus flows delete FLOW_ID --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsDelete,
}

func init() {
	DeleteCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "Skip confirmation prompt")
}

func runFlowsDelete(cmd *cobra.Command, args []string) error {
	flowID := args[0]

	// Confirm deletion unless --yes flag is set
	if !deleteYes {
		fmt.Fprintf(os.Stderr, "Are you sure you want to delete flow %s? This cannot be undone.\n", flowID)
		fmt.Fprintf(os.Stderr, "Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

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

	// Delete flow
	err = flowsClient.DeleteFlow(ctx, flowID)
	if err != nil {
		return fmt.Errorf("error deleting flow: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Flow %s deleted successfully.\n", flowID)

	return nil
}
