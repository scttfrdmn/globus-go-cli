// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/search"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	roleCreateRole string
)

// IndexRoleCreateCmd represents the search index role create command
var IndexRoleCreateCmd = &cobra.Command{
	Use:   "create INDEX_ID PRINCIPAL",
	Short: "Create a role on a Globus Search index",
	Long: `Create a new role granting permissions to a principal on an index.

PRINCIPAL is a Globus Auth identity or group URN, e.g.
urn:globus:auth:identity:<id> or urn:globus:groups:id:<group-id>.

Examples:
  # Grant admin role to an identity
  globus search index role create INDEX_ID urn:globus:auth:identity:USER_ID --role admin

  # Grant writer role to a group
  globus search index role create INDEX_ID urn:globus:groups:id:GROUP_ID --role writer`,
	Args: cobra.ExactArgs(2),
	RunE: runIndexRoleCreate,
}

func init() {
	IndexRoleCreateCmd.Flags().StringVar(&roleCreateRole, "role", "reader", "Role name (owner, admin, writer, reader)")
}

func runIndexRoleCreate(cmd *cobra.Command, args []string) error {
	indexID := args[0]
	principal := args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	searchClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	role, err := searchClient.AddRole(ctx, indexID, &search.RoleCreate{
		RoleName:  roleCreateRole,
		Principal: principal,
	})
	if err != nil {
		return fmt.Errorf("error creating role: %w", err)
	}

	format := viper.GetString("format")
	if format != "text" {
		return output.NewFormatter(format, os.Stdout).FormatOutput(role, nil)
	}
	fmt.Printf("Created role %s (%s) for %s on index %s\n", role.ID, role.RoleName, role.Principal, role.IndexID)
	return nil
}
