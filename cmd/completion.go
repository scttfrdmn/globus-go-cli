// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for Globus CLI.

To load completions:

Bash:
  $ source <(globus completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ globus completion bash > /etc/bash_completion.d/globus
  # macOS:
  $ globus completion bash > $(brew --prefix)/etc/bash_completion.d/globus

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ globus completion zsh > "${fpath[1]}/_globus"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ globus completion fish | source

  # To load completions for each session, execute once:
  $ globus completion fish > ~/.config/fish/completions/globus.fish

PowerShell:
  PS> globus completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> globus completion powershell > globus.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}