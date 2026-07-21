// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package collection implements the `globus collection` and `globus gcs`
// command trees over the v4 SDK's GCS Manager (gcs.CollectionClient).
//
// GCS management has a different auth model from the fixed-resource-server
// services: the client talks to a specific endpoint's GCS Manager host (its
// gcs_manager_url, discovered via the Transfer API), and management operations
// require that endpoint's dynamic `manage_collections` consent — which the
// normal `globus login` scope set does not include. getManagerClient therefore
// (1) resolves the manager URL from the endpoint document and (2) obtains a
// manage_collections authorizer, escalating consent (paste-code login) on first
// use.
package collection

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/gcs"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

// resolveManagerURL looks up the endpoint's GCS Manager base URL via the
// Transfer API. Returns an error if the endpoint is not a GCSv5 endpoint (no
// gcs_manager_url).
func resolveManagerURL(ctx context.Context, endpointID string) (string, error) {
	profile := viper.GetString("profile")
	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load client configuration: %w", err)
	}
	cfg, err := globusauth.ClientConfig(ctx, profile, clientCfg.ClientID, clientCfg.ClientSecret, globusauth.ServiceTransfer)
	if err != nil {
		return "", fmt.Errorf("not logged in: %w", err)
	}
	tc, err := transfer.NewClient(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create transfer client: %w", err)
	}
	ep, err := tc.GetEndpoint(ctx, endpointID)
	if err != nil {
		return "", fmt.Errorf("failed to look up endpoint %s: %w", endpointID, err)
	}
	if ep.GCSManagerURL == "" {
		return "", fmt.Errorf("endpoint %s is not a Globus Connect Server v5 endpoint (no gcs_manager_url); collection/gcs commands require GCSv5", endpointID)
	}
	return ep.GCSManagerURL, nil
}

// getManagerClient builds a GCS CollectionClient for managing the given
// endpoint. It resolves the manager URL from the endpoint document and obtains
// a manage_collections authorizer for the endpoint, escalating consent on first
// use (a browser/paste-code login prompt).
func getManagerClient(ctx context.Context, endpointID string) (*gcs.CollectionClient, error) {
	managerURL, err := resolveManagerURL(ctx, endpointID)
	if err != nil {
		return nil, err
	}

	profile := viper.GetString("profile")
	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Management operations use the endpoint's manage_collections scope (an
	// endpoint scope in URN format), keyed on the endpoint ID as resource
	// server. Escalate consent if we have no token for it yet.
	scope := gcs.EndpointManageCollectionsScope(endpointID)
	cfg, err := globusauth.ScopedClientConfig(ctx, profile, clientCfg.ClientID, clientCfg.ClientSecret, endpointID, scope, true)
	if err != nil {
		return nil, err
	}

	// The CollectionClient targets the endpoint's own GCS Manager; the
	// collection ID used for its default scope requirements is the endpoint ID
	// for management operations.
	client, err := gcs.NewCollectionClient(ctx, managerURL, endpointID, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS collection client: %w", err)
	}
	return client, nil
}
