// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package cmd

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// collectCommandPaths walks the full command tree rooted at root and returns
// the set of full command paths (space-joined, relative to the root, e.g.
// "endpoint role create"). The root itself and the auto-generated "help"
// command are excluded.
func collectCommandPaths(root *cobra.Command) map[string]bool {
	paths := map[string]bool{}
	rootName := root.Name()

	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Name() == "help" {
			return
		}
		if c.Parent() != nil {
			// CommandPath() is e.g. "globus endpoint role create"; trim the
			// leading root name so callers compare against "endpoint ...".
			path := strings.TrimPrefix(c.CommandPath(), rootName+" ")
			paths[path] = true
		}
		for _, child := range c.Commands() {
			walk(child)
		}
	}
	walk(root)
	return paths
}

// TestCommandParity codifies the drop-in command surface that mirrors the
// Python Globus CLI. It asserts that a curated list of expected command paths
// are present (flat top-level auth/transfer, plus grouped services), and that
// legacy nested paths are absent. No network is used.
func TestCommandParity(t *testing.T) {
	root := Execute()
	have := collectCommandPaths(root)

	// Every path here has been verified to exist on the wired command tree.
	// Do NOT add paths that are being introduced concurrently (e.g.
	// endpoint-manager, endpoint set-subscription-id, my-shared-endpoint-list),
	// to keep this test deterministic.
	expected := []string{
		// Flat auth commands.
		"login", "logout", "whoami", "get-identities",
		// Flat transfer file operations.
		"ls", "mkdir", "rm", "rename", "stat", "delete", "transfer",
		// Transfer task operations.
		"task show", "task list", "task cancel", "task wait",
		"task event-list", "task pause-info", "task update",
		// Endpoint operations.
		"endpoint list", "endpoint show", "endpoint search",
		"endpoint update", "endpoint delete",
		"endpoint set-subscription-id", "endpoint my-shared-endpoint-list",
		"endpoint role list", "endpoint role create",
		"endpoint permission list", "endpoint permission create",
		// Endpoint-manager admin surface.
		"endpoint-manager monitored-endpoints", "endpoint-manager task-list",
		"endpoint-manager pause-rule list",
		// Bookmarks.
		"bookmark list", "bookmark create", "bookmark delete",
		// Streams (Go extensions).
		"tunnel list", "tunnel create", "stream-access-point list",
		// Groups.
		"group list", "group create", "group member add",
		"group member accept", "group policies show",
		// Other grouped services.
		"search query", "flows list", "timer list",
		// Search index roles + task list.
		"search index role list", "search index role create", "search index role delete",
		"search task list",
		// Flows run management + validation.
		"flows run delete", "flows run resume", "flows validate",
		// Raw API passthrough.
		"api transfer", "api auth",
		// Session.
		"session show", "session update", "session consent",
		// GCSv5 collections + manager admin.
		"collection list", "collection show", "collection create", "collection delete",
		"gcs info", "gcs storage-gateway list", "gcs role list", "gcs role create",
		// Project / client / credential management (globus-project-manager port).
		"project list", "project create", "project admin add",
		"project client list", "project client create",
		"project credential list", "project credential rotate", "project credential process-deletions",
		// Meta commands.
		"list-commands", "version",
	}

	var missing []string
	for _, path := range expected {
		if !have[path] {
			missing = append(missing, path)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Errorf("missing expected command paths (%d): %s", len(missing), strings.Join(missing, ", "))
	}

	// Legacy nested paths must NOT exist: the tree is flat now.
	forbidden := []string{
		"auth login", "auth whoami",
		"transfer ls", "transfer cp",
	}
	var unexpected []string
	for _, path := range forbidden {
		if have[path] {
			unexpected = append(unexpected, path)
		}
	}
	if len(unexpected) > 0 {
		sort.Strings(unexpected)
		t.Errorf("found legacy nested command paths that should be flattened (%d): %s", len(unexpected), strings.Join(unexpected, ", "))
	}
}
