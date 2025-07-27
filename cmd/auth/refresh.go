// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
)

// RefreshCmd returns the refresh command
func RefreshCmd() *cobra.Command {
	// refreshCmd represents the refresh command
	refreshCmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh access tokens",
		Long: `Refresh your Globus access tokens using the refresh token.

This command refreshes the access token for your current profile
without requiring you to log in again. It requires a valid refresh token.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return refreshToken(cmd)
		},
	}

	return refreshCmd
}

// refreshToken handles refreshing the access token
func refreshToken(cmd *cobra.Command) error {
	// Get the current profile
	profile := viper.GetString("profile")
	fmt.Printf("Using profile: %s\n", profile)

	// Load the current token
	tokenInfo, err := loadToken(profile)
	if err != nil {
		return fmt.Errorf("failed to load token: %w", err)
	}

	// Check if we have a refresh token
	if tokenInfo.RefreshToken == "" {
		return fmt.Errorf("no refresh token available for profile %s, please log in again", profile)
	}

	// Load client configuration
	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create auth client - SDK v0.9.17 compatibility
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

	// Refresh the token
	fmt.Println("Refreshing access token...")
	tokenResp, err := authClient.RefreshToken(ctx, tokenInfo.RefreshToken)
	if err != nil {
		return fmt.Errorf("error refreshing token: %w", err)
	}

	// Convert to our token format
	newTokenInfo := &TokenInfo{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    tokenResp.ExpiryTime,
		Scopes:       strings.Split(tokenResp.Scope, " "),
	}

	// If no new refresh token was provided, keep the old one
	if newTokenInfo.RefreshToken == "" {
		newTokenInfo.RefreshToken = tokenInfo.RefreshToken
	}

	// Save the updated token
	if err := saveToken(profile, newTokenInfo); err != nil {
		return fmt.Errorf("error saving updated token: %w", err)
	}

	// Success!
	fmt.Println("\nToken refresh successful!")
	printTokenInfo(newTokenInfo)

	// Show token validity duration
	expiresIn := time.Until(newTokenInfo.ExpiresAt).Round(time.Second)
	color.Green("Token refreshed and valid for %s", expiresIn)

	return nil
}
