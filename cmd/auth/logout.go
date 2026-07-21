// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

// LogoutCmd returns the logout command
func LogoutCmd() *cobra.Command {
	// logoutCmd represents the logout command
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from Globus",
		Long: `Log out from Globus by revoking your tokens and removing saved credentials.

This command will revoke your access and refresh tokens for every service,
removing your credentials from the local system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return logout(cmd)
		},
	}

	return logoutCmd
}

// logout handles the logout command. It revokes every stored token (one per
// resource server) with Globus Auth, then removes them from local storage.
func logout(cmd *cobra.Command) error {
	// Get the current profile
	profile := viper.GetString("profile")
	fmt.Printf("Using profile: %s\n", profile)

	// Gather all stored tokens across resource servers.
	tokens, err := globusauth.AllTokens(profile)
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}
	if len(tokens) == 0 {
		fmt.Println("No stored credentials; already logged out.")
		return nil
	}

	// Build an unauthenticated auth client for revocation (revoke only needs the
	// client identity, which the config carries).
	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// The revoke endpoint authenticates the client with HTTP Basic auth
	// (client_id:client_secret); native/public clients use an empty secret.
	clientID := clientCfg.ClientID
	if clientID == "" {
		clientID = globusauth.DefaultClientID
	}
	authClient, err := auth.NewClient(ctx, &core.Config{
		Authorizer: authorizers.NewBasicAuthAuthorizer(clientID, clientCfg.ClientSecret),
	})
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	// Revoke access + refresh tokens for each resource server. Warn but continue
	// on failures so local cleanup still happens.
	for _, td := range tokens {
		if td.AccessToken != "" {
			if rerr := authClient.RevokeToken(ctx, td.AccessToken); rerr != nil {
				fmt.Printf("Warning: failed to revoke access token for %s: %v\n", td.ResourceServer, rerr)
			}
		}
		if td.RefreshToken != "" {
			if rerr := authClient.RevokeToken(ctx, td.RefreshToken); rerr != nil {
				fmt.Printf("Warning: failed to revoke refresh token for %s: %v\n", td.ResourceServer, rerr)
			}
		}
	}

	// Remove all stored tokens locally.
	if err := globusauth.RemoveAllTokens(profile); err != nil {
		return fmt.Errorf("failed to remove stored tokens: %w", err)
	}

	fmt.Println("Logged out successfully!")
	return nil
}
