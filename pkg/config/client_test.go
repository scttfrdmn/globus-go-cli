// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers"
)

func TestLoadClientConfig(t *testing.T) {
	// Create a temporary config directory
	tempDir := testhelpers.CreateTempConfigDir(t)
	defer testhelpers.CleanupTempConfigDir(t, tempDir)

	// Save original config dir environment variable if it exists
	origConfigDir := os.Getenv("GLOBUS_CLI_CONFIG_DIR")
	defer os.Setenv("GLOBUS_CLI_CONFIG_DIR", origConfigDir)

	// Set the config dir to our temp directory
	os.Setenv("GLOBUS_CLI_CONFIG_DIR", tempDir)

	// Create a test config file
	configContent := `client_id: test-client-id
client_secret: test-client-secret
`
	testhelpers.CreateTestConfigFile(t, tempDir, configContent)

	// Load the config
	clientCfg, err := LoadClientConfig()
	if err != nil {
		t.Fatalf("Failed to load client config: %v", err)
	}

	// Verify the config
	if clientCfg.ClientID != "test-client-id" {
		t.Errorf("Expected client ID 'test-client-id', got '%s'", clientCfg.ClientID)
	}

	if clientCfg.ClientSecret != "test-client-secret" {
		t.Errorf("Expected client secret 'test-client-secret', got '%s'", clientCfg.ClientSecret)
	}
}

func TestLoadClientConfigDefault(t *testing.T) {
	// Create a temporary config directory
	tempDir := testhelpers.CreateTempConfigDir(t)
	defer testhelpers.CleanupTempConfigDir(t, tempDir)

	// Save original config dir environment variable if it exists
	origConfigDir := os.Getenv("GLOBUS_CLI_CONFIG_DIR")
	defer os.Setenv("GLOBUS_CLI_CONFIG_DIR", origConfigDir)

	// Set the config dir to our temp directory
	os.Setenv("GLOBUS_CLI_CONFIG_DIR", tempDir)

	// Don't create a config file - should use default values

	// Load the config
	clientCfg, err := LoadClientConfig()
	if err != nil {
		t.Fatalf("Failed to load client config: %v", err)
	}

	// Verify the config has default values
	if clientCfg.ClientID == "" {
		t.Errorf("Expected default client ID, got empty string")
	}

	if clientCfg.ClientSecret == "" {
		t.Errorf("Expected default client secret, got empty string")
	}
}
