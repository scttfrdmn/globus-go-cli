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
	cfgFile     string
	profileName string
	verbose     bool
	outputFormat string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "globus",
	Short: "Globus CLI - Command line interface for Globus services",
	Long: `Globus CLI - A command line interface for interacting with Globus services.
	
This CLI provides access to Globus services including Auth, Transfer, Search,
Groups, Flows, Compute, and Timers.`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
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