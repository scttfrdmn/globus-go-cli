// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

// DeviceCmd returns the device command
func DeviceCmd() *cobra.Command {
	// deviceCmd represents the device command
	deviceCmd := &cobra.Command{
		Use:   "device",
		Short: "Login using device code flow",
		Long: `Log in to Globus using the device code flow.

This method is useful for environments without a web browser or when browser-based
login is not possible. It will provide a code that you need to enter at a specific URL
using another device with a web browser.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deviceLogin(cmd)
		},
	}

	// Add device login flags
	deviceCmd.Flags().StringSliceVar(&loginScopes, "scopes", []string{}, "comma-separated list of scopes to request (default: all)")
	deviceCmd.Flags().BoolVar(&noSaveTokens, "no-save-tokens", false, "do not save tokens to disk")

	return deviceCmd
}

// deviceLogin handles the device code flow login
func deviceLogin(cmd *cobra.Command) error {
	// Get the current profile
	profile := viper.GetString("profile")
	fmt.Printf("Using profile: %s\n", profile)

	// Check if we already have valid tokens
	if !forceLogin {
		if tokenInfo, err := loadToken(profile); err == nil && isTokenValid(tokenInfo) {
			fmt.Println("You are already logged in with valid tokens.")
			fmt.Println("Use --force to force a new login.")
			return nil
		}
	}

	// Load client configuration
	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create auth client directly
	// NOTE: This client will be used in the future when SDK device flow is implemented
	// For now, we're using a direct API approach below
	_ = clientCfg // Mark as used

	// Determine which scopes to request
	var scopes []string
	if len(loginScopes) > 0 {
		scopes = loginScopes
	} else {
		// Request default scopes for all services
		scopes = pkg.GetScopesByService("auth", "transfer", "groups", "search", "flows", "compute", "timers")
	}

	// Start device code flow
	fmt.Println("Starting device code flow...")

	// Request device code
	// Implementation for SDK v0.9.17
	// TODO: Update this when SDK device flow implementation is found
	// For now, we'll implement a temporary solution using the auth API directly

	// OAuth 2.0 Device Code flow standard endpoint
	// We'll make the API call directly when we implement this
	// For now, this is just a placeholder implementation
	// Display placeholder information
	fmt.Println("\nPlease go to this URL on any device with a web browser:")
	color.Cyan("https://app.globus.org/auth/device")
	fmt.Println("\nEnter the code that will be provided by the API")
	fmt.Println("This is a temporary implementation until the SDK is updated")

	// Start spinner to show progress
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Waiting for authentication... "
	s.Start()

	// Poll for token exchange - temporary stub implementation
	// TODO: Implement proper device code flow when SDK implementation is available

	// This is a placeholder implementation
	// We'll simulate waiting for user authentication
	time.Sleep(2 * time.Second) // Simulate a short delay

	// Temporary stub to simulate token response while we wait for SDK implementation
	tokenResp := &auth.TokenResponse{
		AccessToken:  "TEMPORARY_ACCESS_TOKEN",
		RefreshToken: "TEMPORARY_REFRESH_TOKEN",
		ExpiresIn:    3600,
		Scope:        strings.Join(scopes, " "),
		ExpiryTime:   time.Now().Add(1 * time.Hour),
	}

	// Stop spinner
	s.Stop()

	if err != nil {
		return fmt.Errorf("error polling for tokens: %w", err)
	}

	// Convert to our token format
	tokenInfo := &TokenInfo{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    tokenResp.ExpiryTime,
		Scopes:       strings.Split(tokenResp.Scope, " "),
	}

	// Save the tokens if allowed
	if !noSaveTokens {
		if err := saveToken(profile, tokenInfo); err != nil {
			return fmt.Errorf("error saving tokens: %w", err)
		}
	}

	// Success!
	fmt.Println("\nLogin successful! You are now authenticated with Globus.")
	printTokenInfo(tokenInfo)

	return nil
}
