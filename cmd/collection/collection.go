// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package collection

import (
	"context"
	"encoding/json"
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
		collectionCatCmd(),
	)

	return collectionCmd
}

// Options for collection listing and mutation.
var (
	collectionMappedID      string
	collectionPageSize      int
	collectionLimit         int
	collectionFilter        []string
	collectionInclPrivate   bool
	collectionType          string
	collectionDisplayName   string
	collectionStorageGWID   string
	collectionBasePath      string
	collectionIdentityID    string
	collectionUserCredID    string
	collectionPublic        bool
	collectionPrivate       bool
	collectionDescription   string
	collectionOrganization  string
	collectionDepartment    string
	collectionContactEmail  string
	collectionContactInfo   string
	collectionInfoLink      string
	collectionKeywords      []string
	collectionForceEnc      bool
	collectionNoForceEnc    bool
	collectionDefaultDir    string
	collectionDomainName    string
	collectionUserMessage   string
	collectionUserMsgLink   string
	collectionVerify        string
	collectionEnableHTTPS   bool
	collectionDisableHTTPS  bool
	collectionAllowGuest    bool
	collectionNoAllowGuest  bool
	collectionShareRestrict string
	collectionShareAllow    []string
	collectionShareDeny     []string
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
	cmd.Flags().StringSliceVar(&collectionFilter, "filter", nil, "Filter results to categories of collections (mapped-collections, guest-collections, managed-by-me, created-by-me); may be given multiple times")
	cmd.Flags().BoolVar(&collectionInclPrivate, "include-private-policies", false, "Include private policies (requires administrator role)")
	cmd.Flags().IntVar(&collectionLimit, "limit", 25, "Maximum number of collections to return")
	cmd.Flags().IntVar(&collectionPageSize, "page-size", 100, "GCS Manager page size")

	return cmd
}

