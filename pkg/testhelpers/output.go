// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package testhelpers

import (
	"bytes"
	"io"
	"os"
)

// CaptureOutput captures stdout and stderr during the execution of a function
// Returns captured stdout and stderr as strings
func CaptureOutput(f func()) (string, string) {
	// Save original stdout and stderr
	oldStdout, oldStderr := os.Stdout, os.Stderr

	// Create pipes for capturing output
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	// Redirect output to pipes
	os.Stdout, os.Stderr = wOut, wErr

	// Create channels to receive captured output
	outC, errC := make(chan string), make(chan string)

	// Goroutine to capture stdout
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rOut)
		outC <- buf.String()
	}()

	// Goroutine to capture stderr
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rErr)
		errC <- buf.String()
	}()

	// Execute the provided function
	f()

	// Close the write ends of the pipes to allow the goroutines to complete
	wOut.Close()
	wErr.Close()

	// Restore original stdout and stderr
	os.Stdout, os.Stderr = oldStdout, oldStderr

	// Return captured output
	return <-outC, <-errC
}
