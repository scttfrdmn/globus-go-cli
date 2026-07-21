// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/scttfrdmn/globus-go-cli/cmd/collection"
)

// getCollectionCommand returns the `collection` command group (GCSv5 collection
// management).
func getCollectionCommand() *cobra.Command {
	return collection.CollectionCmd()
}

// getGCSCommand returns the `gcs` command group (GCSv5 endpoint/storage-gateway/
// role management).
func getGCSCommand() *cobra.Command {
	return collection.GCSCmd()
}
