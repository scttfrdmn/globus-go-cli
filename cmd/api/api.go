// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package api implements the top-level "globus api" raw-passthrough command
// group. It mirrors the Python Globus CLI's `globus api <service> <METHOD>
// <PATH>`: each subcommand issues an authenticated raw HTTP request to a Globus
// service API and prints the JSON response. It is an escape hatch for calling
// endpoints the typed commands do not yet cover.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// serviceSpec describes one "globus api <service>" subcommand: the CLI name,
// the globusauth.Service used to mint a token, and the base URL relative paths
// resolve against. Base URLs match the v4 SDK's service defaults. For auth,
// groups, and search the base already includes a version segment (/v2, /v1);
// PATH is joined onto it as given.
type serviceSpec struct {
	name    string
	svc     globusauth.Service
	baseURL string
}

// specs is the fixed set of raw-passthrough subcommands, one per service,
// mirroring the Python CLI (note: "groups" and "timer" naming).
var specs = []serviceSpec{
	{name: "auth", svc: globusauth.ServiceAuth, baseURL: "https://auth.globus.org/v2"},
	{name: "transfer", svc: globusauth.ServiceTransfer, baseURL: "https://transfer.api.globus.org"},
	{name: "groups", svc: globusauth.ServiceGroups, baseURL: "https://groups.api.globus.org/v2"},
	{name: "search", svc: globusauth.ServiceSearch, baseURL: "https://search.api.globus.org/v1"},
	{name: "flows", svc: globusauth.ServiceFlows, baseURL: "https://flows.globus.org"},
	{name: "timer", svc: globusauth.ServiceTimers, baseURL: "https://timer.automate.globus.org"},
	{name: "compute", svc: globusauth.ServiceCompute, baseURL: "https://compute.api.globus.org"},
}

// APICmd returns the "api" command group with one raw-passthrough subcommand
// per Globus service. It is exported for wiring into the root command.
func APICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Make authenticated raw HTTP requests to Globus service APIs",
		Long: `Make authenticated raw HTTP requests to Globus service APIs.

Each subcommand targets one service and takes an HTTP METHOD and a PATH,
issuing the request with the current profile's stored token for that service
and printing the JSON response. This is an escape hatch for endpoints not yet
covered by the typed commands.

Examples:
  globus api transfer GET /endpoint_search --query filter_fulltext=example
  globus api groups GET /groups/GROUP_ID
  globus api flows POST /flows --body '{"title":"my flow"}'`,
	}

	for _, spec := range specs {
		cmd.AddCommand(newServiceCmd(spec))
	}

	return cmd
}

// newServiceCmd builds the raw-passthrough subcommand for a single service.
func newServiceCmd(spec serviceSpec) *cobra.Command {
	var (
		body        string
		queryParams []string
	)

	c := &cobra.Command{
		Use:   fmt.Sprintf("%s METHOD PATH", spec.name),
		Short: fmt.Sprintf("Make a raw HTTP request to the Globus %s API", spec.name),
		Long: fmt.Sprintf(`Make a raw, authenticated HTTP request to the Globus %s API.

METHOD is an HTTP verb (GET, POST, PUT, PATCH, DELETE). PATH is the request
path relative to %s (a leading slash is optional; any version prefix in the
base URL is preserved). The JSON response is printed to stdout.`, spec.name, spec.baseURL),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			method := strings.ToUpper(args[0])
			rawPath := args[1]

			profile := viper.GetString("profile")
			clientCfg, err := config.LoadClientConfig()
			if err != nil {
				return fmt.Errorf("failed to load client configuration: %w", err)
			}

			cfg, err := globusauth.ClientConfig(ctx, profile, clientCfg.ClientID, clientCfg.ClientSecret, spec.svc)
			if err != nil {
				return fmt.Errorf("not logged in: %w", err)
			}
			cfg.BaseURL = spec.baseURL

			client, err := core.NewClient(cfg)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			// Separate any query string embedded in PATH from the path itself,
			// then merge in the repeated --query key=value entries.
			path := rawPath
			query := url.Values{}
			if i := strings.IndexByte(rawPath, '?'); i >= 0 {
				path = rawPath[:i]
				embedded, perr := url.ParseQuery(rawPath[i+1:])
				if perr != nil {
					return fmt.Errorf("invalid query string in path: %w", perr)
				}
				for k, vs := range embedded {
					for _, v := range vs {
						query.Add(k, v)
					}
				}
			}
			for _, q := range queryParams {
				k, v, found := strings.Cut(q, "=")
				if !found {
					return fmt.Errorf("invalid --query %q: expected key=value", q)
				}
				query.Add(k, v)
			}
			if len(query) == 0 {
				query = nil
			}

			// Parse --body into a generic value so DoRequest JSON-encodes it.
			var reqBody interface{}
			if body != "" {
				if err := json.Unmarshal([]byte(body), &reqBody); err != nil {
					return fmt.Errorf("invalid --body JSON: %w", err)
				}
			}

			var result interface{}
			if err := client.DoRequest(ctx, method, path, query, reqBody, &result); err != nil {
				return fmt.Errorf("request failed: %w", err)
			}

			formatter := output.NewFormatter(viper.GetString("format"), cmd.OutOrStdout())
			return formatter.FormatOutput(result, nil)
		},
	}

	c.Flags().StringVar(&body, "body", "", "Raw JSON request body")
	c.Flags().StringArrayVar(&queryParams, "query", nil, "Query parameter as key=value (repeatable)")

	return c
}
