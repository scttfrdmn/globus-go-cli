// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Compute client authorized for the current profile.
	computeClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Register function. The Compute API takes and returns open-ended documents.
	function, err := computeClient.RegisterFunction(ctx, map[string]interface{}{
		"function_name": registerName,
		"function_code": functionCode,
		"description":   registerDescription,
		"public":        registerPublic,
	})
	if err != nil {
		return fmt.Errorf("error registering function: %w", err)
	}

	// Display success message. The registered id is keyed under function_uuid.
	functionID := mapStr(function, "function_uuid")
	if functionID == "" {
		functionID = mapStr(function, "function_id")
	}
	fmt.Fprintf(os.Stdout, "Function registered successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Function ID:   %s\n", functionID)
	if n := mapStr(function, "function_name"); n != "" {
		fmt.Fprintf(os.Stdout, "Name:          %s\n", n)
	}

	return nil
}
