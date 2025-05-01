// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package testhelpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

// TestCredentials stores credentials and test data for integration tests
type TestCredentials struct {
	// Auth credentials
	ClientID     string
	ClientSecret string
	Username     string
	Password     string

	// Resource IDs
	SourceEndpoint      string
	DestinationEndpoint string
	SourcePath          string
	DestinationPath     string

	// Identity
	TestIdentity string
}

// LoadTestCredentials loads test credentials from .env.test file
// Skip tests that require credentials if the file is not found
func LoadTestCredentials(t *testing.T) *TestCredentials {
	// Look for .env.test file in project root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate up to find project root (containing .env.test)
	var envFile string
	for dir := wd; dir != "/"; dir = filepath.Dir(dir) {
		testFile := filepath.Join(dir, ".env.test")
		if _, err := os.Stat(testFile); err == nil {
			envFile = testFile
			break
		}
	}

	if envFile == "" {
		t.Skip("Skipping test: .env.test file not found. Create one from .env.test.example to run integration tests.")
	}

	// Load .env.test file
	err = godotenv.Load(envFile)
	if err != nil {
		t.Skipf("Error loading .env.test file: %v", err)
	}

	// Check if required credentials are set
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		t.Skip("Skipping test: GLOBUS_TEST_CLIENT_ID or GLOBUS_TEST_CLIENT_SECRET not set in .env.test")
	}

	// Return credentials
	return &TestCredentials{
		ClientID:            clientID,
		ClientSecret:        clientSecret,
		Username:            os.Getenv("GLOBUS_TEST_USERNAME"),
		Password:            os.Getenv("GLOBUS_TEST_PASSWORD"),
		SourceEndpoint:      os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT"),
		DestinationEndpoint: os.Getenv("GLOBUS_TEST_DESTINATION_ENDPOINT"),
		SourcePath:          os.Getenv("GLOBUS_TEST_SOURCE_PATH"),
		DestinationPath:     os.Getenv("GLOBUS_TEST_DESTINATION_PATH"),
		TestIdentity:        os.Getenv("GLOBUS_TEST_IDENTITY"),
	}
}

// RequireTransferEndpoints ensures that transfer endpoints are configured
// Skip tests that require endpoints if they're not configured
func (c *TestCredentials) RequireTransferEndpoints(t *testing.T) {
	if c.SourceEndpoint == "" || c.DestinationEndpoint == "" {
		t.Skip("Skipping test: Transfer endpoints not configured in .env.test")
	}
}

// RequireFullCredentials ensures that all credentials needed for interactive tests are available
// Skip tests that require full credentials if they're not configured
func (c *TestCredentials) RequireFullCredentials(t *testing.T) {
	if c.Username == "" || c.Password == "" {
		t.Skip("Skipping test: Full credentials (username/password) not configured in .env.test")
	}
}