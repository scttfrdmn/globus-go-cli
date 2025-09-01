// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	profileName  string
	verbose      bool
	outputFormat string
)

// Version is set during the build process
var Version = "3.36.0-1"

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
  Most commands support different output formats using the --format flag:
  --format=text                      Human-readable text (default)
  --format=json                      JSON format for programmatic use
  --format=csv                       CSV format for importing into spreadsheets

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

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.globus-cli/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&profileName, "profile", "p", "default", "profile to use")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "text", "output format (text, json, csv)")

	// Bind flags to viper
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))

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
}

// addServiceCommands adds all service commands to the root command
func addServiceCommands() {
	// Import and add all service commands
	rootCmd.AddCommand(
		getAuthCommand(),
		getTransferCommand(),
		// Other service commands will be added as they are implemented
		getConfigCommand(),
	)
}
