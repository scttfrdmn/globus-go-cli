// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadClientConfig(t *testing.T) {
	// Save original environment variables
	originalClientID := os.Getenv("GLOBUS_CLIENT_ID")
	originalClientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")
	defer func() {
		os.Setenv("GLOBUS_CLIENT_ID", originalClientID)
		os.Setenv("GLOBUS_CLIENT_SECRET", originalClientSecret)
	}()

	// Set environment variables for test
	os.Setenv("GLOBUS_CLIENT_ID", "test-client-id-from-env")
	os.Setenv("GLOBUS_CLIENT_SECRET", "test-client-secret-from-env")

	// Test loading from environment variables
	config, err := LoadClientConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if config.ClientID != "test-client-id-from-env" {
		t.Errorf("Expected client ID from environment, got: %s", config.ClientID)
	}
	if config.ClientSecret != "test-client-secret-from-env" {
		t.Errorf("Expected client secret from environment, got: %s", config.ClientSecret)
	}

	// Test loading from viper
	os.Unsetenv("GLOBUS_CLIENT_ID")
	os.Unsetenv("GLOBUS_CLIENT_SECRET")
	viper.Set("client.id", "test-client-id-from-viper")
	viper.Set("client.secret", "test-client-secret-from-viper")

	config, err = LoadClientConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if config.ClientID != "test-client-id-from-viper" {
		t.Errorf("Expected client ID from viper, got: %s", config.ClientID)
	}
	if config.ClientSecret != "test-client-secret-from-viper" {
		t.Errorf("Expected client secret from viper, got: %s", config.ClientSecret)
	}

	// Test default client ID when not set
	viper.Set("client.id", "")
	config, err = LoadClientConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if config.ClientID != DefaultClientID {
		t.Errorf("Expected default client ID, got: %s", config.ClientID)
	}
}

func TestSaveClientConfig(t *testing.T) {
	t.Skip("Skipping test that modifies viper global state")

	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "globus-cli-config-test")
	if err != nil {
		t.Fatalf("Error creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up viper to use our temporary directory
	originalConfigFile := viper.ConfigFileUsed()
	defer func() {
		// Attempt to restore viper's state after the test
		if originalConfigFile != "" {
			viper.SetConfigFile(originalConfigFile)
		}
	}()

	// Configure viper to use our temp directory
	configFile := filepath.Join(tmpDir, "config.yaml")
	viper.SetConfigFile(configFile)

	// Create test config
	testConfig := &ClientConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	// Save the config
	err = SaveClientConfig(testConfig)
	if err != nil {
		t.Fatalf("Error saving config: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatalf("Config file was not created")
	}

	// Reset viper and read the config file
	viper.Reset()
	viper.SetConfigFile(configFile)
	err = viper.ReadInConfig()
	if err != nil {
		t.Fatalf("Error reading saved config: %v", err)
	}

	// Verify the config values
	if viper.GetString("client.id") != testConfig.ClientID {
		t.Errorf("Expected client ID %s, got %s", testConfig.ClientID, viper.GetString("client.id"))
	}
	if viper.GetString("client.secret") != testConfig.ClientSecret {
		t.Errorf("Expected client secret %s, got %s", testConfig.ClientSecret, viper.GetString("client.secret"))
	}
}

func TestLoadClientConfigDefault(t *testing.T) {
	// Ensure environment variables are not set
	os.Unsetenv("GLOBUS_CLIENT_ID")
	os.Unsetenv("GLOBUS_CLIENT_SECRET")

	// Reset viper values
	viper.Reset()

	// Test loading with no values set (should use default client ID)
	config, err := LoadClientConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if config.ClientID != DefaultClientID {
		t.Errorf("Expected default client ID, got: %s", config.ClientID)
	}
	if config.ClientSecret != "" {
		t.Errorf("Expected empty client secret, got: %s", config.ClientSecret)
	}
}
