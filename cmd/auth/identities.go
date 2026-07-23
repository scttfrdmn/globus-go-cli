// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	sdkauth "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
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

// GetIdentitiesCmd returns the flat top-level `get-identities` command, matching
// the Python CLI's `globus get-identities`. It is the same lookup used by
// `identities lookup`, exposed directly at the top level.
func GetIdentitiesCmd() *cobra.Command {
	cmd := identitiesLookupCmd()
	cmd.Use = "get-identities [VALUES...]"
	cmd.Short = "Look up Globus Auth identities"
	return cmd
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
	var provision bool

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

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			authClient, err := getClient(ctx)
			if err != nil {
				return err
			}

			// Build the lookup options. ID lookups query by identity ID;
			// username/email lookups query by username (email is a username in
			// Globus Auth).
			opts := &sdkauth.GetIdentitiesOptions{Provision: provision}
			switch {
			case id != "":
				opts.IDs = []string{id}
			case username != "":
				opts.Usernames = []string{username}
			case email != "":
				opts.Usernames = []string{email}
			}

			sdkIdentities, err := authClient.GetIdentities(ctx, opts)
			if err != nil {
				return fmt.Errorf("failed to look up identities: %w", err)
			}

			if len(sdkIdentities) == 0 {
				fmt.Println("No identities found")
				return nil
			}

			// Project to the CLI's display shape.
			identities := make([]Identity, 0, len(sdkIdentities))
			for _, si := range sdkIdentities {
				identities = append(identities, Identity{
					ID:         si.ID,
					Username:   si.Username,
					Name:       si.Name,
					Email:      si.Email,
					Status:     si.Status,
					IDProvider: si.IdentityProvider,
				})
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
	cmd.Flags().BoolVar(&provision, "provision", false, "Create identities if they do not exist (only affects username lookups)")

	return cmd
}
