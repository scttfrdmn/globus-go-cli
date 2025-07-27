// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getConfigCommand returns the config command
func getConfigCommand() *cobra.Command {
	// configCmd represents the config command
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Commands for configuration management",
		Long: `Commands for managing Globus CLI configuration including
profiles and settings.`,
	}

	// Add config subcommands
	configCmd.AddCommand(
		configShowCmd(),
		configInitCmd(),
	)

	return configCmd
}

// configShowCmd returns the config show command
func configShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show configuration",
		Long: `Show the current Globus CLI configuration.

This command displays the current configuration settings for the CLI.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Current Configuration:")
			fmt.Printf("  Profile: %s\n", viper.GetString("profile"))
			fmt.Printf("  Config File: %s\n", viper.ConfigFileUsed())

			// Print all configuration values
			allSettings := viper.AllSettings()
			for key, value := range allSettings {
				// Skip sensitive values
				if key == "client" {
					clientMap, ok := value.(map[string]interface{})
					if ok {
						if id, exists := clientMap["id"]; exists {
							fmt.Printf("  Client ID: %v\n", id)
						}
						fmt.Println("  Client Secret: [hidden]")
					}
					continue
				}

				fmt.Printf("  %s: %v\n", key, value)
			}
		},
	}
}

// configInitCmd returns the config init command
func configInitCmd() *cobra.Command {
	var clientID, clientSecret string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		Long: `Initialize the Globus CLI configuration.

This command initializes the configuration for the CLI, creating
the necessary directories and files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set client ID and secret in viper
			if clientID != "" {
				viper.Set("client.id", clientID)
			}
			if clientSecret != "" {
				viper.Set("client.secret", clientSecret)
			}

			// Set default values if not already set
			if viper.GetString("client.id") == "" {
				viper.Set("client.id", "e6c75d97-532a-4c88-b031-f5a3014430e3")
			}

			// Write the configuration to a file
			configFile := viper.ConfigFileUsed()
			if configFile == "" {
				// Get the home directory
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("failed to get home directory: %w", err)
				}

				// Create the config directory
				configDir := fmt.Sprintf("%s/.globus-cli", homeDir)
				if err := os.MkdirAll(configDir, 0700); err != nil {
					return fmt.Errorf("failed to create config directory: %w", err)
				}

				// Set the config file path
				configFile = fmt.Sprintf("%s/config.yaml", configDir)
			}

			// Write the config file
			if err := viper.WriteConfigAs(configFile); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			fmt.Printf("Configuration initialized at %s\n", configFile)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&clientID, "client-id", "", "Client ID for Globus Auth")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "Client Secret for Globus Auth")

	return cmd
}
