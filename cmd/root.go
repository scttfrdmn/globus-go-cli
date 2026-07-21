// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
)

// statusFromError extracts an HTTP status code from an SDK API error, or 0.
func statusFromError(err error) int {
	var apiErr *core.Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode
	}
	return 0
}

var (
	cfgFile       string
	profileName   string
	verbose       bool
	quiet         bool
	outputFormat  string
	jmesPath      string
	jqPath        string
	mapHTTPStatus string
)

// Version is set during the build process
var Version = "4.5.0-1"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "globus",
	Short: "Globus CLI - Command line interface for Globus services",
	Long: `Globus CLI - A command line interface for interacting with Globus services.

This CLI provides access to Globus services including Auth, Transfer, Search,
Groups, Flows, Compute, and Timers. It is designed to be a fast, modern
alternative to the Python-based Globus CLI.

Basic Usage:
  globus auth login                  Log in to Globus
  globus auth whoami                 Show current user information
  globus transfer endpoint list      List available Globus endpoints
  globus transfer ls ENDPOINT:PATH   List files on an endpoint
  globus transfer cp SOURCE DEST     Transfer files between endpoints
  globus transfer task show TASK_ID  Show status of a transfer task

Configuration:
  The CLI stores its configuration in ~/.globus-cli/ directory.
  You can use multiple profiles with the --profile flag.

Output Formats:
  Most commands support different output formats using the -F/--format flag:
  -F text                            Human-readable text (default)
  -F json                            JSON format for programmatic use
  -F unix                            Tab-delimited, no header (line-oriented tools)
  --jmespath / --jq EXPR             Filter JSON output with a JMESPath expression
  --map-http-status "404=50,..."     Map HTTP error statuses to process exit codes

For more information and examples, visit:
https://github.com/scttfrdmn/globus-go-cli`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() *cobra.Command {
	return rootCmd
}

// ExecuteCmd executes the root command.
func ExecuteCmd() error {
	return rootCmd.Execute()
}

// ExitCode runs the root command and returns the process exit code, honoring
// --map-http-status: if the command fails with an error carrying an HTTP status
// that the user mapped, that mapped code is returned. Otherwise a non-nil error
// yields 1 and success yields 0. The error (if any) is also returned so the
// caller can print it.
func ExitCode() (int, error) {
	err := rootCmd.Execute()
	if err == nil {
		return 0, nil
	}
	if statusMap, perr := output.ParseHTTPStatusMap(mapHTTPStatus); perr == nil {
		if code, ok := output.ExitCodeForError(wrapHTTPStatus(err), statusMap); ok {
			return code, err
		}
	}
	return 1, err
}

// httpStatusError adapts an SDK error (which exposes its status via a
// StatusCode field) to output.HTTPStatusError.
type httpStatusError struct {
	err    error
	status int
}

func (e httpStatusError) Error() string   { return e.err.Error() }
func (e httpStatusError) HTTPStatus() int { return e.status }

// wrapHTTPStatus wraps err so its HTTP status (if any) is discoverable by
// output.ExitCodeForError. Returns err unchanged if no status is found.
func wrapHTTPStatus(err error) error {
	if s := statusFromError(err); s != 0 {
		return httpStatusError{err: err, status: s}
	}
	return err
}

func init() {
	cobra.OnInitialize(initConfig)

	// Let every command's formatter honor the global --jmespath/--jq flag.
	output.JMESPathHook = EffectiveJMESPath

	// Global flags. These mirror the Python Globus CLI so scripts are portable:
	//   -F/--format [unix|json|text], --jmespath/--jq, --map-http-status, --quiet.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.globus-cli/config.yaml)")
	// No -p shorthand: it collides with per-command flags (e.g. `mkdir -p`), and
	// the Python CLI has no -p for profile either (it uses GLOBUS_PROFILE).
	rootCmd.PersistentFlags().StringVar(&profileName, "profile", "default", "CLI profile to use (also settable via GLOBUS_PROFILE)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "control level of output, make it more verbose")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "suppress non-essential output (higher precedence than --verbose)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "F", "text", "output format: unix, json, or text")
	rootCmd.PersistentFlags().StringVar(&jmesPath, "jmespath", "", "a JMESPath expression to apply to json output; forces json format")
	rootCmd.PersistentFlags().StringVar(&jqPath, "jq", "", "alias for --jmespath")
	rootCmd.PersistentFlags().StringVar(&mapHTTPStatus, "map-http-status", "", "map HTTP statuses to exit codes, e.g. \"404=50,403=51\"")

	// Bind flags to viper
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("jmespath", rootCmd.PersistentFlags().Lookup("jmespath"))
	viper.BindPFlag("map_http_status", rootCmd.PersistentFlags().Lookup("map-http-status"))

	// Add service commands
	addServiceCommands()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error determining home directory: %v\n", err)
			os.Exit(1)
		}

		// Create the config directory if it doesn't exist
		configDir := filepath.Join(home, ".globus-cli")
		if err := os.MkdirAll(configDir, 0700); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
			os.Exit(1)
		}

		// Search for config in .globus-cli directory
		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		// Create tokens directory if it doesn't exist
		tokensDir := filepath.Join(configDir, "tokens")
		if err := os.MkdirAll(tokensDir, 0700); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating tokens directory: %v\n", err)
			os.Exit(1)
		}

		// Create profiles directory if it doesn't exist
		profilesDir := filepath.Join(configDir, "profiles")
		if err := os.MkdirAll(profilesDir, 0700); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating profiles directory: %v\n", err)
			os.Exit(1)
		}
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
		}
	}

	// Environment variable support
	viper.SetEnvPrefix("GLOBUS")
	viper.AutomaticEnv() // read in environment variables that match

	// Honor GLOBUS_PROFILE (the Python CLI's profile-switching env var) when the
	// --profile flag was left at its default.
	if env := os.Getenv("GLOBUS_PROFILE"); env != "" && !rootCmd.PersistentFlags().Changed("profile") {
		profileName = env
		viper.Set("profile", env)
	}
}

// EffectiveJMESPath returns the JMESPath expression from either --jmespath or
// its --jq alias (--jmespath wins if both are set).
func EffectiveJMESPath() string {
	if jmesPath != "" {
		return jmesPath
	}
	return jqPath
}

// OutputFormat returns the resolved output format string.
func OutputFormat() string { return viper.GetString("format") }

// MapHTTPStatus returns the raw --map-http-status spec.
func MapHTTPStatus() string { return mapHTTPStatus }

// addServiceCommands adds all service commands to the root command
func addServiceCommands() {
	// Import and add all service commands
	rootCmd.AddCommand(
		getAuthCommand(),
		getTransferCommand(),
		getGroupCommand(),
		getTimerCommand(),
		getSearchCommand(),
		getFlowsCommand(),
		getComputeCommand(),
		// All services now implemented!
		getConfigCommand(),
	)
}
