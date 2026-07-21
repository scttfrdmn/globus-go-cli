// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

// getClient builds a v4 Auth client authorized for the current profile from the
// per-resource-server GlobusApp token store. Mirrors the getClient helper in
// each service command package.
func getClient(ctx context.Context) (*auth.Client, error) {
	profile := viper.GetString("profile")

	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load client configuration: %w", err)
	}

	cfg, err := globusauth.AuthClientConfig(ctx, profile, clientCfg.ClientID, clientCfg.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("not logged in: %w", err)
	}

	client, err := auth.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth client: %w", err)
	}
	return client, nil
}
