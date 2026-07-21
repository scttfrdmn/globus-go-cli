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

// CollectionCmd returns the collection command tree.
//
// Every subcommand takes the owning endpoint's ID as its first positional
// argument: GCS Manager operations are scoped to a specific endpoint, and
// getManagerClient resolves that endpoint's GCS Manager URL and manage_collections
// consent from it.
func CollectionCmd() *cobra.Command {
	collectionCmd := &cobra.Command{
		Use:   "collection",
		Short: "Commands for managing Globus Connect Server collections",
		Long: `Commands for managing collections on a Globus Connect Server v5
endpoint's GCS Manager, including listing, showing, creating, updating,
and deleting mapped and guest collections.

Every subcommand takes the owning endpoint's ID as its first argument. The
CLI resolves that endpoint's GCS Manager URL and escalates the endpoint's
manage_collections consent on first use (a paste-code login prompt).`,
	}

	collectionCmd.AddCommand(
		collectionListCmd(),
		collectionShowCmd(),
		collectionCreateCmd(),
		collectionUpdateCmd(),
		collectionDeleteCmd(),
	)

	return collectionCmd
}

// Options for collection listing and mutation.
var (
	collectionMappedID     string
	collectionPageSize     int
	collectionType         string
	collectionDisplayName  string
	collectionStorageGWID  string
	collectionBasePath     string
	collectionIdentityID   string
	collectionPublic       bool
	collectionDescription  string
	collectionOrganization string
)

// collectionListCmd returns the collection list command.
func collectionListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list ENDPOINT_ID",
		Short: "List collections on a GCS endpoint",
		Long: `List the collections hosted by the given endpoint's GCS Manager.

Optionally filter to the guest collections of a single mapped collection.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listCollections(cmd, args[0])
		},
	}

	cmd.Flags().StringVar(&collectionMappedID, "mapped-collection-id", "", "Filter to guest collections of this mapped collection")
	cmd.Flags().IntVar(&collectionPageSize, "page-size", 100, "Maximum number of collections to return")

	return cmd
}

// collectionShowCmd returns the collection show command.
func collectionShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show ENDPOINT_ID COLLECTION_ID",
		Short: "Show collection details",
		Long:  `Show detailed information about a collection on the given endpoint's GCS Manager.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCollection(cmd, args[0], args[1])
		},
	}
}

// collectionCreateCmd returns the collection create command.
func collectionCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create ENDPOINT_ID",
		Short: "Create a collection",
		Long: `Create a mapped or guest collection on the given endpoint's GCS Manager.

Mapped collections require --storage-gateway-id; guest collections require
--mapped-collection-id (and typically --base-path).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createCollection(cmd, args[0])
		},
	}

	cmd.Flags().StringVar(&collectionType, "type", "", "Collection type: mapped or guest (required)")
	cmd.Flags().StringVar(&collectionDisplayName, "display-name", "", "Display name for the collection (required)")
	cmd.Flags().StringVar(&collectionStorageGWID, "storage-gateway-id", "", "Storage gateway ID (required for mapped collections)")
	cmd.Flags().StringVar(&collectionBasePath, "base-path", "", "Collection base path")
	cmd.Flags().StringVar(&collectionMappedID, "mapped-collection-id", "", "Mapped collection ID (for guest collections)")
	cmd.Flags().StringVar(&collectionIdentityID, "identity-id", "", "Identity ID that owns the collection")
	cmd.Flags().BoolVar(&collectionPublic, "public", false, "Make the collection publicly visible")
	cmd.Flags().StringVar(&collectionDescription, "description", "", "Description of the collection")
	cmd.Flags().StringVar(&collectionOrganization, "organization", "", "Organization for the collection")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("display-name")

	return cmd
}

// collectionUpdateCmd returns the collection update command.
func collectionUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update ENDPOINT_ID COLLECTION_ID",
		Short: "Update a collection",
		Long: `Update mutable fields of a collection on the given endpoint's GCS Manager.

Only the fields you supply as flags are changed.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateCollection(cmd, args[0], args[1])
		},
	}

	cmd.Flags().StringVar(&collectionDisplayName, "display-name", "", "New display name")
	cmd.Flags().StringVar(&collectionDescription, "description", "", "New description")
	cmd.Flags().StringVar(&collectionOrganization, "organization", "", "New organization")
	cmd.Flags().BoolVar(&collectionPublic, "public", false, "Set public visibility")

	return cmd
}

