// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package main

import (
	"fmt"
	"os"

	"github.com/scttfrdmn/globus-go-cli/cmd"
)

func main() {
	code, err := cmd.ExitCode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(code)
}
