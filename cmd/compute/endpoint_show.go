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

// EndpointShowCmd represents the compute endpoint show command
var EndpointShowCmd = &cobra.Command{
	Use:   "show ENDPOINT_ID",
	Short: "Show details of a Globus Compute endpoint",
	Long: `Display detailed information about a specific compute endpoint.

This includes the endpoint's status, configuration, and metrics.

Examples:
  # Show endpoint details
  globus compute endpoint show ENDPOINT_ID

  # Show endpoint with JSON output
  globus compute endpoint show ENDPOINT_ID --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runEndpointShow,
}

func runEndpointShow(cmd *cobra.Command, args []string) error {
	endpointID := args[0]

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

	// Get endpoint
	endpoint, err := computeClient.GetEndpoint(ctx, endpointID)
	if err != nil {
		return fmt.Errorf("error getting endpoint: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable. The endpoint is an open-ended document;
		// render the well-known fields and fall back to keys where present.
		fmt.Printf("Endpoint Details\n")
		fmt.Printf("================\n\n")

		uuid := mapStr(endpoint, "uuid")
		if uuid == "" {
			uuid = mapStr(endpoint, "endpoint_id")
		}
		fmt.Printf("Endpoint ID:   %s\n", uuid)
		fmt.Printf("Name:          %s\n", mapStr(endpoint, "name"))
		if d := mapStr(endpoint, "description"); d != "" {
			fmt.Printf("Description:   %s\n", d)
		}
		fmt.Printf("Status:        %s\n", mapStr(endpoint, "status"))
		if b, ok := endpoint["connected"].(bool); ok {
			fmt.Printf("Connected:     %t\n", b)
		}
		if b, ok := endpoint["public"].(bool); ok {
			fmt.Printf("Public:        %t\n", b)
		}
		if t := mapStr(endpoint, "endpoint_type"); t != "" {
			fmt.Printf("Type:          %s\n", t)
		}
		if o := mapStr(endpoint, "owner"); o != "" {
			fmt.Printf("Owner:         %s\n", o)
		}
	} else {
		// JSON or CSV output — emit the raw passthrough document.
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"uuid", "name", "description", "status", "connected", "public", "endpoint_type", "owner"}
		if err := formatter.FormatOutput(endpoint, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
