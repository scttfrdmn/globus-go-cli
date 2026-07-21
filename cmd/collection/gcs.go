// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/gcs"
)

// GCSCmd returns the gcs command tree for endpoint-level GCS Manager admin.
//
// Every subcommand takes the owning endpoint's ID as its first positional
// argument, used by getManagerClient to resolve the GCS Manager URL and
// manage_collections consent.
func GCSCmd() *cobra.Command {
	gcsCmd := &cobra.Command{
		Use:   "gcs",
		Short: "Commands for administering a Globus Connect Server endpoint",
		Long: `Endpoint-level administration of a Globus Connect Server v5 endpoint's
GCS Manager, including server info, storage gateways, and role assignments.

Every subcommand takes the endpoint's ID as its first argument.`,
	}

	gcsCmd.AddCommand(
		gcsInfoCmd(),
		gcsStorageGatewayCmd(),
		gcsRoleCmd(),
	)

	return gcsCmd
}

// Options for role creation.
var (
	rolePrincipal  string
	roleName       string
	roleCollection string
)

// gcsInfoCmd returns the gcs info command.
func gcsInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info ENDPOINT_ID",
		Short: "Show GCS Manager info",
		Long:  `Show the GCS Manager info document (client ID and endpoint ID) for the given endpoint.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return gcsInfo(cmd, args[0])
		},
	}
}

// gcsStorageGatewayCmd returns the storage-gateway subgroup.
func gcsStorageGatewayCmd() *cobra.Command {
	sgCmd := &cobra.Command{
		Use:   "storage-gateway",
		Short: "Commands for managing storage gateways",
		Long:  `List and show storage gateways on the given endpoint's GCS Manager.`,
	}

	sgCmd.AddCommand(
		storageGatewayListCmd(),
		storageGatewayShowCmd(),
	)

	return sgCmd
}

// storageGatewayListCmd returns the storage-gateway list command.
func storageGatewayListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list ENDPOINT_ID",
		Short: "List storage gateways",
		Long:  `List the storage gateways on the given endpoint's GCS Manager.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listStorageGateways(cmd, args[0])
		},
	}
}

// storageGatewayShowCmd returns the storage-gateway show command.
func storageGatewayShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show ENDPOINT_ID STORAGE_GATEWAY_ID",
		Short: "Show storage gateway details",
		Long:  `Show detailed information about a storage gateway on the given endpoint's GCS Manager.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showStorageGateway(cmd, args[0], args[1])
		},
	}
}

// gcsRoleCmd returns the role subgroup.
func gcsRoleCmd() *cobra.Command {
	roleCmd := &cobra.Command{
		Use:   "role",
		Short: "Commands for managing role assignments",
		Long:  `List, show, create, and delete role assignments on the given endpoint's GCS Manager.`,
	}

	roleCmd.AddCommand(
		roleListCmd(),
		roleShowCmd(),
		roleCreateCmd(),
		roleDeleteCmd(),
	)

	return roleCmd
}

// roleListCmd returns the role list command.
func roleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list ENDPOINT_ID",
		Short: "List role assignments",
		Long:  `List the role assignments on the given endpoint's GCS Manager.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listRoles(cmd, args[0])
		},
	}
}

// roleShowCmd returns the role show command.
func roleShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show ENDPOINT_ID ROLE_ID",
		Short: "Show role assignment details",
		Long:  `Show detailed information about a role assignment on the given endpoint's GCS Manager.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showRole(cmd, args[0], args[1])
		},
	}
}

// roleCreateCmd returns the role create command.
func roleCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create ENDPOINT_ID",
		Short: "Create a role assignment",
		Long: `Create a role assignment on the given endpoint's GCS Manager.

Omit --collection to assign an endpoint-level role.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createRole(cmd, args[0])
		},
	}

	cmd.Flags().StringVar(&rolePrincipal, "principal", "", "Principal (identity or group URN) to assign the role to (required)")
	cmd.Flags().StringVar(&roleName, "role", "", "Role name (owner, administrator, access_manager, activity_manager, activity_monitor) (required)")
	cmd.Flags().StringVar(&roleCollection, "collection", "", "Collection ID for a collection-level role (omit for an endpoint role)")
	_ = cmd.MarkFlagRequired("principal")
	_ = cmd.MarkFlagRequired("role")

	return cmd
}

// roleDeleteCmd returns the role delete command.
func roleDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ENDPOINT_ID ROLE_ID",
		Short: "Delete a role assignment",
		Long:  `Delete a role assignment from the given endpoint's GCS Manager.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteRole(cmd, args[0], args[1])
		},
	}
}

