package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tmpltr/internal/storage"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell|nushell]",
	Short: "Generate shell completion scripts",
	Long: `Generate completion scripts for various shells to enable auto-completion
of commands, flags, and arguments.

To load completions:

Bash:
  $ source <(tmpltr completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ tmpltr completion bash > /etc/bash_completion.d/tmpltr
  # macOS:
  $ tmpltr completion bash > $(brew --prefix)/etc/bash_completion.d/tmpltr

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ tmpltr completion zsh > "${fpath[1]}/_tmpltr"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ tmpltr completion fish | source

  # To load completions for each session, execute once:
  $ tmpltr completion fish > ~/.config/fish/completions/tmpltr.fish

PowerShell:
  PS> tmpltr completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> tmpltr completion powershell > tmpltr.ps1
  # and source this file from your PowerShell profile.

Nushell:
  > tmpltr completion nushell | save tmpltr-completions.nu
  # Add the following to your Nushell config file (~/.config/nushell/config.nu):
  source tmpltr-completions.nu
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell", "nushell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:                  runCompletion,
}

// runCompletion executes the completion command
func runCompletion(cmd *cobra.Command, args []string) error {
	shell := args[0]

	switch shell {
	case "bash":
		return rootCmd.GenBashCompletion(os.Stdout)
	case "zsh":
		return rootCmd.GenZshCompletion(os.Stdout)
	case "fish":
		return rootCmd.GenFishCompletion(os.Stdout, true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
	case "nushell":
		return generateNushellCompletion(os.Stdout)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

// generateNushellCompletion generates Nushell completion script
func generateNushellCompletion(out *os.File) error {
	nushellScript := `# tmpltr completions for Nushell
# Save this file as tmpltr-completions.nu and source it in your config.nu

export extern "tmpltr" [
    --help(-h)                    # Show help
]

export extern "tmpltr make" [
    path: string                  # Target directory path
    --name(-n): string           # Template name (required)
    --ignore-contents            # Only save file structure, ignore contents
    --ignore-files: list<string> # Files/patterns to ignore
    --no-compression             # Disable compression
    --help(-h)                   # Show help
]

export extern "tmpltr restore" [
    --name(-n): string           # Template name to restore (required)
    --output(-o): string         # Output directory path (required)
    --help(-h)                   # Show help
]

export extern "tmpltr list" [
    --help(-h)                   # Show help
]

export extern "tmpltr delete" [
    --name(-n): string           # Template name to delete (required)
    --force(-f)                  # Skip confirmation prompt
    --help(-h)                   # Show help
]

export extern "tmpltr completion" [
    shell: string@"nu-complete tmpltr completion shells"  # Shell type
    --help(-h)                   # Show help
]

def "nu-complete tmpltr completion shells" [] {
    ["bash", "zsh", "fish", "powershell", "nushell"]
}

export extern "tmpltr help" [
    command?: string             # Command to get help for
    --help(-h)                   # Show help
]
`
	_, err := out.WriteString(nushellScript)
	return err
}

// templateNameCompletion provides completion for template names
func templateNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	storage, err := storage.NewStorage("")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	templates, err := storage.ListTemplates()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return templates, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(completionCmd)
}