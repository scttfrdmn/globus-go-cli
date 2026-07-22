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

// IndexRoleListCmd represents the search index role list command
var IndexRoleListCmd = &cobra.Command{
	Use:   "list INDEX_ID",
	Short: "List roles on a Globus Search index",
	Long: `List all role assignments on a Globus Search index.

Examples:
  # List roles on an index
  globus search index role list INDEX_ID

  # List with JSON output
  globus search index role list INDEX_ID --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runIndexRoleList,
}

func runIndexRoleList(cmd *cobra.Command, args []string) error {
	indexID := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	searchClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	roles, err := searchClient.ListRoles(ctx, indexID)
	if err != nil {
		return fmt.Errorf("error listing roles: %w", err)
	}

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, os.Stdout)
	if format != "text" {
		return formatter.FormatOutput(roles, nil)
	}

	if len(roles.Roles) == 0 {
		fmt.Println("No roles found.")
		return nil
	}
	return formatter.FormatOutput(roles.Roles, []string{"ID", "RoleName", "Principal"})
}
