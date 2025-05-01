// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
)

// IdentitiesCmd returns the identities command
func IdentitiesCmd() *cobra.Command {
	// identitiesCmd represents the identities command
	identitiesCmd := &cobra.Command{
		Use:   "identities",
		Short: "Commands for Globus Auth identities",
		Long: `Commands for working with Globus Auth identities.

This command group provides subcommands for looking up and
managing Globus Auth identities.`,
	}

	// Add subcommands
	identitiesCmd.AddCommand(
		identitiesLookupCmd(),
	)

	return identitiesCmd
}

// Identity represents a Globus Auth identity
type Identity struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Status     string `json:"status"`
	IDProvider string `json:"identity_provider"`
}

// identitiesLookupCmd returns the identities lookup command
func identitiesLookupCmd() *cobra.Command {
	var username string
	var email string
	var id string

	cmd := &cobra.Command{
		Use:   "lookup",
		Short: "Look up Globus Auth identities",
		Long: `Look up Globus Auth identities by username, email, or ID.

This command queries Globus Auth to find identities matching the
provided criteria.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate that at least one search parameter is provided
			if username == "" && email == "" && id == "" && len(args) == 0 {
				return fmt.Errorf("must provide at least one of: --username, --email, --id, or a search term as an argument")
			}

			// If an argument is provided, use it for search
			if len(args) > 0 && username == "" && email == "" && id == "" {
				// Determine if the argument looks like an email
				if strings.Contains(args[0], "@") {
					email = args[0]
				} else if strings.HasPrefix(args[0], "urn:globus:auth:identity:") {
					id = args[0]
				} else {
					username = args[0]
				}
			}

			// Get the current profile
			profile := viper.GetString("profile")
			
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

			// Set up bearer token
			authClient.SetAccessToken(tokenInfo.AccessToken)

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Look up identities
			var identities []Identity
			
			if id != "" {
				// Look up by ID
				resp, err := authClient.GetIdentities(ctx, id)
				if err != nil {
					return fmt.Errorf("failed to look up identity: %w", err)
				}
				
				for _, identity := range resp.Identities {
					identities = append(identities, Identity{
						ID:         identity.ID,
						Username:   identity.Username,
						Name:       identity.Name,
						Email:      identity.Email,
						Status:     identity.Status,
						IDProvider: identity.IdentityProvider,
					})
				}
			} else {
				// Look up by username or email
				params := make(map[string]string)
				if username != "" {
					params["username"] = username
				}
				if email != "" {
					params["email"] = email
				}
				
				resp, err := authClient.LookupIdentities(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to look up identities: %w", err)
				}
				
				for _, identity := range resp.Identities {
					identities = append(identities, Identity{
						ID:         identity.ID,
						Username:   identity.Username,
						Name:       identity.Name,
						Email:      identity.Email,
						Status:     identity.Status,
						IDProvider: identity.IdentityProvider,
					})
				}
			}

			// Check if we found any identities
			if len(identities) == 0 {
				fmt.Println("No identities found")
				return nil
			}

			// Format and display the results
			format := viper.GetString("format")
			formatter := output.NewFormatter(format, cmd.OutOrStdout())
			
			headers := []string{"ID", "Username", "Name", "Email", "Status", "IDProvider"}
			if err := formatter.FormatOutput(identities, headers); err != nil {
				return fmt.Errorf("error formatting output: %w", err)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&username, "username", "", "Look up by username")
	cmd.Flags().StringVar(&email, "email", "", "Look up by email")
	cmd.Flags().StringVar(&id, "id", "", "Look up by identity ID")

	return cmd
}