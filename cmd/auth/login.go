// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
)

// TokenInfo holds the token data
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scopes       []string  `json:"scopes,omitempty"`
}

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
	loginCmd.Flags().StringSliceVar(&loginScopes, "scopes", []string{}, "comma-separated list of scopes to request (default: all)")
	loginCmd.Flags().BoolVar(&noLocalServer, "no-local-server", false, "do not start a local server for the OAuth callback")
	loginCmd.Flags().BoolVar(&noSaveTokens, "no-save-tokens", false, "do not save tokens to disk")
	loginCmd.Flags().BoolVar(&noOpenBrowser, "no-browser", false, "do not automatically open the browser")
	loginCmd.Flags().BoolVar(&forceLogin, "force", false, "force login even if valid tokens exist")

	return loginCmd
}

// login handles the login command. It uses the v4 SDK's GlobusApp
// (globusauth.NewApp), which runs the OAuth2 authorization-code flow and stores
// one token per resource server. For backward compatibility with commands not
// yet migrated to per-resource-server auth, it also writes the legacy combined
// token file (see writeLegacyBridgeToken).
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

	userApp, err := globusauth.NewApp(profile, clientCfg.ClientID, clientCfg.ClientSecret, globusauth.AllServices...)
	if err != nil {
		return fmt.Errorf("failed to initialize login: %w", err)
	}
	defer userApp.Close()

	fmt.Println()
	fmt.Println("You will be prompted to open a URL in your browser, authenticate,")
	fmt.Println("and paste back the resulting authorization code.")
	fmt.Println()

	// The v4 CommandLineLoginFlowManager prints the auth URL and reads the code
	// from stdin. Offer to open the browser first unless suppressed.
	if !noOpenBrowser {
		// The URL is printed by RunLoginFlow; we cannot pre-open it here without
		// duplicating URL construction, so we simply note the behavior. A future
		// enhancement can add a loopback local-server flow.
		_ = noLocalServer
	}

	if err := userApp.Login(context.Background()); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Compatibility bridge: mirror the transfer token into the legacy token file
	// so commands still using authcmd.LoadToken keep working during migration.
	if !noSaveTokens {
		if err := writeLegacyBridgeToken(profile, clientCfg.ClientID, clientCfg.ClientSecret); err != nil {
			fmt.Printf("Warning: could not write legacy token bridge: %v\n", err)
		}
	}

	fmt.Println()
	fmt.Println("Login successful! You are now authenticated with Globus.")
	return nil
}

// writeLegacyBridgeToken writes the legacy single-token file (used by
// not-yet-migrated commands) populated from the per-resource-server store. It
// uses the transfer token, which most legacy commands need; auth-only commands
// still work because they re-derive from the store via globusauth.
func writeLegacyBridgeToken(profile, clientID, clientSecret string) error {
	authz, err := globusauth.Authorizer(context.Background(), profile, clientID, clientSecret, globusauth.ServiceTransfer)
	if err != nil {
		return err
	}
	header, err := authz.GetAuthorizationHeader(context.Background())
	if err != nil {
		return err
	}
	// Header is "Bearer <token>".
	token := header
	if len(header) > 7 && header[:7] == "Bearer " {
		token = header[7:]
	}
	return saveToken(profile, &TokenInfo{
		AccessToken: token,
		ExpiresAt:   time.Now().Add(24 * time.Hour), // refreshed automatically by the authorizer
	})
}

// saveToken saves a token to disk
func saveToken(profile string, token *TokenInfo) error {
	// Get the token file path
	tokenFile, err := getTokenFilePath(profile)
	if err != nil {
		return err
	}

	// Marshal the token to JSON
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling token: %w", err)
	}

	// Ensure the tokens directory exists.
	if err := os.MkdirAll(filepath.Dir(tokenFile), 0700); err != nil {
		return fmt.Errorf("error creating tokens directory: %w", err)
	}

	// Write the token to the file
	if err := os.WriteFile(tokenFile, data, 0600); err != nil {
		return fmt.Errorf("error writing token file: %w", err)
	}

	return nil
}

// loadToken loads a token from disk
func loadToken(profile string) (*TokenInfo, error) {
	// Get the token file path
	tokenFile, err := getTokenFilePath(profile)
	if err != nil {
		return nil, err
	}

	// Check if the token file exists
	if _, err := os.Stat(tokenFile); err != nil {
		return nil, fmt.Errorf("token file does not exist: %w", err)
	}

	// Read the token file
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("error reading token file: %w", err)
	}

	// Unmarshal the token
	var token TokenInfo
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("error unmarshaling token: %w", err)
	}

	return &token, nil
}

// getTokenFilePath returns the path to the token file for a profile
func getTokenFilePath(profile string) (string, error) {
	// Get the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %w", err)
	}

	// Create the token file path
	tokensDir := filepath.Join(homeDir, ".globus-cli", "tokens")
	tokenFile := filepath.Join(tokensDir, profile+".json")

	return tokenFile, nil
}

// isTokenValid checks if a token is valid
func isTokenValid(token *TokenInfo) bool {
	// Check if the token is expired
	// Add a 5-minute buffer to avoid edge cases
	return token != nil && time.Now().Add(5*time.Minute).Before(token.ExpiresAt)
}

// printTokenInfo prints information about a token
func printTokenInfo(token *TokenInfo) {
	expiresIn := time.Until(token.ExpiresAt).Round(time.Second)

	fmt.Println("\nToken Information:")
	fmt.Printf("  Access Token: %s...%s\n",
		token.AccessToken[:10],
		token.AccessToken[len(token.AccessToken)-10:])

	if token.RefreshToken != "" {
		fmt.Printf("  Refresh Token: %s...%s\n",
			token.RefreshToken[:10],
			token.RefreshToken[len(token.RefreshToken)-10:])
	}

	fmt.Printf("  Expires At: %s\n", token.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("  Expires In: %s\n", expiresIn)

	if len(token.Scopes) > 0 {
		fmt.Println("  Scopes:")
		for _, scope := range token.Scopes {
			fmt.Printf("    - %s\n", scope)
		}
	}
}

