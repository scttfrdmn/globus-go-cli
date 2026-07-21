// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

// getClient builds a v4 Transfer client authorized for the current profile from
// the per-resource-server GlobusApp token store. It replaces the previous
// per-command recipe of loading the legacy token file and wrapping a static
// authorizer. On a missing/expired token it returns an error advising login.
func getClient(ctx context.Context) (*transfer.Client, error) {
	profile := viper.GetString("profile")

	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load client configuration: %w", err)
	}

	cfg, err := globusauth.ClientConfig(ctx, profile, clientCfg.ClientID, clientCfg.ClientSecret, globusauth.ServiceTransfer)
	if err != nil {
		return nil, fmt.Errorf("not logged in: %w", err)
	}

	client, err := transfer.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer client: %w", err)
	}
	return client, nil
}