// collectionDeleteCmd returns the collection delete command.
func collectionDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ENDPOINT_ID COLLECTION_ID",
		Short: "Delete a collection",
		Long:  `Permanently delete a collection from the given endpoint's GCS Manager.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteCollection(cmd, args[0], args[1])
		},
	}
}

// listCollections lists the collections on an endpoint's GCS Manager.
func listCollections(cmd *cobra.Command, endpointID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getManagerClient(ctx, endpointID)
	if err != nil {
		return err
	}

	opts := &gcs.ListCollectionsOptions{
		MappedCollectionID: collectionMappedID,
		PageSize:           collectionPageSize,
	}

	resp, err := client.ListCollections(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	// Route all formats through the shared formatter so -F (text/json/unix) and
	// --jmespath/--jq work uniformly. For JSON, emit the raw GCS response
	// envelope; for text/unix, a projected row set.
	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())

	if formatter.Format == output.FormatJSON {
		return formatter.FormatOutput(resp, nil)
	}

	type collectionRow struct {
		ID             string
		DisplayName    string
		Type           string
		StorageGateway string
		Public         bool
	}
	rows := make([]collectionRow, 0, len(resp.Data))
	for _, c := range resp.Data {
		rows = append(rows, collectionRow{
			ID: c.ID, DisplayName: c.DisplayName, Type: c.CollectionType,
			StorageGateway: c.StorageGatewayID, Public: c.PubliclyVisible,
		})
	}
	return formatter.FormatOutput(rows, []string{"ID", "DisplayName", "Type", "StorageGateway", "Public"})
}

// showCollection shows details for a single collection.
func showCollection(cmd *cobra.Command, endpointID, collectionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getManagerClient(ctx, endpointID)
	if err != nil {
		return err
	}

	col, err := client.GetCollection(ctx, collectionID, nil)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(col, nil)
	}

	fmt.Println("Collection Details:")
	fmt.Printf("  ID:                  %s\n", col.ID)
	fmt.Printf("  Display Name:        %s\n", col.DisplayName)
	fmt.Printf("  Type:                %s\n", col.CollectionType)
	fmt.Printf("  Public:              %t\n", col.PubliclyVisible)
	fmt.Printf("  High Assurance:      %t\n", col.HighAssurance)
	if col.StorageGatewayID != "" {
		fmt.Printf("  Storage Gateway:     %s\n", col.StorageGatewayID)
	}
	if col.MappedCollectionID != "" {
		fmt.Printf("  Mapped Collection:   %s\n", col.MappedCollectionID)
	}
	if col.Organization != "" {
		fmt.Printf("  Organization:        %s\n", col.Organization)
	}
	if col.ContactEmail != "" {
		fmt.Printf("  Contact Email:       %s\n", col.ContactEmail)
	}
	if col.DomainName != "" {
		fmt.Printf("  Domain Name:         %s\n", col.DomainName)
	}
	if col.HTTPSURL != "" {
		fmt.Printf("  HTTPS URL:           %s\n", col.HTTPSURL)
	}

	return nil
}

// createCollection creates a collection, setting only the fields provided.
func createCollection(cmd *cobra.Command, endpointID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getManagerClient(ctx, endpointID)
	if err != nil {
		return err
	}

	// Build the document from the provided fields only. DataType is left empty:
	// the SDK sets the appropriate default version.
	doc := &gcs.CollectionDocument{
		CollectionType: collectionType,
		DisplayName:    collectionDisplayName,
	}
	if collectionStorageGWID != "" {
		doc.StorageGatewayID = collectionStorageGWID
	}
	if collectionBasePath != "" {
		doc.CollectionBasePath = collectionBasePath
	}
	if collectionMappedID != "" {
		doc.MappedCollectionID = collectionMappedID
	}
	if collectionIdentityID != "" {
		doc.IdentityID = collectionIdentityID
	}
	if collectionOrganization != "" {
		doc.Organization = collectionOrganization
	}
	if cmd.Flags().Changed("public") {
		public := collectionPublic
		doc.Public = &public
	}
	if cmd.Flags().Changed("description") {
		desc := collectionDescription
		doc.Description = &desc
	}

	col, err := client.CreateCollection(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	fmt.Printf("Created collection %s (%s)\n", col.ID, col.DisplayName)
	return nil
}

// updateCollection updates a collection with only the provided fields.
func updateCollection(cmd *cobra.Command, endpointID, collectionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getManagerClient(ctx, endpointID)
	if err != nil {
		return err
	}

	doc := &gcs.CollectionDocument{}
	if cmd.Flags().Changed("display-name") {
		doc.DisplayName = collectionDisplayName
	}
	if cmd.Flags().Changed("organization") {
		doc.Organization = collectionOrganization
	}
	if cmd.Flags().Changed("public") {
		public := collectionPublic
		doc.Public = &public
	}
	if cmd.Flags().Changed("description") {
		desc := collectionDescription
		doc.Description = &desc
	}

	if _, err := client.UpdateCollection(ctx, collectionID, doc); err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	fmt.Printf("Updated collection %s\n", collectionID)
	return nil
}

// deleteCollection deletes a collection.
func deleteCollection(cmd *cobra.Command, endpointID, collectionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getManagerClient(ctx, endpointID)
	if err != nil {
		return err
	}

	if err := client.DeleteCollection(ctx, collectionID); err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	fmt.Printf("Deleted collection %s\n", collectionID)
	return nil
}
