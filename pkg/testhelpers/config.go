// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package testhelpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigData represents the configuration data for the CLI
type ConfigData struct {
	DefaultProfile string                 `yaml:"default_profile"`
	Profiles       map[string]ProfileData `yaml:"profiles"`
}

// ProfileData represents a configuration profile
type ProfileData struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	Description string `yaml:"description"`
}

// SetupTestConfig creates a temporary config file structure for testing
// This creates a complete test environment with config and profiles
func SetupTestConfig(t *testing.T) (string, func()) {
	// Create a temporary home directory for testing
	homeDir, err := os.MkdirTemp("", "globus-cli-test-home-")
	if err != nil {
		t.Fatalf("Failed to create temp home directory: %v", err)
	}

	// Create the .globus-cli structure
	configDir := filepath.Join(homeDir, ".globus-cli")
	tokensDir := filepath.Join(configDir, "tokens")
	profilesDir := filepath.Join(configDir, "profiles")

	// Create directories
	for _, dir := range []string{configDir, tokensDir, profilesDir} {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			os.RemoveAll(homeDir)
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create config.yaml
	config := ConfigData{
		DefaultProfile: "default",
		Profiles: map[string]ProfileData{
			"default": {
				Name:        "default",
				Environment: "production",
				Description: "Default profile for testing",
			},
			"test": {
				Name:        "test",
				Environment: "test",
				Description: "Test profile for testing",
			},
		},
	}

	// Marshal config to YAML
	configYAML, err := yaml.Marshal(config)
	if err != nil {
		os.RemoveAll(homeDir)
		t.Fatalf("Failed to marshal config YAML: %v", err)
	}

	// Write config file
	configFilePath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configFilePath, configYAML, 0600)
	if err != nil {
		os.RemoveAll(homeDir)
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create token files for both profiles
	createTokenFile(t, tokensDir, "default", homeDir)
	createTokenFile(t, tokensDir, "test", homeDir)

	// Save original HOME environment variable
	originalHome := os.Getenv("HOME")

	// Set HOME to our temp directory for testing
	os.Setenv("HOME", homeDir)

	// Return the config file path and a cleanup function
	return configFilePath, func() {
		// Restore original HOME
		os.Setenv("HOME", originalHome)
		// Remove the temp directory
		os.RemoveAll(homeDir)
	}
}

// Helper function to create a token file for a profile
func createTokenFile(t *testing.T, tokensDir, profile string, homeDir string) {
	token := TokenInfo{
		AccessToken:  "mock-access-token-for-" + profile,
		RefreshToken: "mock-refresh-token-for-" + profile,
		ExpiresIn:    3600,
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "openid profile email urn:globus:auth:scope:transfer.api.globus.org:all",
		TokenType:    "Bearer",
		Profile:      profile,
	}

	// Marshal the token to JSON
	tokenJSON, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		os.RemoveAll(homeDir)
		t.Fatalf("Failed to marshal token JSON: %v", err)
	}

	// Write the token file
	tokenFilePath := filepath.Join(tokensDir, profile+".json")
	err = os.WriteFile(tokenFilePath, tokenJSON, 0600)
	if err != nil {
		os.RemoveAll(homeDir)
		t.Fatalf("Failed to write token file: %v", err)
	}
}
