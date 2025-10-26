// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
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
	registerName        string
	registerDescription string
	registerFile        string
	registerCode        string
	registerPublic      bool
)

// FunctionRegisterCmd represents the compute function register command
var FunctionRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new function with Globus Compute",
	Long: `Register a Python function with Globus Compute.

You must provide the function code either from a file or as a string.
The function will be serialized and stored for later execution.

Examples:
  # Register a function from a file
  globus compute function register --name "my_function" --file function.py

  # Register with inline code
  globus compute function register --name "simple" --code "def add(a, b): return a + b"

  # Register a public function
  globus compute function register --name "public_func" --file func.py --public`,
	RunE: runFunctionRegister,
}

func init() {
	FunctionRegisterCmd.Flags().StringVar(&registerName, "name", "", "Function name")
	FunctionRegisterCmd.Flags().StringVar(&registerDescription, "description", "", "Function description")
	FunctionRegisterCmd.Flags().StringVar(&registerFile, "file", "", "Path to Python file containing function")
	FunctionRegisterCmd.Flags().StringVar(&registerCode, "code", "", "Function code as string")
	FunctionRegisterCmd.Flags().BoolVar(&registerPublic, "public", false, "Make function publicly accessible")
}

func runFunctionRegister(cmd *cobra.Command, args []string) error {
	// Validate input
	if registerFile == "" && registerCode == "" {
		return fmt.Errorf("either --file or --code must be provided")
	}
	if registerFile != "" && registerCode != "" {
		return fmt.Errorf("cannot specify both --file and --code")
	}

	// Read function code
	var functionCode string
	if registerFile != "" {
		codeBytes, err := os.ReadFile(registerFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		functionCode = string(codeBytes)
	} else {
		functionCode = registerCode
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

	// Build register request
	request := &compute.FunctionRegisterRequest{
		Function:    functionCode,
		Name:        registerName,
		Description: registerDescription,
		Public:      registerPublic,
	}

	// Register function
	function, err := computeClient.RegisterFunction(ctx, request)
	if err != nil {
		return fmt.Errorf("error registering function: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Function registered successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Function ID:   %s\n", function.ID)
	if function.Name != "" {
		fmt.Fprintf(os.Stdout, "Name:          %s\n", function.Name)
	}
	fmt.Fprintf(os.Stdout, "Public:        %t\n", function.Public)
	fmt.Fprintf(os.Stdout, "Created:       %s\n", function.CreatedAt.Format(time.RFC3339))

	return nil
}
