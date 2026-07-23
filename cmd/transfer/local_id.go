// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// endpointLocalIDCmd returns the `endpoint local-id` command, matching the
// Python globus CLI. It reads the local Globus Connect Personal installation's
// endpoint ID from the config directory (~/.globusonline), with no network
// call. See globus_sdk.LocalGlobusConnectPersonal.
func endpointLocalIDCmd() *cobra.Command {
	// --personal is the default (and only) mode, accepted for Python parity.
	var personal bool
	cmd := &cobra.Command{
		Use:   "local-id",
		Short: "Display the local Globus Connect Personal endpoint ID",
		Long: `Display the endpoint ID of the Globus Connect Personal installation for the
current user, by reading the local GCP configuration directory
(~/.globusonline). No network request is made.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := localGCPEndpointID()
			if err != nil {
				return err
			}
			if id == "" {
				return fmt.Errorf("no Globus Connect Personal installation found for the current user")
			}
			fmt.Fprintln(cmd.OutOrStdout(), id)
			return nil
		},
	}
	cmd.Flags().BoolVar(&personal, "personal", true, "Use the local Globus Connect Personal endpoint (default)")
	return cmd
}

// localGCPEndpointID reads the GCP endpoint ID from the local config directory,
// mirroring globus_sdk.LocalGlobusConnectPersonal: <config>/lta/client-id.txt
// on non-Windows, <config>/client-id.txt on Windows, where <config> defaults to
// ~/.globusonline. A missing file yields ("", nil).
func localGCPEndpointID() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	configDir := filepath.Join(home, ".globusonline")
	var idFile string
	if runtime.GOOS == "windows" {
		idFile = filepath.Join(configDir, "client-id.txt")
	} else {
		idFile = filepath.Join(configDir, "lta", "client-id.txt")
	}
	data, err := os.ReadFile(idFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("cannot read local endpoint ID: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}