// collectionShowCmd returns the collection show command.
func collectionShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show ENDPOINT_ID COLLECTION_ID",
		Short: "Show collection details",
		Long:  `Show detailed information about a collection on the given endpoint's GCS Manager.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCollection(cmd, args[0], args[1])
		},
	}

	cmd.Flags().BoolVar(&collectionInclPrivate, "include-private-policies", false, "Include private policies (requires administrator role)")

	return cmd
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
	cmd.Flags().StringVar(&collectionBasePath, "base-path", "", "The location within the storage gateway where the collection is rooted")
	cmd.Flags().StringVar(&collectionMappedID, "mapped-collection-id", "", "Mapped collection ID (for guest collections)")
	cmd.Flags().StringVar(&collectionUserCredID, "user-credential-id", "", "Registered local user credential to associate with a guest collection")
	cmd.Flags().StringVar(&collectionIdentityID, "identity-id", "", "User who should own the collection (defaults to the current user)")
	cmd.Flags().BoolVar(&collectionPublic, "public", false, "Set the collection to be public")
	cmd.Flags().BoolVar(&collectionPrivate, "private", false, "Set the collection to be private")
	cmd.Flags().StringVar(&collectionDescription, "description", "", "Description for the collection")
	cmd.Flags().StringVar(&collectionOrganization, "organization", "", "Organization for the collection")
	cmd.Flags().StringVar(&collectionDepartment, "department", "", "Department which operates the collection")
	cmd.Flags().StringVar(&collectionContactEmail, "contact-email", "", "Contact email for the collection")
	cmd.Flags().StringVar(&collectionContactInfo, "contact-info", "", "Contact info for the collection")
	cmd.Flags().StringVar(&collectionInfoLink, "info-link", "", "Link for info about the collection")
	cmd.Flags().StringSliceVar(&collectionKeywords, "keywords", nil, "Comma separated list of keywords to help searches for the collection")
	cmd.Flags().BoolVar(&collectionForceEnc, "force-encryption", false, "Force the collection to encrypt transfers")
	cmd.Flags().BoolVar(&collectionNoForceEnc, "no-force-encryption", false, "Do not force the collection to encrypt transfers")
	cmd.Flags().StringVar(&collectionDefaultDir, "default-directory", "", "Default directory when browsing or executing tasks on the collection")
	cmd.Flags().StringVar(&collectionDomainName, "domain-name", "", "DNS host name for the collection (mapped collections only)")
	cmd.Flags().StringVar(&collectionUserMessage, "user-message", "", "A message for clients to display to users when interacting with this collection")
	cmd.Flags().StringVar(&collectionUserMsgLink, "user-message-link", "", "Link to additional messaging for clients to display to users")
	cmd.Flags().StringVar(&collectionVerify, "verify", "", "File integrity verification policy: force, disable, or default")
	cmd.Flags().BoolVar(&collectionEnableHTTPS, "enable-https", false, "Explicitly enable HTTPS support")
	cmd.Flags().BoolVar(&collectionDisableHTTPS, "disable-https", false, "Explicitly disable HTTPS support")
	cmd.Flags().BoolVar(&collectionAllowGuest, "allow-guest-collections", false, "Allow guest collections on this mapped collection")
	cmd.Flags().BoolVar(&collectionNoAllowGuest, "no-allow-guest-collections", false, "Disallow guest collections on this mapped collection")
	cmd.Flags().StringVar(&collectionShareRestrict, "sharing-restrict-paths", "", "JSON path restrictions for sharing on guest collections (mapped collections only)")
	cmd.Flags().StringArrayVar(&collectionShareAllow, "sharing-user-allow", nil, "Connector-specific username allowed to create guest collections; may be given multiple times")
	cmd.Flags().StringArrayVar(&collectionShareDeny, "sharing-user-deny", nil, "Connector-specific username denied permission to create guest collections; may be given multiple times")
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

	cmd.Flags().StringVar(&collectionDisplayName, "display-name", "", "Name for the collection")
	cmd.Flags().StringVar(&collectionDescription, "description", "", "Description for the collection")
	cmd.Flags().StringVar(&collectionOrganization, "organization", "", "Organization for the collection")
	cmd.Flags().StringVar(&collectionDepartment, "department", "", "Department which operates the collection")
	cmd.Flags().StringVar(&collectionContactEmail, "contact-email", "", "Contact email for the collection")
	cmd.Flags().StringVar(&collectionContactInfo, "contact-info", "", "Contact info for the collection")
	cmd.Flags().StringVar(&collectionInfoLink, "info-link", "", "Link for info about the collection")
	cmd.Flags().StringSliceVar(&collectionKeywords, "keywords", nil, "Comma separated list of keywords to help searches for the collection")
	cmd.Flags().BoolVar(&collectionPublic, "public", false, "Set the collection to be public")
	cmd.Flags().BoolVar(&collectionPrivate, "private", false, "Set the collection to be private")
	cmd.Flags().BoolVar(&collectionForceEnc, "force-encryption", false, "Force the collection to encrypt transfers")
	cmd.Flags().BoolVar(&collectionNoForceEnc, "no-force-encryption", false, "Do not force the collection to encrypt transfers")
	cmd.Flags().StringVar(&collectionDefaultDir, "default-directory", "", "Default directory when browsing or executing tasks on the collection")
	cmd.Flags().StringVar(&collectionDomainName, "domain-name", "", "DNS host name for the collection (mapped collections only)")
	cmd.Flags().StringVar(&collectionUserMessage, "user-message", "", "A message for clients to display to users when interacting with this collection")
	cmd.Flags().StringVar(&collectionUserMsgLink, "user-message-link", "", "Link to additional messaging for clients to display to users")
	cmd.Flags().StringVar(&collectionVerify, "verify", "", "File integrity verification policy: force, disable, or default")
	cmd.Flags().BoolVar(&collectionEnableHTTPS, "enable-https", false, "Explicitly enable HTTPS support")
	cmd.Flags().BoolVar(&collectionDisableHTTPS, "disable-https", false, "Explicitly disable HTTPS support")
	cmd.Flags().BoolVar(&collectionAllowGuest, "allow-guest-collections", false, "Allow guest collections on this mapped collection")
	cmd.Flags().BoolVar(&collectionNoAllowGuest, "no-allow-guest-collections", false, "Disallow guest collections on this mapped collection")
	cmd.Flags().StringVar(&collectionShareRestrict, "sharing-restrict-paths", "", "JSON path restrictions for sharing on guest collections (mapped collections only)")
	cmd.Flags().StringArrayVar(&collectionShareAllow, "sharing-user-allow", nil, "Connector-specific username allowed to create guest collections; may be given multiple times")
	cmd.Flags().StringArrayVar(&collectionShareDeny, "sharing-user-deny", nil, "Connector-specific username denied permission to create guest collections; may be given multiple times")

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

	// --limit sets the page size the user cares about; --page-size is the raw
	// GCS Manager page size. Prefer --limit when explicitly set.
	pageSize := collectionPageSize
	if cmd.Flags().Changed("limit") {
		pageSize = collectionLimit
	}

	opts := &gcs.ListCollectionsOptions{
		MappedCollectionID: collectionMappedID,
		PageSize:           pageSize,
		Filter:             collectionFilter,
	}
	if collectionInclPrivate {
		opts.Include = append(opts.Include, "private_policies")
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

	var getOpts *gcs.GetCollectionOptions
	if collectionInclPrivate {
		getOpts = &gcs.GetCollectionOptions{
			QueryParams: map[string]string{"include": "private_policies"},
		}
	}

	col, err := client.GetCollection(ctx, collectionID, getOpts)
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
	if collectionUserCredID != "" {
		doc.UserCredentialID = collectionUserCredID
	}
	if collectionIdentityID != "" {
		doc.IdentityID = collectionIdentityID
	}
	applyCollectionFlags(cmd, doc)

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
	applyCollectionFlags(cmd, doc)

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

// strPtr returns a pointer to s.
func strPtr(s string) *string { return &s }

// boolPtr returns a pointer to b.
func boolPtr(b bool) *bool { return &b }

// applyCollectionFlags sets the shared CollectionDocument fields from the flags
// the user explicitly changed. It is used by both create and update so the two
// commands accept the same field set (matching the Python CLI's shared options).
func applyCollectionFlags(cmd *cobra.Command, doc *gcs.CollectionDocument) {
	if cmd.Flags().Changed("display-name") {
		doc.DisplayName = collectionDisplayName
	}
	if cmd.Flags().Changed("description") {
		doc.Description = strPtr(collectionDescription)
	}
	if cmd.Flags().Changed("organization") {
		doc.Organization = collectionOrganization
	}
	if cmd.Flags().Changed("department") {
		doc.Department = strPtr(collectionDepartment)
	}
	if cmd.Flags().Changed("contact-email") {
		doc.ContactEmail = strPtr(collectionContactEmail)
	}
	if cmd.Flags().Changed("contact-info") {
		doc.ContactInfo = strPtr(collectionContactInfo)
	}
	if cmd.Flags().Changed("info-link") {
		doc.InfoLink = strPtr(collectionInfoLink)
	}
	if cmd.Flags().Changed("keywords") {
		doc.Keywords = collectionKeywords
	}
	if cmd.Flags().Changed("default-directory") {
		doc.DefaultDirectory = collectionDefaultDir
	}
	if cmd.Flags().Changed("domain-name") {
		doc.DomainName = collectionDomainName
	}
	if cmd.Flags().Changed("user-message") {
		doc.UserMessage = strPtr(collectionUserMessage)
	}
	if cmd.Flags().Changed("user-message-link") {
		doc.UserMessageLink = strPtr(collectionUserMsgLink)
	}

	// --public / --private
	if cmd.Flags().Changed("public") {
		doc.Public = boolPtr(collectionPublic)
	}
	if collectionPrivate {
		doc.Public = boolPtr(false)
	}

	// --force-encryption / --no-force-encryption
	if cmd.Flags().Changed("force-encryption") {
		doc.ForceEncryption = boolPtr(collectionForceEnc)
	}
	if collectionNoForceEnc {
		doc.ForceEncryption = boolPtr(false)
	}

	// --enable-https / --disable-https
	if cmd.Flags().Changed("enable-https") {
		doc.EnableHTTPS = boolPtr(collectionEnableHTTPS)
	}
	if collectionDisableHTTPS {
		doc.EnableHTTPS = boolPtr(false)
	}

	// --allow-guest-collections / --no-allow-guest-collections (mapped only)
	if cmd.Flags().Changed("allow-guest-collections") {
		doc.AllowGuestCollections = boolPtr(collectionAllowGuest)
	}
	if collectionNoAllowGuest {
		doc.AllowGuestCollections = boolPtr(false)
	}

	// --verify [force|disable|default]
	if cmd.Flags().Changed("verify") {
		switch collectionVerify {
		case "force":
			doc.ForceVerify = boolPtr(true)
			doc.DisableVerify = boolPtr(false)
		case "disable":
			doc.DisableVerify = boolPtr(true)
			doc.ForceVerify = boolPtr(false)
		case "default":
			doc.ForceVerify = boolPtr(false)
			doc.DisableVerify = boolPtr(false)
		}
	}

	// Sharing controls (mapped collections only).
	if cmd.Flags().Changed("sharing-user-allow") {
		doc.SharingUsersAllow = collectionShareAllow
	}
	if cmd.Flags().Changed("sharing-user-deny") {
		doc.SharingUsersDeny = collectionShareDeny
	}
	if cmd.Flags().Changed("sharing-restrict-paths") {
		doc.SharingRestrictPaths = json.RawMessage(collectionShareRestrict)
	}
}
