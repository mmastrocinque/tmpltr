package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tmpltr",
	Short: "A CLI tool for creating and managing directory templates",
	Long: `tmpltr is a command-line tool that enables users to create, save, and restore 
directory templates. It supports saving files with hashed identifiers for deduplication, 
optionally ignoring file contents, and preserving relative paths.

Key features:
  • Create templates from existing directories
  • Restore templates to new locations
  • Hash-based file deduplication
  • Optional content-only or structure-only modes
  • List and manage saved templates

Examples:
  tmpltr make ./my-project --name="my-template"
  tmpltr restore --name="my-template" --output="./new-project"
  tmpltr list
  tmpltr delete --name="my-template"`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add all subcommands to root command
	rootCmd.AddCommand(makeCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(deleteCmd)

	// Global flags can be added here if needed
	// rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}