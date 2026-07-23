// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// credentialRecord tracks CLI-side rotation metadata for a single client
// credential. The Globus Auth API has no rotation/scheduling primitive, so
// `rotate`/`process-deletions` maintain this state locally (mirroring the
// globus-project-manager tool's credential_metadata).
type credentialRecord struct {
	ClientID          string `json:"client_id"`
	Name              string `json:"name,omitempty"`
	RotatedTo         string `json:"rotated_to,omitempty"`
	RotatedFrom       string `json:"rotated_from,omitempty"`
	ScheduledDeletion string `json:"scheduled_deletion,omitempty"` // RFC3339
}

// credentialState is the on-disk per-profile rotation state, keyed by
// credential ID.
type credentialState struct {
	Credentials map[string]credentialRecord `json:"credentials"`
}

// statePath returns the per-profile credential-state file path
// (~/.globus-cli/credential-state-<profile>.json), alongside the token store.
func statePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	profile := viper.GetString("profile")
	if profile == "" {
		profile = "default"
	}
	dir := filepath.Join(home, ".globus-cli")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create state directory: %w", err)
	}
	return filepath.Join(dir, "credential-state-"+profile+".json"), nil
}

// loadState reads the per-profile credential state, returning an empty (but
// non-nil) state if the file does not exist.
func loadState() (*credentialState, error) {
	path, err := statePath()
	if err != nil {
		return nil, err
	}
	st := &credentialState{Credentials: map[string]credentialRecord{}}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return st, nil
		}
		return nil, fmt.Errorf("cannot read credential state: %w", err)
	}
	if len(data) == 0 {
		return st, nil
	}
	if err := json.Unmarshal(data, st); err != nil {
		return nil, fmt.Errorf("invalid credential state file %s: %w", path, err)
	}
	if st.Credentials == nil {
		st.Credentials = map[string]credentialRecord{}
	}
	return st, nil
}

// saveState writes the per-profile credential state atomically (0600).
func saveState(st *credentialState) error {
	path, err := statePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal credential state: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("cannot write credential state: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("cannot save credential state: %w", err)
	}
	return nil
}
