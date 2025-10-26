// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package search

import (
	"github.com/spf13/cobra"
)

// GetSubjectCmd returns the subject command
func GetSubjectCmd() *cobra.Command {
	subjectCmd := &cobra.Command{
		Use:   "subject",
		Short: "Manage subjects in Globus Search indices",
		Long: `Commands for working with specific subjects (documents) in Search indices.

A subject is a unique identifier for a document or entry in a Search index.

Examples:
  # Show a specific subject
  globus search subject show INDEX_ID SUBJECT_ID

  # Delete a subject
  globus search subject delete INDEX_ID SUBJECT_ID`,
	}

	// Add subcommands
	subjectCmd.AddCommand(SubjectShowCmd)
	subjectCmd.AddCommand(SubjectDeleteCmd)

	return subjectCmd
}
