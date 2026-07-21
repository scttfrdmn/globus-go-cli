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

	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

// SessionCmd returns the `session` command group, matching the Python Globus
// CLI's `globus session`. Currently only `show` is available: `update` and
// `consent` require a reauthentication / session-boundary flow the SDK does not
// yet expose.
func SessionCmd() *cobra.Command {
	sessionCmd := &cobra.Command{
		Use:   "session",
		Short: "Commands for managing your Globus Auth session",
		Long: `Commands for inspecting your current Globus Auth session.

A session records which of your identities have authenticated and when — the
basis for high-assurance ("session boundary") access decisions.`,
	}

	sessionCmd.AddCommand(sessionShowCmd())
	return sessionCmd
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
