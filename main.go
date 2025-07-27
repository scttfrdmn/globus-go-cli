// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"fmt"
	"os"

	"github.com/scttfrdmn/globus-go-cli/cmd"
)

func main() {
	if err := cmd.ExecuteCmd(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
