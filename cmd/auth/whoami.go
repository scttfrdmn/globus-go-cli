// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	sdkauth "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

var whoamiLinkedIdentities bool

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
			if whoamiLinkedIdentities {
				return whoamiLinked(cmd)
			}
			return whoami(cmd)
		},
	}

	whoamiCmd.Flags().BoolVar(&whoamiLinkedIdentities, "linked-identities", false, "Also show identities linked to the currently logged-in primary identity")

	return whoamiCmd
}

// whoami handles the whoami command
func whoami(cmd *cobra.Command) error {
	// Get the current profile
	profile := viper.GetString("profile")
	fmt.Printf("Using profile: %s\n", profile)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	authClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// The OIDC userinfo endpoint returns the caller's identity directly.
	userInfo, err := authClient.GetUserInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user identity: %w", err)
	}

	fmt.Println("\nCurrent User:")
	fmt.Printf("  Username: %s\n", userInfo.PreferredUsername)
	fmt.Printf("  Identity ID: %s\n", userInfo.Sub)
	fmt.Printf("  Email: %s\n", userInfo.Email)
	fmt.Printf("  Name: %s\n", userInfo.Name)
	if userInfo.Organization != "" {
		fmt.Printf("  Organization: %s\n", userInfo.Organization)
	}

	// Show token expiry for the auth resource server from the stored tokens.
	if td, terr := globusauth.TokenFor(profile, globusauth.ServiceAuth); terr == nil {
		fmt.Println("\nToken Information:")
		fmt.Printf("  Expires At: %s\n", td.ExpiresAt.Format(time.RFC3339))
		fmt.Printf("  Expires In: %s\n", time.Until(td.ExpiresAt).Round(time.Second))
	}

	return nil
}

// whoamiLinked shows the identities linked to the caller's primary identity by
// introspecting the stored auth token with include=identity_set_detail.
func whoamiLinked(cmd *cobra.Command) error {
	profile := viper.GetString("profile")

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

	intro, err := authClient.IntrospectToken(ctx, td.AccessToken, &sdkauth.IntrospectOptions{Include: "identity_set_detail"})
	if err != nil {
		return fmt.Errorf("failed to introspect token: %w", err)
	}

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(intro.IdentitySetDetail, nil)
	}

	if len(intro.IdentitySetDetail) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No linked identities found.")
		return nil
	}

	headers := []string{"ID", "Username", "Name", "Email", "IdentityProvider"}
	return formatter.FormatOutput(intro.IdentitySetDetail, headers)
}
