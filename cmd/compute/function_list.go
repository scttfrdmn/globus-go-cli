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
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/compute"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	functionListLimit int
)

// FunctionListCmd represents the compute function list command
var FunctionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered Globus Compute functions",
	Long: `List functions you have registered with Globus Compute.

Examples:
  # List all your functions
  globus compute function list

  # Limit results
  globus compute function list --limit 50

  # JSON output for scripting
  globus compute function list --format json`,
	RunE: runFunctionList,
}

func init() {
	FunctionListCmd.Flags().IntVar(&functionListLimit, "limit", 25, "Maximum number of functions to return")
}

func runFunctionList(cmd *cobra.Command, args []string) error {
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

	// Build list options
	options := &compute.ListFunctionsOptions{
		PerPage: functionListLimit,
	}

	// List functions
	functionList, err := computeClient.ListFunctions(ctx, options)
	if err != nil {
		return fmt.Errorf("error listing functions: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable table
		if len(functionList.Functions) == 0 {
			fmt.Println("No functions found.")
			return nil
		}

		fmt.Printf("%-36s  %-30s  %-10s  %-10s\n", "Function ID", "Name", "Status", "Public")
		fmt.Printf("%s  %s  %s  %s\n",
			"------------------------------------",
			"------------------------------",
			"----------",
			"----------")

		for _, function := range functionList.Functions {
			name := function.Name
			if name == "" {
				name = "(unnamed)"
			}
			if len(name) > 30 {
				name = name[:27] + "..."
			}

			status := function.Status
			if status == "" {
				status = "active"
			}

			public := "No"
			if function.Public {
				public = "Yes"
			}

			fmt.Printf("%-36s  %-30s  %-10s  %-10s\n",
				function.ID,
				name,
				status,
				public)
		}

		fmt.Printf("\nTotal: %d function(s)\n", len(functionList.Functions))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"ID", "Name", "Description", "Status", "Public", "Owner", "CreatedAt"}
		if err := formatter.FormatOutput(functionList.Functions, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
