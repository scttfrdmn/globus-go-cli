// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
)

var (
	loginScopes   []string
	noLocalServer bool
	noSaveTokens  bool
	noOpenBrowser bool
	forceLogin    bool
)

// LoginCmd returns the login command
func LoginCmd() *cobra.Command {
	// loginCmd represents the login command
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Globus",
		Long: `Log in to Globus to get credentials for the CLI.

This command runs the OAuth2 authorization-code flow with Globus Auth. It
prints an authorization URL to open in your browser; after you consent, paste
the resulting authorization code back into the CLI. Tokens are stored per
resource server so each service is authorized independently.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return login(cmd)
		},
	}

	// Add login flags
	loginCmd.Flags().StringSliceVar(&loginScopes, "scopes", []string{}, "comma-separated services to request tokens for: auth,transfer,groups,search,flows,compute,timers (default: all except timers)")
	loginCmd.Flags().BoolVar(&noLocalServer, "no-local-server", false, "do not start a local server for the OAuth callback")
	loginCmd.Flags().BoolVar(&noSaveTokens, "no-save-tokens", false, "do not save tokens to disk")
	loginCmd.Flags().BoolVar(&noOpenBrowser, "no-browser", false, "do not automatically open the browser")
	loginCmd.Flags().BoolVar(&forceLogin, "force", false, "force login even if valid tokens exist")

	return loginCmd
}

// login handles the login command using the v4 SDK's GlobusApp
// (globusauth.NewApp), which runs the OAuth2 authorization-code flow and stores
// one token per resource server.
func login(cmd *cobra.Command) error {
	profile := viper.GetString("profile")
	fmt.Printf("Using profile: %s\n", profile)

	// Load client configuration (client ID/secret; native client by default).
	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Already-logged-in short-circuit: if the transfer resource server has a
	// valid stored token and --force was not given, do nothing.
	if !forceLogin {
		if _, aerr := globusauth.Authorizer(context.Background(), profile, clientCfg.ClientID, clientCfg.ClientSecret, globusauth.ServiceTransfer); aerr == nil {
			fmt.Println("You are already logged in. Use --force to log in again.")
			return nil
		}
	}

	// Resolve which services to request. Default (no --scopes) is the safe set
	// that every scope of works with the default native client; Timers is
	// opt-in because its client-specific scope isn't requestable by a generic
	// client (issue #40).
	services := globusauth.DefaultLoginServices
	if len(loginScopes) > 0 {
		services = nil
		var unknown []string
		for _, name := range loginScopes {
			if svc, ok := globusauth.ServiceByName(name); ok {
				services = append(services, svc)
			} else {
				unknown = append(unknown, name)
			}
		}
		if len(unknown) > 0 {
			return fmt.Errorf("unknown service(s) in --scopes: %v (valid: auth,transfer,groups,search,flows,compute,timers)", unknown)
		}
	}

	userApp, err := globusauth.NewApp(profile, clientCfg.ClientID, clientCfg.ClientSecret, services...)
	if err != nil {
		return fmt.Errorf("failed to initialize login: %w", err)
	}
	defer userApp.Close()

	fmt.Println()
	fmt.Println("You will be prompted to open a URL in your browser, authenticate,")
	fmt.Println("and paste back the resulting authorization code.")
	fmt.Println()

	// The v4 CommandLineLoginFlowManager prints the auth URL and reads the code
	// from stdin. A future enhancement can add a loopback local-server flow;
	// --no-local-server / --no-browser are accepted now for forward
	// compatibility with that behavior.
	_ = noLocalServer
	_ = noOpenBrowser

	if err := userApp.Login(context.Background()); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Println()
	fmt.Println("Login successful! You are now authenticated with Globus.")
	return nil
}
