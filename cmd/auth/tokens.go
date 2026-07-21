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

// TokensCmd returns the tokens command
func TokensCmd() *cobra.Command {
	// tokensCmd represents the tokens command
	tokensCmd := &cobra.Command{
		Use:   "tokens",
		Short: "Commands for Globus Auth tokens",
		Long: `Commands for working with Globus Auth tokens.

This command group provides subcommands for managing your Globus Auth
tokens including listing, viewing details, and revoking tokens.`,
	}

	// Add subcommands
	tokensCmd.AddCommand(
		tokensShowCmd(),
		tokensRevokeCmd(),
		tokensIntrospectCmd(),
	)

	return tokensCmd
}

// tokensShowCmd returns the tokens show command
func tokensShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show token information",
		Long: `Show information about the stored Globus Auth tokens.

This command lists the stored tokens (one per resource server) for the
current profile, including when each expires and its scopes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the current profile
			profile := viper.GetString("profile")
			fmt.Printf("Using profile: %s\n", profile)

			tokens, err := globusauth.AllTokens(profile)
			if err != nil {
				return fmt.Errorf("not logged in: %w", err)
			}
			if len(tokens) == 0 {
				fmt.Println("No stored tokens.")
				return nil
			}

			for _, td := range tokens {
				fmt.Printf("\nResource Server: %s\n", td.ResourceServer)
				fmt.Printf("  Expires At: %s\n", td.ExpiresAt.Format(time.RFC3339))
				fmt.Printf("  Expires In: %s\n", time.Until(td.ExpiresAt).Round(time.Second))
				if td.Scope != "" {
					fmt.Printf("  Scopes: %s\n", td.Scope)
				}
				fmt.Printf("  Has Refresh Token: %t\n", td.RefreshToken != "")
			}
			return nil
		},
	}
}

// newRevokeAuthClient builds an auth client authenticated with client Basic
// auth, used for the token revocation endpoint.
func newRevokeAuthClient(ctx context.Context) (*auth.Client, error) {
	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load client configuration: %w", err)
	}
	clientID := clientCfg.ClientID
	if clientID == "" {
		clientID = globusauth.DefaultClientID
	}
	return auth.NewClient(ctx, &core.Config{
		Authorizer: authorizers.NewBasicAuthAuthorizer(clientID, clientCfg.ClientSecret),
	})
}

// tokensRevokeCmd returns the tokens revoke command
func tokensRevokeCmd() *cobra.Command {
	var tokenType string

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a token",
		Long: `Revoke Globus Auth tokens.

This command revokes the access token, refresh token, or all tokens for the
Auth resource server, invalidating them with Globus Auth. Use 'globus logout'
to revoke and remove tokens for every service.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the current profile
			profile := viper.GetString("profile")
			fmt.Printf("Using profile: %s\n", profile)

			if tokenType == "all" {
				return logout(cmd)
			}

			// Load the stored auth token.
			td, err := globusauth.TokenFor(profile, globusauth.ServiceAuth)
			if err != nil {
				return fmt.Errorf("not logged in: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			authClient, err := newRevokeAuthClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to create auth client: %w", err)
			}

			switch tokenType {
			case "access":
				fmt.Println("Revoking access token...")
				if err := authClient.RevokeToken(ctx, td.AccessToken); err != nil {
					return fmt.Errorf("failed to revoke access token: %w", err)
				}
				fmt.Println("Access token revoked successfully")
			case "refresh":
				if td.RefreshToken == "" {
					return fmt.Errorf("no refresh token available")
				}
				fmt.Println("Revoking refresh token...")
				if err := authClient.RevokeToken(ctx, td.RefreshToken); err != nil {
					return fmt.Errorf("failed to revoke refresh token: %w", err)
				}
				fmt.Println("Refresh token revoked successfully")
			default:
				return fmt.Errorf("invalid token type: %s. Must be 'access', 'refresh', or 'all'", tokenType)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&tokenType, "type", "access", "Token type to revoke (access, refresh, all)")

	return cmd
}

// tokensIntrospectCmd returns the tokens introspect command
func tokensIntrospectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "introspect",
		Short: "Introspect the current token",
		Long: `Introspect the current Globus Auth token.

This command shows detailed information about your current Auth access token
by introspecting it with Globus Auth.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the current profile
			profile := viper.GetString("profile")
			fmt.Printf("Using profile: %s\n", profile)

			// Load the stored auth token.
			td, err := globusauth.TokenFor(profile, globusauth.ServiceAuth)
			if err != nil {
				return fmt.Errorf("not logged in: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Introspection authenticates the client with Basic auth.
			authClient, err := newRevokeAuthClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to create auth client: %w", err)
			}

			introspection, err := authClient.IntrospectToken(ctx, td.AccessToken, nil)
			if err != nil {
				return fmt.Errorf("failed to introspect token: %w", err)
			}

			fmt.Println("\nToken Introspection:")
			fmt.Printf("  Active: %t\n", introspection.Active)
			fmt.Printf("  Scope: %s\n", introspection.Scope)
			fmt.Printf("  Client ID: %s\n", introspection.ClientID)
			fmt.Printf("  Username: %s\n", introspection.Username)
			fmt.Printf("  Email: %s\n", introspection.Email)
			fmt.Printf("  Subject: %s\n", introspection.Sub)
			fmt.Printf("  Expires At: %d\n", introspection.Exp)

			if len(introspection.IdentitySet) > 0 {
				fmt.Println("  Identity Set:")
				for _, identity := range introspection.IdentitySet {
					fmt.Printf("    - %s\n", identity)
				}
			}

			return nil
		},
	}
}
