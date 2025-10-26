// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/compute"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

You can provide function arguments as JSON input either from a file or inline.
The input should be a JSON object with "args" (array) and/or "kwargs" (object).

Examples:
  # Run a function with no arguments
  globus compute task run FUNCTION_ID ENDPOINT_ID

  # Run with positional arguments
  globus compute task run FUNCTION_ID ENDPOINT_ID --input '{"args": [1, 2, 3]}'

  # Run with keyword arguments
  globus compute task run FUNCTION_ID ENDPOINT_ID --input '{"kwargs": {"x": 10, "y": 20}}'

  # Run with input from file
  globus compute task run FUNCTION_ID ENDPOINT_ID --file input.json`,
	Args: cobra.ExactArgs(2),
	RunE: runTaskRun,
}

func init() {
	TaskRunCmd.Flags().StringVar(&taskRunInput, "input", "", "Function input as JSON string")
	TaskRunCmd.Flags().StringVar(&taskRunFile, "file", "", "Path to JSON file containing function input")
}

func runTaskRun(cmd *cobra.Command, args []string) error {
	functionID := args[0]
	endpointID := args[1]

	// Parse input if provided
	var taskArgs []interface{}
	var taskKwargs map[string]interface{}

	if taskRunFile != "" || taskRunInput != "" {
		if taskRunFile != "" && taskRunInput != "" {
			return fmt.Errorf("cannot specify both --file and --input")
		}

		var inputJSON []byte
		var err error

		if taskRunFile != "" {
			inputJSON, err = os.ReadFile(taskRunFile)
			if err != nil {
				return fmt.Errorf("failed to read input file: %w", err)
			}
		} else {
			inputJSON = []byte(taskRunInput)
		}

		// Parse input JSON
		var input struct {
			Args   []interface{}          `json:"args"`
			Kwargs map[string]interface{} `json:"kwargs"`
		}
		if err := json.Unmarshal(inputJSON, &input); err != nil {
			return fmt.Errorf("failed to parse input JSON: %w", err)
		}

		taskArgs = input.Args
		taskKwargs = input.Kwargs
	}

	// Get current profile
	profile := viper.GetString("profile")

	// Load token
	tokenInfo, err := authcmd.LoadToken(profile)
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Check if token is valid
	if !authcmd.IsTokenValid(tokenInfo) {
		return fmt.Errorf("token is expired, please login again")
	}

	// Load client configuration
	_, err = config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create authorizer
	tokenAuthorizer := authorizers.NewStaticTokenAuthorizer(tokenInfo.AccessToken)
	coreAuthorizer := authorizers.ToCore(tokenAuthorizer)

	// Create compute client
	computeClient, err := compute.NewClient(
		compute.WithAuthorizer(coreAuthorizer),
	)
	if err != nil {
		return fmt.Errorf("failed to create compute client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build task request
	request := &compute.TaskRequest{
		FunctionID: functionID,
		EndpointID: endpointID,
		Args:       taskArgs,
		Kwargs:     taskKwargs,
	}

	// Run the function
	response, err := computeClient.RunFunction(ctx, request)
	if err != nil {
		return fmt.Errorf("error running function: %w", err)
	}

	// Display task information
	fmt.Fprintf(os.Stdout, "Function execution started successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Task ID:    %s\n", response.TaskID)
	fmt.Fprintf(os.Stdout, "\nMonitor task status with: globus compute task show %s\n", response.TaskID)

	return nil
}
