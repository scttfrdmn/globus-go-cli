// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"fmt"
	"os"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/compute"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// FunctionShowCmd represents the compute function show command
var FunctionShowCmd = &cobra.Command{
	Use:   "show FUNCTION_ID",
	Short: "Show details of a registered function",
	Long: `Display detailed information about a specific registered function.

This includes the function code, metadata, and configuration.

Examples:
  # Show function details
  globus compute function show FUNCTION_ID

  # Show function with JSON output
  globus compute function show FUNCTION_ID --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runFunctionShow,
}

func runFunctionShow(cmd *cobra.Command, args []string) error {
	functionID := args[0]

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

	// Get function
	function, err := computeClient.GetFunction(ctx, functionID)
	if err != nil {
		return fmt.Errorf("error getting function: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable. The function is an open-ended document.
		fmt.Printf("Function Details\n")
		fmt.Printf("================\n\n")

		id := mapStr(function, "function_uuid")
		if id == "" {
			id = mapStr(function, "function_id")
		}
		fmt.Printf("Function ID:   %s\n", id)
		if n := mapStr(function, "function_name"); n != "" {
			fmt.Printf("Name:          %s\n", n)
		}
		if d := mapStr(function, "description"); d != "" {
			fmt.Printf("Description:   %s\n", d)
		}
		if b, ok := function["public"].(bool); ok {
			fmt.Printf("Public:        %t\n", b)
		}

		// Display function code (truncated if very long) if present.
		if code := mapStr(function, "function_code"); code != "" {
			fmt.Printf("\nFunction Code:\n")
			if len(code) > 500 {
				fmt.Printf("%s\n... (truncated, %d total characters)\n", code[:500], len(code))
			} else {
				fmt.Printf("%s\n", code)
			}
		}
	} else {
		// JSON or CSV output — emit the raw passthrough document.
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"function_uuid", "function_name", "description", "public"}
		if err := formatter.FormatOutput(function, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
