// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package collection

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/gcs"
)

// collectionCatCmd returns the `collection cat` command: read a file's contents
// from an HTTPS-enabled GCSv5 collection's data plane.
func collectionCatCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cat ENDPOINT_ID COLLECTION_ID PATH",
		Short: "Print a file's contents from an HTTPS-enabled collection",
		Long: `Read and print a file from a Globus Connect Server v5 collection over its
HTTPS data-plane interface.

The collection must have HTTPS enabled (an https_url). This is a data-plane
operation: it requires the collection's data-access consent, which the CLI
escalates on first use (separately from the endpoint's manage_collections
consent used by the other collection commands).

Examples:
  globus collection cat ENDPOINT_ID COLLECTION_ID /path/to/file.txt`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return catCollectionFile(cmd, args[0], args[1], args[2])
		},
	}
}

// catCollectionFile reads COLLECTION_ID:PATH over HTTPS and writes it to stdout.
func catCollectionFile(cmd *cobra.Command, endpointID, collectionID, path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Resolve the collection's HTTPS data-plane base URL via the GCS Manager.
	managerClient, err := getManagerClient(ctx, endpointID)
	if err != nil {
		return err
	}
	coll, err := managerClient.GetCollection(ctx, collectionID, nil)
	if err != nil {
		return fmt.Errorf("failed to look up collection %s: %w", collectionID, err)
	}
	if coll.HTTPSURL == "" {
		return fmt.Errorf("collection %s does not have HTTPS enabled (no https_url); cannot read files over HTTPS", collectionID)
	}

	// Obtain a data-access (https) token for this collection.
	token, err := collectionHTTPSToken(ctx, collectionID)
	if err != nil {
		return err
	}

	// Build the file URI: <https_url>/<path> (single slash join).
	fileURI := strings.TrimRight(coll.HTTPSURL, "/") + "/" + strings.TrimPrefix(path, "/")

	// A CollectionClient is required to construct the Downloader; the raw token
	// passed to NewDownloaderWithToken is what actually authorizes the
	// data-plane request.
	dlClient, err := gcs.NewCollectionClient(ctx, coll.HTTPSURL, collectionID,
		&core.Config{Authorizer: authorizers.NewAccessTokenAuthorizer(token)})
	if err != nil {
		return fmt.Errorf("failed to create data-plane client: %w", err)
	}
	downloader := gcs.NewDownloaderWithToken(dlClient, token)
	defer func() { _ = downloader.Close() }()

	data, err := downloader.ReadFile(ctx, fileURI)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", fileURI, err)
	}

	_, err = cmd.OutOrStdout().Write(data)
	return err
}
