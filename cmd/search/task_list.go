// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TaskListCmd represents the search task list command
var TaskListCmd = &cobra.Command{
	Use:   "list INDEX_ID",
	Short: "List recent tasks for a Globus Search index",
	Long: `List recent indexing and deletion tasks for a Globus Search index.

Examples:
  # List recent tasks
  globus search task list INDEX_ID

  # List with JSON output
  globus search task list INDEX_ID --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runTaskList,
}

func runTaskList(cmd *cobra.Command, args []string) error {
	indexID := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	searchClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	tasks, err := searchClient.GetTaskList(ctx, indexID)
	if err != nil {
		return fmt.Errorf("error listing tasks: %w", err)
	}

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, os.Stdout)
	if format != "text" {
		return formatter.FormatOutput(tasks, nil)
	}

	if len(tasks.Tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	type taskRow struct {
		TaskID  string
		State   string
		Created string
		Message string
	}
	rows := make([]taskRow, 0, len(tasks.Tasks))
	for _, t := range tasks.Tasks {
		created := ""
		if !t.Created.IsZero() {
			created = t.Created.Format(time.RFC3339)
		}
		rows = append(rows, taskRow{TaskID: t.TaskID, State: t.State, Created: created, Message: t.Message})
	}
	return formatter.FormatOutput(rows, []string{"TaskID", "State", "Created", "Message"})
}
