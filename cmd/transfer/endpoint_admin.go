// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
)

// endpointUpdateCmd returns the "endpoint update" command, which patches an
// endpoint document with only the fields the user explicitly set.
func endpointUpdateCmd() *cobra.Command {
	var (
		displayName  string
		description  string
		organization string
		contactEmail string
		keywords     []string
		public       bool
	)

	cmd := &cobra.Command{
		Use:   "update ENDPOINT_ID",
		Short: "Update a Globus endpoint",
		Long: `Update mutable fields on a Globus endpoint.

Only the flags you provide are sent; unset fields are left unchanged.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			doc := map[string]interface{}{"DATA_TYPE": "endpoint"}
			if cmd.Flags().Changed("display-name") {
				doc["display_name"] = displayName
			}
			if cmd.Flags().Changed("description") {
				doc["description"] = description
			}
			if cmd.Flags().Changed("organization") {
				doc["organization"] = organization
			}
			if cmd.Flags().Changed("contact-email") {
				doc["contact_email"] = contactEmail
			}
			if cmd.Flags().Changed("keywords") {
				doc["keywords"] = keywords
			}
			if cmd.Flags().Changed("public") {
				doc["public"] = public
			}

			resp, err := client.UpdateEndpoint(ctx, args[0], doc)
			if err != nil {
				return fmt.Errorf("failed to update endpoint: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Endpoint %s updated.\n", args[0])
			printResponseCodeMessage(cmd, resp)
			return nil
		},
	}

	cmd.Flags().StringVar(&displayName, "display-name", "", "New display name")
	cmd.Flags().StringVar(&description, "description", "", "New description")
	cmd.Flags().StringVar(&organization, "organization", "", "New organization")
	cmd.Flags().StringVar(&contactEmail, "contact-email", "", "New contact email")
	cmd.Flags().StringSliceVar(&keywords, "keywords", nil, "Keywords for the endpoint")
	cmd.Flags().BoolVar(&public, "public", false, "Whether the endpoint is public")

	return cmd
}

// endpointDeleteCmd returns the "endpoint delete" command.
func endpointDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ENDPOINT_ID",
		Short: "Delete a Globus endpoint",
		Long:  `Delete a Globus endpoint by ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.DeleteEndpoint(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to delete endpoint: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Endpoint %s deleted.\n", args[0])
			printResponseCodeMessage(cmd, resp)
			return nil
		},
	}
}

// endpointRoleCmd returns the "endpoint role" command group for managing
// endpoint role assignments.
func endpointRoleCmd() *cobra.Command {
	roleCmd := &cobra.Command{
		Use:   "role",
		Short: "Manage endpoint role assignments",
		Long:  `List, show, create, and delete role assignments on a Globus endpoint.`,
	}

	roleCmd.AddCommand(
		endpointRoleListCmd(),
		endpointRoleShowCmd(),
		endpointRoleCreateCmd(),
		endpointRoleDeleteCmd(),
	)

	return roleCmd
}

func endpointRoleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list ENDPOINT_ID",
		Short: "List role assignments for an endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointRoleList(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to list endpoint roles: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointRoleShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show ENDPOINT_ID ROLE_ID",
		Short: "Show a role assignment for an endpoint",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.GetEndpointRole(ctx, args[0], args[1])
			if err != nil {
				return fmt.Errorf("failed to get endpoint role: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointRoleCreateCmd() *cobra.Command {
	var (
		principal     string
		principalType string
		role          string
	)

	cmd := &cobra.Command{
		Use:   "create ENDPOINT_ID",
		Short: "Create a role assignment on an endpoint",
		Long: `Assign a role to a principal on a Globus endpoint.

The --role value is one of administrator, access_manager, activity_manager,
or activity_monitor.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			doc := map[string]interface{}{
				"DATA_TYPE":      "role",
				"principal_type": principalType,
				"principal":      principal,
				"role":           role,
			}

			resp, err := client.AddEndpointRole(ctx, args[0], doc)
			if err != nil {
				return fmt.Errorf("failed to create endpoint role: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&principal, "principal", "", "Principal (identity or group ID) to assign the role to")
	cmd.Flags().StringVar(&principalType, "principal-type", "identity", "Principal type (identity or group)")
	cmd.Flags().StringVar(&role, "role", "", "Role name (administrator, access_manager, activity_manager, activity_monitor)")
	_ = cmd.MarkFlagRequired("principal")
	_ = cmd.MarkFlagRequired("role")

	return cmd
}

func endpointRoleDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ENDPOINT_ID ROLE_ID",
		Short: "Delete a role assignment from an endpoint",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.DeleteEndpointRole(ctx, args[0], args[1])
			if err != nil {
				return fmt.Errorf("failed to delete endpoint role: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Role %s deleted from endpoint %s.\n", args[1], args[0])
			printResponseCodeMessage(cmd, resp)
			return nil
		},
	}
}

// endpointPermissionCmd returns the "endpoint permission" command group for
// managing endpoint access rules (ACLs).
func endpointPermissionCmd() *cobra.Command {
	permCmd := &cobra.Command{
		Use:   "permission",
		Short: "Manage endpoint access rules (ACLs)",
		Long:  `List, show, create, update, and delete access rules on a Globus endpoint.`,
	}

	permCmd.AddCommand(
		endpointPermissionListCmd(),
		endpointPermissionShowCmd(),
		endpointPermissionCreateCmd(),
		endpointPermissionUpdateCmd(),
		endpointPermissionDeleteCmd(),
	)

	return permCmd
}

func endpointPermissionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list ENDPOINT_ID",
		Short: "List access rules for an endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointACLList(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to list endpoint access rules: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointPermissionShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show ENDPOINT_ID RULE_ID",
		Short: "Show an access rule for an endpoint",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.GetEndpointACLRule(ctx, args[0], args[1])
			if err != nil {
				return fmt.Errorf("failed to get endpoint access rule: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointPermissionCreateCmd() *cobra.Command {
	var (
		permissions   string
		principal     string
		principalType string
		path          string
	)

	cmd := &cobra.Command{
		Use:   "create ENDPOINT_ID",
		Short: "Create an access rule on an endpoint",
		Long: `Create an access rule (ACL) granting permissions on a path.

The --principal-type value is one of identity, group,
all_authenticated_users, or anonymous.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			doc := map[string]interface{}{
				"DATA_TYPE":      "access",
				"principal_type": principalType,
				"principal":      principal,
				"path":           path,
				"permissions":    permissions,
			}

			resp, err := client.AddEndpointACLRule(ctx, args[0], doc)
			if err != nil {
				return fmt.Errorf("failed to create endpoint access rule: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&permissions, "permissions", "", `Permissions to grant (e.g. "r" or "rw")`)
	cmd.Flags().StringVar(&principal, "principal", "", "Principal (identity or group ID) to grant access to")
	cmd.Flags().StringVar(&principalType, "principal-type", "identity", "Principal type (identity, group, all_authenticated_users, anonymous)")
	cmd.Flags().StringVar(&path, "path", "", "Path the rule applies to")
	_ = cmd.MarkFlagRequired("permissions")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}

func endpointPermissionUpdateCmd() *cobra.Command {
	var permissions string

	cmd := &cobra.Command{
		Use:   "update ENDPOINT_ID RULE_ID",
		Short: "Update an access rule on an endpoint",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			doc := map[string]interface{}{
				"DATA_TYPE":   "access",
				"permissions": permissions,
			}

			resp, err := client.UpdateEndpointACLRule(ctx, args[0], args[1], doc)
			if err != nil {
				return fmt.Errorf("failed to update endpoint access rule: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&permissions, "permissions", "", `New permissions (e.g. "r" or "rw")`)
	_ = cmd.MarkFlagRequired("permissions")

	return cmd
}

func endpointPermissionDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ENDPOINT_ID RULE_ID",
		Short: "Delete an access rule from an endpoint",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.DeleteEndpointACLRule(ctx, args[0], args[1])
			if err != nil {
				return fmt.Errorf("failed to delete endpoint access rule: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Access rule %s deleted from endpoint %s.\n", args[1], args[0])
			printResponseCodeMessage(cmd, resp)
			return nil
		},
	}
}

// formatGenericResponse routes a GenericResponse (map[string]interface{})
// through the shared formatter so -F (text/json/unix) and --jmespath/--jq
// work uniformly. For text/unix, the whole map is emitted.
func formatGenericResponse(cmd *cobra.Command, resp map[string]interface{}) error {
	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	return formatter.FormatOutput(resp, nil)
}

// printResponseCodeMessage prints the "code" and "message" fields of a Transfer
// GenericResponse when present, for success/status feedback on mutations.
func printResponseCodeMessage(cmd *cobra.Command, resp map[string]interface{}) {
	if resp == nil {
		return
	}
	if code, ok := resp["code"].(string); ok && code != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  Code:    %s\n", code)
	}
	if msg, ok := resp["message"].(string); ok && msg != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  Message: %s\n", msg)
	}
}
