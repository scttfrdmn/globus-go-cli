// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package gcp implements the `globus gcp` command tree: management of Globus
// Connect Personal endpoints and their collections. Matching the Python globus
// CLI, these are CLOUD API operations (registering/updating GCP endpoints and
// guest collections via the Transfer API) — they do not install, start, or stop
// a local Globus Connect Personal agent. `gcp create mapped` prints a
// setup-key used to configure an installed agent.
package gcp

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

// getClient builds a v4 Transfer client authorized for the current profile.
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
