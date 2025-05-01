// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
)

// TokenInfo holds the token data
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scopes       []string  `json:"scopes,omitempty"`
}

var (
	loginScopes  []string
	noLocalServer bool
	noSaveTokens bool
	noOpenBrowser bool
	forceLogin bool
)

// LoginCmd returns the login command
func LoginCmd() *cobra.Command {
	// loginCmd represents the login command
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Globus",
		Long: `Log in to Globus to get credentials for the CLI.

This command starts an OAuth2 authorization code flow with Globus Auth.
By default, it will open your browser to the Globus Auth consent page
and start a local web server to handle the redirect.

You can specify which scopes to request with --scopes.`,
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

// login handles the login command
func login(cmd *cobra.Command) error {
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

	// Create SDK config
	sdkConfig := pkg.NewConfig().
		WithClientID(clientCfg.ClientID).
		WithClientSecret(clientCfg.ClientSecret)

	// Create auth client
	authClient := sdkConfig.NewAuthClient()

	// Determine which scopes to request
	var scopes []string
	if len(loginScopes) > 0 {
		scopes = loginScopes
	} else {
		// Request default scopes for all services
		scopes = pkg.GetScopesByService("auth", "transfer", "groups", "search", "flows", "compute", "timers")
	}

	// Handle login based on method
	if noLocalServer {
		return loginWithoutLocalServer(authClient, profile, scopes)
	} else {
		return loginWithLocalServer(authClient, profile, scopes)
	}
}

// loginWithLocalServer performs login using a local server for the callback
func loginWithLocalServer(authClient *pkg.AuthClient, profile string, scopes []string) error {
	// Create channels for auth code or error
	authCode := make(chan string, 1)
	authErr := make(chan error, 1)

	// Start a local server to handle the callback
	server := &http.Server{Addr: ":8888"}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Get the authorization code from the query parameters
		code := r.URL.Query().Get("code")
		if code == "" {
			authErr <- fmt.Errorf("no authorization code in callback")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error: No authorization code received")
			return
		}

		// Send the code to the channel
		authCode <- code

		// Send a success response
		w.WriteHeader(http.StatusOK)
		// Simple HTML response
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Globus CLI - Login Successful</title>
			<style>
				body { font-family: Arial, sans-serif; max-width: 600px; margin: 40px auto; padding: 20px; }
				h1 { color: #2962FF; }
				.success { color: #00C853; font-weight: bold; }
				.info { color: #666; margin-top: 20px; }
			</style>
		</head>
		<body>
			<h1>Globus CLI</h1>
			<p class="success">✓ Login successful!</p>
			<p>Authentication complete. You can close this browser window and return to the CLI.</p>
			<p class="info">This window will automatically close in 5 seconds.</p>
			<script>
				setTimeout(function() { window.close(); }, 5000);
			</script>
		</body>
		</html>
		`
		fmt.Fprint(w, html)
	})

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			authErr <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Set the redirect URL
	authClient.SetRedirectURL("http://localhost:8888/callback")

	// Generate a random state parameter
	state := fmt.Sprintf("globus-cli-%d", time.Now().Unix())

	// Get the authorization URL
	authURL := authClient.GetAuthorizationURL(state, scopes...)

	// Print the URL for the user to open
	fmt.Println("Please open the following URL in your browser:")
	color.Cyan(authURL)

	// Automatically open the browser if allowed
	if !noOpenBrowser {
		fmt.Println("Attempting to open your browser...")
		openBrowser(authURL)
	}

	fmt.Println("\nWaiting for authentication...")

	// Wait for the authorization code or an error
	var code string
	select {
	case code = <-authCode:
		// Continue with the token exchange
	case err := <-authErr:
		return fmt.Errorf("authentication error: %w", err)
	case <-time.After(5 * time.Minute):
		return fmt.Errorf("authentication timed out after 5 minutes")
	}

	// Exchange the code for tokens
	fmt.Println("Exchanging authorization code for tokens...")
	tokenResp, err := authClient.ExchangeAuthorizationCode(context.Background(), code)
	if err != nil {
		return fmt.Errorf("error exchanging code for tokens: %w", err)
	}

	// Close the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Warning: Error shutting down server: %v\n", err)
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

// loginWithoutLocalServer performs login without using a local server
func loginWithoutLocalServer(authClient *pkg.AuthClient, profile string, scopes []string) error {
	// Use device code flow or manual copy-paste
	return fmt.Errorf("login without local server is not yet implemented")
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

// openBrowser tries to open the default browser with the given URL
func openBrowser(url string) {
	var err error

	switch os.Getenv("GOOS") {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}
}