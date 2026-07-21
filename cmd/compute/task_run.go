// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	taskRunInput string
	taskRunFile  string
)

// TaskRunCmd represents the compute task run command
var TaskRunCmd = &cobra.Command{
	Use:   "run FUNCTION_ID ENDPOINT_ID",
	Short: "Execute a function on an endpoint",
	Long: `Execute a registered function on a specified compute endpoint.

NOTE: The Globus Compute submit API requires arguments serialized with the
Globus Compute (dill/base64) format, which this CLI cannot produce. Use the
Globus Compute Python SDK/Executor to run functions. The web-service submit
endpoint is available in the Go SDK as compute.Client.Submit for callers that
build the serialized payload themselves.`,
	Args: cobra.ExactArgs(2),
	RunE: runTaskRun,
}

func init() {
	TaskRunCmd.Flags().StringVar(&taskRunInput, "input", "", "Function input as JSON string")
	TaskRunCmd.Flags().StringVar(&taskRunFile, "file", "", "Path to JSON file containing function input")
}

func runTaskRun(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("running functions from the CLI is not supported: the Compute submit API requires Globus Compute-serialized arguments; use the Globus Compute Python SDK")
}
