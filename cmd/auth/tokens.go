// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
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
		Long: `Show information about the current Globus Auth tokens.

This command displays details about your current access and refresh tokens
including when they expire and what scopes they have.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the current profile
			profile := viper.GetString("profile")
			fmt.Printf("Using profile: %s\n", profile)

			// Load the token
			tokenInfo, err := loadToken(profile)
			if err != nil {
				return fmt.Errorf("not logged in: %w", err)
			}

			// Print token information
			printTokenInfo(tokenInfo)
			return nil
		},
	}
}

// tokensRevokeCmd returns the tokens revoke command
func tokensRevokeCmd() *cobra.Command {
	var tokenType string

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a token",
		Long: `Revoke a Globus Auth token.

This command revokes either your access token, refresh token, or both,
invalidating them with Globus Auth.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the current profile
			profile := viper.GetString("profile")
			fmt.Printf("Using profile: %s\n", profile)

			// Load the token
			tokenInfo, err := loadToken(profile)
			if err != nil {
				return fmt.Errorf("not logged in: %w", err)
			}

			// Load client configuration
			clientCfg, err := config.LoadClientConfig()
			if err != nil {
				return fmt.Errorf("failed to load client configuration: %w", err)
			}

			// Create auth client - SDK v0.9.10 compatibility
			authOptions := []auth.ClientOption{
				auth.WithClientID(clientCfg.ClientID),
				auth.WithClientSecret(clientCfg.ClientSecret),
			}
			
			authClient, err := auth.NewClient(authOptions...)
			if err != nil {
				return fmt.Errorf("failed to create auth client: %w", err)
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Revoke the specified token type
			switch tokenType {
			case "access":
				fmt.Println("Revoking access token...")
				if err := authClient.RevokeToken(ctx, tokenInfo.AccessToken); err != nil {
					return fmt.Errorf("failed to revoke access token: %w", err)
				}
				fmt.Println("Access token revoked successfully")
			case "refresh":
				if tokenInfo.RefreshToken == "" {
					return fmt.Errorf("no refresh token available")
				}
				fmt.Println("Revoking refresh token...")
				if err := authClient.RevokeToken(ctx, tokenInfo.RefreshToken); err != nil {
					return fmt.Errorf("failed to revoke refresh token: %w", err)
				}
				fmt.Println("Refresh token revoked successfully")
			case "all":
				return logout(cmd)
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

This command shows detailed information about your current access token
by introspecting it with Globus Auth.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the current profile
			profile := viper.GetString("profile")
			fmt.Printf("Using profile: %s\n", profile)

			// Load the token
			tokenInfo, err := loadToken(profile)
			if err != nil {
				return fmt.Errorf("not logged in: %w", err)
			}

			// Check if the token is valid
			if !isTokenValid(tokenInfo) {
				return fmt.Errorf("token is expired, please login again")
			}

			// Load client configuration
			clientCfg, err := config.LoadClientConfig()
			if err != nil {
				return fmt.Errorf("failed to load client configuration: %w", err)
			}

			// Create auth client - SDK v0.9.10 compatibility
			authOptions := []auth.ClientOption{
				auth.WithClientID(clientCfg.ClientID),
				auth.WithClientSecret(clientCfg.ClientSecret),
			}
			
			authClient, err := auth.NewClient(authOptions...)
			if err != nil {
				return fmt.Errorf("failed to create auth client: %w", err)
			}

			// Introspect the token
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			
			introspection, err := authClient.IntrospectToken(ctx, tokenInfo.AccessToken)
			if err != nil {
				return fmt.Errorf("failed to introspect token: %w", err)
			}

			// Print token introspection - SDK v0.9.10 compatibility
			// Field names may have changed in the TokenInfo struct
			fmt.Println("\nToken Introspection:")
			fmt.Printf("  Active: %t\n", introspection.Active)
			fmt.Printf("  Scope: %s\n", introspection.Scope)
			fmt.Printf("  Client ID: %s\n", introspection.ClientID)
			fmt.Printf("  Username: %s\n", introspection.Username)
			fmt.Printf("  Email: %s\n", introspection.Email)
			
			// Updated field names for Subject in v0.9.10
			fmt.Printf("  Subject: %s\n", introspection.Subject)
			fmt.Printf("  Expires At: %d\n", introspection.Exp)
			
			// IdentitySet field in v0.9.10 replaces Audiences
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