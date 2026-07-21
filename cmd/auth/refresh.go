// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/tokenstorage"
)

// RefreshCmd returns the refresh command
func RefreshCmd() *cobra.Command {
	// refreshCmd represents the refresh command
	refreshCmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh access tokens",
		Long: `Refresh your Globus access tokens using the refresh token.

This command refreshes the access token for the Auth resource server without
requiring you to log in again, and stores the result. It requires a valid
refresh token. (Service clients also refresh automatically on demand.)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return refreshToken(cmd)
		},
	}

	return refreshCmd
}

// refreshToken handles refreshing the Auth access token and storing it back.
func refreshToken(cmd *cobra.Command) error {
	// Get the current profile
	profile := viper.GetString("profile")
	fmt.Printf("Using profile: %s\n", profile)

	// Load the stored auth token (carries the refresh token).
	td, err := globusauth.TokenFor(profile, globusauth.ServiceAuth)
	if err != nil {
		return fmt.Errorf("failed to load token: %w", err)
	}
	if td.RefreshToken == "" {
		return fmt.Errorf("no refresh token available for profile %s, please log in again", profile)
	}

	// Load client configuration.
	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}
	clientID := clientCfg.ClientID
	if clientID == "" {
		clientID = globusauth.DefaultClientID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// The token endpoint authenticates the client with Basic auth.
	authClient, err := auth.NewClient(ctx, &core.Config{
		Authorizer: authorizers.NewBasicAuthAuthorizer(clientID, clientCfg.ClientSecret),
	})
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	// Refresh the token.
	fmt.Println("Refreshing access token...")
	tokenResp, err := authClient.RefreshToken(ctx, td.RefreshToken, clientID, clientCfg.ClientSecret)
	if err != nil {
		return fmt.Errorf("error refreshing token: %w", err)
	}

	// Persist the refreshed token back to the store. Keep the prior refresh
	// token if the response did not include a new one.
	refresh := tokenResp.RefreshToken
	if refresh == "" {
		refresh = td.RefreshToken
	}
	store, err := globusauth.Store(profile)
	if err != nil {
		return err
	}
	updated := &tokenstorage.TokenData{
		ResourceServer: td.ResourceServer,
		IdentityID:     td.IdentityID,
		Scope:          tokenResp.Scope,
		AccessToken:    tokenResp.AccessToken,
		RefreshToken:   refresh,
		ExpiresAt:      time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		TokenType:      tokenResp.TokenType,
	}
	if updated.Scope == "" {
		updated.Scope = td.Scope
	}
	if err := store.Store(updated); err != nil {
		return fmt.Errorf("error saving updated token: %w", err)
	}

	fmt.Println("\nToken refresh successful!")
	expiresIn := time.Until(updated.ExpiresAt).Round(time.Second)
	color.Green("Token refreshed and valid for %s", expiresIn)

	return nil
}
