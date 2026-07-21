// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/compute"
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
	// The Compute endpoints API has no server-side paging or name search, so
	// these two flags are no-ops kept for backward compatibility. --status is
	// applied client-side.
	EndpointListCmd.Flags().IntVar(&endpointListLimit, "limit", 25, "Deprecated: the API does not paginate endpoints")
	EndpointListCmd.Flags().StringVar(&endpointListSearch, "search", "", "Deprecated: the API has no endpoint name search")
	EndpointListCmd.Flags().StringVar(&endpointListStatus, "status", "", "Filter by status (applied client-side)")
	_ = EndpointListCmd.Flags().MarkDeprecated("limit", "the API does not paginate endpoints")
	_ = EndpointListCmd.Flags().MarkDeprecated("search", "the API has no endpoint name search")
}

func runEndpointList(cmd *cobra.Command, args []string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Compute client authorized for the current profile.
	computeClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// The Compute API returns a top-level array of endpoint documents. It has no
	// server-side search/status/per-page filters, so those flags are applied
	// client-side where possible. Only the "role" query param is supported.
	endpoints, err := computeClient.GetEndpoints(ctx, &compute.GetEndpointsOptions{Role: "owner"})
	if err != nil {
		return fmt.Errorf("error listing endpoints: %w", err)
	}

	// Optional client-side status filter.
	if endpointListStatus != "" {
		filtered := endpoints[:0]
		for _, ep := range endpoints {
			if mapStr(ep, "status") == endpointListStatus {
				filtered = append(filtered, ep)
			}
		}
		endpoints = filtered
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable table
		if len(endpoints) == 0 {
			fmt.Println("No endpoints found.")
			return nil
		}

		fmt.Printf("%-36s  %-30s  %-10s  %-10s\n", "Endpoint ID", "Name", "Status", "Connected")
		fmt.Printf("%s  %s  %s  %s\n",
			"------------------------------------",
			"------------------------------",
			"----------",
			"----------")

		for _, endpoint := range endpoints {
			name := mapStr(endpoint, "name")
			if len(name) > 30 {
				name = name[:27] + "..."
			}

			status := mapStr(endpoint, "status")
			if status == "" {
				status = "unknown"
			}

			connected := "No"
			if b, ok := endpoint["connected"].(bool); ok && b {
				connected = "Yes"
			}

			uuid := mapStr(endpoint, "uuid")
			if uuid == "" {
				uuid = mapStr(endpoint, "endpoint_id")
			}

			fmt.Printf("%-36s  %-30s  %-10s  %-10s\n",
				uuid,
				name,
				status,
				connected)
		}

		fmt.Printf("\nTotal: %d endpoint(s)\n", len(endpoints))
	} else {
		// JSON or CSV output — emit the raw passthrough documents.
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"uuid", "name", "description", "status", "connected", "owner"}
		if err := formatter.FormatOutput(endpoints, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}

// mapStr returns the string value at key in m, or "" if absent/not a string.
func mapStr(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
