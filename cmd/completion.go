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
    --help(-h)                    # Show help information
]

export extern "tmpltr make" [
    path: string                  # Target directory path to create template from
    --name(-n): string@"nu-complete tmpltr template-names-new"  # Template name (required)
    --ignore-contents            # Only save file structure, ignore file contents
    --ignore-files: list<string> # Files/patterns to ignore (comma-separated)
    --no-compression             # Disable compression of stored files
    --help(-h)                   # Show help for make command
]

export extern "tmpltr restore" [
    --name(-n): string@"nu-complete tmpltr template-names"  # Template name to restore (required)
    --output(-o): string         # Output directory path where template will be restored (required)
    --help(-h)                   # Show help for restore command
]

export extern "tmpltr list" [
    --help(-h)                   # Show help for list command
]

export extern "tmpltr delete" [
    --name(-n): string@"nu-complete tmpltr template-names"  # Template name to delete (required)
    --force(-f)                  # Skip confirmation prompt and delete immediately
    --help(-h)                   # Show help for delete command
]

export extern "tmpltr completion" [
    shell: string@"nu-complete tmpltr completion shells"  # Shell type to generate completions for
    --help(-h)                   # Show help for completion command
]

def "nu-complete tmpltr completion shells" [] {
    [
        {value: "bash", description: "Bash shell completion"},
        {value: "zsh", description: "Zsh shell completion"},
        {value: "fish", description: "Fish shell completion"}, 
        {value: "powershell", description: "PowerShell completion"},
        {value: "nushell", description: "Nushell completion"}
    ]
}

def "nu-complete tmpltr template-names" [] {
    try {
        ^tmpltr list | lines | each { |line|
            # Look for lines that start with 📁 (template names)
            if ($line | str starts-with "📁 ") {
                let name = ($line | str replace "📁 " "" | str trim)
                {value: $name, description: "Template"}
            }
        } | where value != null
    } catch {
        []
    }
}

def "nu-complete tmpltr template-names-new" [] {
    # For make command, we don't show existing templates since we're creating new ones
    []
}

export extern "tmpltr help" [
    command?: string             # Command to get detailed help information for
    --help(-h)                   # Show help for help command
]
`
	_, err := out.WriteString(nushellScript)
	return err
}

// templateNameCompletion provides completion for template names with descriptions
func templateNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	storage, err := storage.NewStorage("")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	templates, err := storage.ListTemplates()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// For shells that support descriptions, provide template info
	var completions []string
	for _, templateName := range templates {
		// Try to load manifest to get template info for description
		if manifest, err := storage.LoadManifest(templateName); err == nil {
			fileCount := manifest.GetFileCount()
			createdAt := manifest.CreatedAt.Format("2006-01-02")
			description := fmt.Sprintf("%s\t%d files, created %s", templateName, fileCount, createdAt)
			completions = append(completions, description)
		} else {
			// Fallback to just the name if manifest can't be loaded
			completions = append(completions, templateName)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(completionCmd)
}