// gcsInfo shows the GCS Manager info document.
func gcsInfo(cmd *cobra.Command, endpointID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getManagerClient(ctx, endpointID)
	if err != nil {
		return err
	}

	info, err := client.GetGCSInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get GCS info: %w", err)
	}

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(info, nil)
	}

	type infoRow struct {
		ClientID   string
		EndpointID string
	}
	return formatter.FormatOutput(
		[]infoRow{{ClientID: info.ClientID, EndpointID: info.EndpointID}},
		[]string{"ClientID", "EndpointID"},
	)
}

// listStorageGateways lists the storage gateways on an endpoint's GCS Manager.
func listStorageGateways(cmd *cobra.Command, endpointID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getManagerClient(ctx, endpointID)
	if err != nil {
		return err
	}

	resp, err := client.GetStorageGatewayList(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list storage gateways: %w", err)
	}

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())

	if formatter.Format == output.FormatJSON {
		return formatter.FormatOutput(resp, nil)
	}

	type gatewayRow struct {
		ID            string
		DisplayName   string
		ConnectorID   string
		HighAssurance bool
	}
	rows := make([]gatewayRow, 0, len(resp.Data))
	for _, g := range resp.Data {
		rows = append(rows, gatewayRow{
			ID: g.ID, DisplayName: g.DisplayName,
			ConnectorID: g.ConnectorID, HighAssurance: g.HighAssurance,
		})
	}
	return formatter.FormatOutput(rows, []string{"ID", "DisplayName", "ConnectorID", "HighAssurance"})
}

// showStorageGateway shows details for a single storage gateway.
func showStorageGateway(cmd *cobra.Command, endpointID, storageGatewayID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getManagerClient(ctx, endpointID)
	if err != nil {
		return err
	}

	gw, err := client.GetStorageGateway(ctx, storageGatewayID, nil)
	if err != nil {
		return fmt.Errorf("failed to get storage gateway: %w", err)
	}

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(gw, nil)
	}

	fmt.Println("Storage Gateway Details:")
	fmt.Printf("  ID:              %s\n", gw.ID)
	fmt.Printf("  Display Name:    %s\n", gw.DisplayName)
	fmt.Printf("  Connector ID:    %s\n", gw.ConnectorID)
	fmt.Printf("  High Assurance:  %t\n", gw.HighAssurance)
	fmt.Printf("  Require MFA:     %t\n", gw.RequireMFA)

	return nil
}

// listRoles lists the role assignments on an endpoint's GCS Manager.
func listRoles(cmd *cobra.Command, endpointID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getManagerClient(ctx, endpointID)
	if err != nil {
		return err
	}

	resp, err := client.GetRoleList(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())

	if formatter.Format == output.FormatJSON {
		return formatter.FormatOutput(resp, nil)
	}

	type roleRow struct {
		ID         string
		Principal  string
		Role       string
		Collection string
	}
	rows := make([]roleRow, 0, len(resp.Data))
	for _, r := range resp.Data {
		collection := ""
		if r.Collection != nil {
			collection = *r.Collection
		}
		rows = append(rows, roleRow{
			ID: r.ID, Principal: r.Principal, Role: r.Role, Collection: collection,
		})
	}
	return formatter.FormatOutput(rows, []string{"ID", "Principal", "Role", "Collection"})
}

// showRole shows details for a single role assignment.
func showRole(cmd *cobra.Command, endpointID, roleID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getManagerClient(ctx, endpointID)
	if err != nil {
		return err
	}

	role, err := client.GetRole(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(role, nil)
	}

	collection := "(endpoint)"
	if role.Collection != nil {
		collection = *role.Collection
	}
	fmt.Println("Role Details:")
	fmt.Printf("  ID:          %s\n", role.ID)
	fmt.Printf("  Principal:   %s\n", role.Principal)
	fmt.Printf("  Role:        %s\n", role.Role)
	fmt.Printf("  Collection:  %s\n", collection)

	return nil
}

// createRole creates a role assignment.
func createRole(cmd *cobra.Command, endpointID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getManagerClient(ctx, endpointID)
	if err != nil {
		return err
	}

	doc := &gcs.GCSRoleDocument{
		Principal: rolePrincipal,
		Role:      roleName,
	}
	if roleCollection != "" {
		doc.Collection = roleCollection
	}

	role, err := client.CreateRole(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	fmt.Printf("Created role %s (%s for %s)\n", role.ID, role.Role, role.Principal)
	return nil
}

// deleteRole deletes a role assignment.
func deleteRole(cmd *cobra.Command, endpointID, roleID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getManagerClient(ctx, endpointID)
	if err != nil {
		return err
	}

	if err := client.DeleteRole(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	fmt.Printf("Deleted role %s\n", roleID)
	return nil
}
