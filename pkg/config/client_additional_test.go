// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package config

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadClientConfigWithEnvAndViper(t *testing.T) {
	// Save original environment variables and viper settings
	originalClientID := os.Getenv("GLOBUS_CLIENT_ID")
	originalClientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")
	defer func() {
		os.Setenv("GLOBUS_CLIENT_ID", originalClientID)
		os.Setenv("GLOBUS_CLIENT_SECRET", originalClientSecret)
		viper.Set("client.id", "")
		viper.Set("client.secret", "")
	}()

	// Test case: environment variables take precedence over viper settings
	tests := []struct {
		name               string
		envClientID        string
		envClientSecret    string
		viperClientID      string
		viperClientSecret  string
		expectClientID     string
		expectClientSecret string
	}{
		{
			name:               "env vars take precedence",
			envClientID:        "env-client-id",
			envClientSecret:    "env-client-secret",
			viperClientID:      "viper-client-id",
			viperClientSecret:  "viper-client-secret",
			expectClientID:     "env-client-id",
			expectClientSecret: "env-client-secret",
		},
		{
			name:               "env vars only client ID",
			envClientID:        "env-client-id",
			envClientSecret:    "",
			viperClientID:      "viper-client-id",
			viperClientSecret:  "viper-client-secret",
			expectClientID:     "env-client-id",
			expectClientSecret: "viper-client-secret",
		},
		{
			name:               "env vars only client secret",
			envClientID:        "",
			envClientSecret:    "env-client-secret",
			viperClientID:      "viper-client-id",
			viperClientSecret:  "viper-client-secret",
			expectClientID:     "viper-client-id",
			expectClientSecret: "env-client-secret",
		},
		{
			name:               "default client ID when not set",
			envClientID:        "",
			envClientSecret:    "env-client-secret",
			viperClientID:      "",
			viperClientSecret:  "",
			expectClientID:     DefaultClientID,
			expectClientSecret: "env-client-secret",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("GLOBUS_CLIENT_ID", tc.envClientID)
			os.Setenv("GLOBUS_CLIENT_SECRET", tc.envClientSecret)

			// Set viper values
			viper.Set("client.id", tc.viperClientID)
			viper.Set("client.secret", tc.viperClientSecret)

			// Load the client config
			config, err := LoadClientConfig()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify the config
			if config.ClientID != tc.expectClientID {
				t.Errorf("Expected client ID %s, got %s", tc.expectClientID, config.ClientID)
			}
			if config.ClientSecret != tc.expectClientSecret {
				t.Errorf("Expected client secret %s, got %s", tc.expectClientSecret, config.ClientSecret)
			}
		})
	}
}

func TestGetConfigFilePath(t *testing.T) {
	// Call the function
	configPath, err := getConfigFilePath()

	// Check for errors
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the path ends with .globus-cli/config.yaml
	if !strings.HasSuffix(configPath, ".globus-cli/config.yaml") {
		t.Errorf("Expected path to end with .globus-cli/config.yaml, got %s", configPath)
	}

	// Verify the path is based on user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Could not get home directory: %v", err)
	}

	expectedPrefix := homeDir
	if !strings.HasPrefix(configPath, expectedPrefix) {
		t.Errorf("Expected path to start with %s, got %s", expectedPrefix, configPath)
	}
}
