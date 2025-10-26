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
	endpointListLimit  int
	endpointListSearch string
	endpointListStatus string
)

// EndpointListCmd represents the compute endpoint list command
var EndpointListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Globus Compute endpoints",
	Long: `List available Globus Compute endpoints.

This shows endpoints you own or have access to, including their connection
status and basic information.

Examples:
  # List all accessible endpoints
  globus compute endpoint list

  # Search for specific endpoints
  globus compute endpoint list --search "cluster"

  # Filter by status
  globus compute endpoint list --status online

  # JSON output for scripting
  globus compute endpoint list --format json`,
	RunE: runEndpointList,
}

func init() {
	EndpointListCmd.Flags().IntVar(&endpointListLimit, "limit", 25, "Maximum number of endpoints to return")
	EndpointListCmd.Flags().StringVar(&endpointListSearch, "search", "", "Search endpoints by name")
	EndpointListCmd.Flags().StringVar(&endpointListStatus, "status", "", "Filter by status")
}

func runEndpointList(cmd *cobra.Command, args []string) error {
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
	options := &compute.ListEndpointsOptions{
		PerPage:     endpointListLimit,
		IncludeInfo: true,
	}
	if endpointListSearch != "" {
		options.Search = endpointListSearch
	}
	if endpointListStatus != "" {
		options.FilterStatus = endpointListStatus
	}

	// List endpoints
	endpointList, err := computeClient.ListEndpoints(ctx, options)
	if err != nil {
		return fmt.Errorf("error listing endpoints: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable table
		if len(endpointList.Endpoints) == 0 {
			fmt.Println("No endpoints found.")
			return nil
		}

		fmt.Printf("%-36s  %-30s  %-10s  %-10s\n", "Endpoint ID", "Name", "Status", "Connected")
		fmt.Printf("%s  %s  %s  %s\n",
			"------------------------------------",
			"------------------------------",
			"----------",
			"----------")

		for _, endpoint := range endpointList.Endpoints {
			name := endpoint.Name
			if len(name) > 30 {
				name = name[:27] + "..."
			}

			status := endpoint.Status
			if status == "" {
				status = "unknown"
			}

			connected := "No"
			if endpoint.Connected {
				connected = "Yes"
			}

			fmt.Printf("%-36s  %-30s  %-10s  %-10s\n",
				endpoint.UUID,
				name,
				status,
				connected)
		}

		fmt.Printf("\nTotal: %d endpoint(s)\n", len(endpointList.Endpoints))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"UUID", "Name", "Description", "Status", "Connected", "Owner", "CreatedAt"}
		if err := formatter.FormatOutput(endpointList.Endpoints, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
