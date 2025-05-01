// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
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

			// Create SDK config
			sdkConfig := pkg.NewConfig().
				WithClientID(clientCfg.ClientID).
				WithClientSecret(clientCfg.ClientSecret)

			// Create auth client
			authClient := sdkConfig.NewAuthClient()

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

			// Create SDK config
			sdkConfig := pkg.NewConfig().
				WithClientID(clientCfg.ClientID).
				WithClientSecret(clientCfg.ClientSecret)

			// Create auth client
			authClient := sdkConfig.NewAuthClient()

			// Introspect the token
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			
			introspection, err := authClient.IntrospectToken(ctx, tokenInfo.AccessToken)
			if err != nil {
				return fmt.Errorf("failed to introspect token: %w", err)
			}

			// Print token introspection
			fmt.Println("\nToken Introspection:")
			fmt.Printf("  Active: %t\n", introspection.Active)
			fmt.Printf("  Scope: %s\n", introspection.Scope)
			fmt.Printf("  Client ID: %s\n", introspection.ClientID)
			fmt.Printf("  Username: %s\n", introspection.Username)
			fmt.Printf("  Email: %s\n", introspection.Email)
			fmt.Printf("  Subject: %s\n", introspection.Sub)
			fmt.Printf("  Expires At: %d\n", introspection.Exp)
			
			if len(introspection.Aud) > 0 {
				fmt.Println("  Audiences:")
				for _, aud := range introspection.Aud {
					fmt.Printf("    - %s\n", aud)
				}
			}
			
			return nil
		},
	}
}