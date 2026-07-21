// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

// getListCommandsCommand returns the list-commands command, which prints the
// full command tree (every command's full path plus its short description),
// mirroring the Python Globus CLI's `globus list-commands`.
//
// Output is a deterministic, sorted, flat list of full command paths (e.g.
// "globus endpoint role create") so it is easy to diff and grep. No network is
// used; it only walks the in-memory cobra command tree.
func getListCommandsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-commands",
		Short: "List all commands available in the CLI",
		Long: `List every command available in the CLI as a flat, sorted list of
full command paths, each with its short description.

This mirrors the Python Globus CLI's "globus list-commands" and is useful for
discovering the command surface and for scripting/diffing.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()

			type entry struct {
				path  string
				short string
			}
			var entries []entry

			var walk func(c *cobra.Command)
			walk = func(c *cobra.Command) {
				// Skip the auto-generated "help" command (and its children):
				// it is noisy and not part of the real command surface.
				if c.Name() == "help" {
					return
				}
				// Record every command except the root itself.
				if c.Parent() != nil {
					entries = append(entries, entry{path: c.CommandPath(), short: c.Short})
				}
				children := c.Commands()
				sort.Slice(children, func(i, j int) bool {
					return children[i].Name() < children[j].Name()
				})
				for _, child := range children {
					walk(child)
				}
			}
			walk(root)

			sort.Slice(entries, func(i, j int) bool {
				return entries[i].path < entries[j].path
			})

			out := cmd.OutOrStdout()
			for _, e := range entries {
				if e.short != "" {
					fmt.Fprintf(out, "%s\t%s\n", e.path, e.short)
				} else {
					fmt.Fprintln(out, e.path)
				}
			}
			return nil
		},
	}
}
