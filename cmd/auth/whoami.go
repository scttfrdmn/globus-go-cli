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
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
)

// WhoamiCmd returns the whoami command
func WhoamiCmd() *cobra.Command {
	// whoamiCmd represents the whoami command
	whoamiCmd := &cobra.Command{
		Use:   "whoami",
		Short: "Display the current user",
		Long: `Display information about the current logged-in user.

This command shows details about your Globus identity based on your
current tokens.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return whoami(cmd)
		},
	}

	return whoamiCmd
}

// whoami handles the whoami command
func whoami(cmd *cobra.Command) error {
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

	// Create auth client directly
	authOptions := []auth.ClientOption{
		auth.WithClientID(clientCfg.ClientID),
		auth.WithClientSecret(clientCfg.ClientSecret),
	}
	
	authClient, err := auth.NewClient(authOptions...)
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	// Get the current user's identity
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	introspection, err := authClient.IntrospectToken(ctx, tokenInfo.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get user identity: %w", err)
	}

	// Print user information
	fmt.Println("\nCurrent User:")
	fmt.Printf("  Username: %s\n", introspection.Username)
	fmt.Printf("  Identity ID: %s\n", introspection.Sub)
	fmt.Printf("  Email: %s\n", introspection.Email)
	fmt.Printf("  Name: %s\n", introspection.Name)
	
	if len(introspection.IdentitiesSets) > 0 && len(introspection.IdentitiesSets[0]) > 0 {
		fmt.Println("  Linked Identities:")
		for _, idSet := range introspection.IdentitiesSets {
			for _, id := range idSet {
				fmt.Printf("    - %s\n", id)
			}
		}
	}
	
	// Print token information
	fmt.Println("\nToken Information:")
	fmt.Printf("  Expires At: %s\n", tokenInfo.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("  Expires In: %s\n", time.Until(tokenInfo.ExpiresAt).Round(time.Second))
	
	return nil
}