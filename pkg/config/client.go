// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ClientConfig holds the client configuration
type ClientConfig struct {
	ClientID     string `json:"client_id" yaml:"client_id"`
	ClientSecret string `json:"client_secret" yaml:"client_secret"`
}

// DefaultClientID is the default client ID for the CLI
const DefaultClientID = "e6c75d97-532a-4c88-b031-f5a3014430e3"

// LoadClientConfig loads the client configuration
func LoadClientConfig() (*ClientConfig, error) {
	// Check if client ID/secret are in environment variables
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	// If not in environment, check viper/config file
	if clientID == "" {
		clientID = viper.GetString("client.id")
	}
	if clientSecret == "" {
		clientSecret = viper.GetString("client.secret")
	}

	// If still not found, use default values
	if clientID == "" {
		clientID = DefaultClientID
	}
	// No default for client secret

	return &ClientConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}

// SaveClientConfig saves the client configuration
func SaveClientConfig(config *ClientConfig) error {
	// Set in viper
	viper.Set("client.id", config.ClientID)
	viper.Set("client.secret", config.ClientSecret)

	// Get config file path
	configFile, err := getConfigFilePath()
	if err != nil {
		return err
	}

	// Create the config directory if it doesn't exist
	configDir := filepath.Dir(configFile)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write the config file
	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getConfigFilePath returns the path to the config file
func getConfigFilePath() (string, error) {
	// Get the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create the config file path
	configDir := filepath.Join(homeDir, ".globus-cli")
	configFile := filepath.Join(configDir, "config.yaml")

	return configFile, nil
}
