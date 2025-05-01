// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
)

// LogoutCmd returns the logout command
func LogoutCmd() *cobra.Command {
	// logoutCmd represents the logout command
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from Globus",
		Long: `Log out from Globus by revoking your tokens and removing saved credentials.

This command will revoke your access and refresh tokens, removing your credentials
from the local system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return logout(cmd)
		},
	}

	return logoutCmd
}

// logout handles the logout command
func logout(cmd *cobra.Command) error {
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

	// Revoke the access token
	fmt.Println("Revoking access token...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := authClient.RevokeToken(ctx, tokenInfo.AccessToken); err != nil {
		fmt.Printf("Warning: Failed to revoke access token: %v\n", err)
	}

	// Revoke the refresh token if present
	if tokenInfo.RefreshToken != "" {
		fmt.Println("Revoking refresh token...")
		if err := authClient.RevokeToken(ctx, tokenInfo.RefreshToken); err != nil {
			fmt.Printf("Warning: Failed to revoke refresh token: %v\n", err)
		}
	}

	// Get the token file path
	tokenFile, err := getTokenFilePath(profile)
	if err != nil {
		return fmt.Errorf("failed to get token file path: %w", err)
	}

	// Delete the token file
	if err := os.Remove(tokenFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}

	fmt.Println("Logged out successfully!")
	return nil
}