// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/search"
)

// getClient builds a v4 Search client authorized for the current profile from
// the per-resource-server GlobusApp token store. It replaces the previous
// per-command recipe of loading the legacy token file and wrapping a static
// authorizer. On a missing/expired token it returns an error advising login.
func getClient(ctx context.Context) (*search.Client, error) {
	profile := viper.GetString("profile")

	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load client configuration: %w", err)
	}

	cfg, err := globusauth.ClientConfig(ctx, profile, clientCfg.ClientID, clientCfg.ClientSecret, globusauth.ServiceSearch)
	if err != nil {
		return nil, fmt.Errorf("not logged in: %w", err)
	}

	client, err := search.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create search client: %w", err)
	}
	return client, nil
}
