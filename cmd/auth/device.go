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
)

// DeviceCmd returns the device command
func DeviceCmd() *cobra.Command {
	// deviceCmd represents the device command
	deviceCmd := &cobra.Command{
		Use:   "device",
		Short: "Login using device code flow",
		Long: `Log in to Globus using the OAuth2 device code flow.

This method is useful for environments without a web browser or when browser-based
login is not possible. It provides a code to enter at a URL using another device
with a web browser.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deviceLogin(cmd)
		},
	}

	// Add device login flags
	deviceCmd.Flags().StringSliceVar(&loginScopes, "scopes", []string{}, "comma-separated list of scopes to request (default: all)")
	deviceCmd.Flags().BoolVar(&noSaveTokens, "no-save-tokens", false, "do not save tokens to disk")

	return deviceCmd
}

// deviceLogin handles the device code flow login using the v4 SDK.
func deviceLogin(cmd *cobra.Command) error {
	// Get the current profile
	profile := viper.GetString("profile")
	fmt.Printf("Using profile: %s\n", profile)

	// Check if we already have valid tokens for the transfer resource server.
	if !forceLogin {
		if _, err := globusauth.TokenFor(profile, globusauth.ServiceTransfer); err == nil {
			fmt.Println("You are already logged in with valid tokens.")
			fmt.Println("Use --force to force a new login.")
			return nil
		}
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

	// Determine which scopes to request (default: all services).
	scopes := loginScopes
	if len(scopes) == 0 {
		for _, svc := range globusauth.AllServices {
			if s, ok := globusauth.Scope(svc); ok {
				scopes = append(scopes, s)
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// The device authorization endpoints authenticate the client with Basic auth.
	authClient, err := auth.NewClient(ctx, &core.Config{
		Authorizer: authorizers.NewBasicAuthAuthorizer(clientID, clientCfg.ClientSecret),
	})
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	fmt.Println("Starting device code flow...")
	deviceResp, err := authClient.StartDeviceAuthorization(ctx, clientID, scopes)
	if err != nil {
		return fmt.Errorf("failed to start device authorization: %w", err)
	}

	// Prompt the user to visit the verification URL and enter the code.
	fmt.Println("\nPlease go to this URL on any device with a web browser:")
	if deviceResp.VerificationURIComplete != "" {
		color.Cyan("  %s", deviceResp.VerificationURIComplete)
		fmt.Println("\n(or enter the code below at the base URL)")
	}
	color.Cyan("  %s", deviceResp.VerificationURI)
	fmt.Printf("\nEnter this code when prompted: ")
	color.Green("%s", deviceResp.UserCode)
	fmt.Println("\nWaiting for authorization...")

	// Poll until the user completes authorization (or the code expires).
	tokenResp, err := authClient.WaitForDeviceAuthorization(ctx, clientID, deviceResp)
	if err != nil {
		return fmt.Errorf("device authorization failed: %w", err)
	}

	// Persist the returned tokens (primary + per-resource-server other_tokens).
	if !noSaveTokens {
		now := time.Now()
		toStore := []globusauth.StoredToken{tokenResponseToStored(tokenResp)}
		for i := range tokenResp.OtherTokens {
			toStore = append(toStore, tokenResponseToStored(&tokenResp.OtherTokens[i]))
		}
		if err := globusauth.StoreTokens(profile, now, toStore...); err != nil {
			return fmt.Errorf("error saving tokens: %w", err)
		}
	}

	fmt.Println("\nLogin successful! You are now authenticated with Globus.")
	return nil
}

// tokenResponseToStored maps a v4 auth TokenResponse to a globusauth.StoredToken.
func tokenResponseToStored(tr *auth.TokenResponse) globusauth.StoredToken {
	return globusauth.StoredToken{
		ResourceServer: tr.ResourceServer,
		AccessToken:    tr.AccessToken,
		RefreshToken:   tr.RefreshToken,
		Scope:          tr.Scope,
		ExpiresIn:      tr.ExpiresIn,
		TokenType:      tr.TokenType,
	}
}
