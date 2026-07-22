// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

// SessionCmd returns the `session` command group, matching the Python Globus
// CLI's `globus session` (show / update / consent).
func SessionCmd() *cobra.Command {
	sessionCmd := &cobra.Command{
		Use:   "session",
		Short: "Commands for managing your Globus Auth session",
		Long: `Commands for inspecting and updating your Globus Auth session.

A session records which of your identities have authenticated and when — the
basis for high-assurance ("session boundary") access decisions.`,
	}

	sessionCmd.AddCommand(sessionShowCmd(), sessionUpdateCmd(), sessionConsentCmd())
	return sessionCmd
}

// loadClientCreds returns the configured client ID/secret for the current
// profile (native client by default).
func loadClientCreds() (clientID, clientSecret string, err error) {
	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return "", "", fmt.Errorf("failed to load client configuration: %w", err)
	}
	return clientCfg.ClientID, clientCfg.ClientSecret, nil
}

// sessionUpdateCmd returns the `session update` command: it re-runs the login
// flow with session-enforcement parameters so Globus Auth forces step-up
// re-authentication of specific identities/domains or per an auth policy.
func sessionUpdateCmd() *cobra.Command {
	var (
		identities []string
		domain     string
		policies   []string
		mfa        bool
	)
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update your session by re-authenticating",
		Long: `Update your current CLI auth session by re-authenticating. You can require
specific identities, a single identity-provider domain, authentication
policies, or MFA. This starts a browser/paste-code login flow like 'login'.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if domain != "" && len(identities) > 0 {
				return fmt.Errorf("--domain is mutually exclusive with --identity")
			}
			profile := viper.GetString("profile")
			clientID, clientSecret, err := loadClientCreds()
			if err != nil {
				return err
			}

			sp := &globusauth.SessionParams{
				RequiredIdentities: identities,
				RequiredPolicies:   policies,
				RequiredMFA:        mfa,
			}
			if domain != "" {
				sp.RequiredSingleDomain = []string{domain}
			}

			// Re-auth for the auth resource server (openid identity scopes).
			scope, _ := globusauth.Scope(globusauth.ServiceAuth)
			if _, err := globusauth.SessionLogin(cmd.Context(), profile, clientID, clientSecret, []string{scope}, sp); err != nil {
				return fmt.Errorf("session update failed: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Session updated.")
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&identities, "identity", nil, "Require re-authentication of these identity IDs/usernames")
	cmd.Flags().StringVar(&domain, "domain", "", "Require an identity from this single domain (mutually exclusive with --identity)")
	cmd.Flags().StringSliceVar(&policies, "policy", nil, "Comma-separated authentication policy UUIDs to satisfy")
	cmd.Flags().BoolVar(&mfa, "mfa", false, "Require multi-factor authentication")
	return cmd
}

// sessionConsentCmd returns the `session consent` command: it runs the login
// flow requesting one or more explicit scopes, granting those consents.
func sessionConsentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "consent SCOPE [SCOPE...]",
		Short: "Add specific consents to your session",
		Long: `Update your current CLI auth session by authenticating with a specific scope
or set of scopes. Use this when a command needs a consent you have not yet
granted (e.g. a collection's data_access scope).`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := viper.GetString("profile")
			clientID, clientSecret, err := loadClientCreds()
			if err != nil {
				return err
			}
			stored, err := globusauth.SessionLogin(cmd.Context(), profile, clientID, clientSecret, args, nil)
			if err != nil {
				return fmt.Errorf("session consent failed: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Consent granted; stored tokens for %d resource server(s).\n", len(stored))
			return nil
		},
	}
}

// sessionShowCmd returns the `session show` command.
func sessionShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show your current Globus Auth session",
		Long: `Show the identities that have authenticated in your current Globus Auth
session and when each authenticated.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := viper.GetString("profile")

			// Read the stored auth token to introspect.
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

			intro, err := authClient.IntrospectToken(ctx, td.AccessToken, &auth.IntrospectOptions{Include: "session_info"})
			if err != nil {
				return fmt.Errorf("failed to introspect session: %w", err)
			}

			// For JSON/unix or a --jq expression, emit the raw session_info.
			format := viper.GetString("format")
			formatter := output.NewFormatter(format, cmd.OutOrStdout())
			if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
				return formatter.FormatOutput(intro.SessionInfo, nil)
			}

			if intro.SessionInfo == nil || len(intro.SessionInfo.Authentications) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No authenticated identities in the current session.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Session ID: %s\n\n", intro.SessionInfo.SessionID)

			// Stable output: sort by identity ID.
			ids := make([]string, 0, len(intro.SessionInfo.Authentications))
			for id := range intro.SessionInfo.Authentications {
				ids = append(ids, id)
			}
			sort.Strings(ids)

			type row struct {
				IdentityID     string
				AuthenticatedAt string
				MFA            bool
			}
			rows := make([]row, 0, len(ids))
			for _, id := range ids {
				a := intro.SessionInfo.Authentications[id]
				at := ""
				if a.AuthTime > 0 {
					at = time.Unix(a.AuthTime, 0).Format(time.RFC3339)
				}
				mfa := false
				for _, m := range a.AMR {
					if m == "mfa" {
						mfa = true
						break
					}
				}
				rows = append(rows, row{IdentityID: id, AuthenticatedAt: at, MFA: mfa})
			}
			return formatter.FormatOutput(rows, []string{"IdentityID", "AuthenticatedAt", "MFA"})
		},
	}
}
