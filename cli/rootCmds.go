package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mdu",
		Short: "Metadata reader and updater for various formats",
		Long: `A tool for reading and updating metadata in multiple formats.
Currently supports:
  - EPUB files (complete support)
  - PDF files (coming soon)
  - ComicInfo.xml (coming soon)

Preserves original file structure and formatting.`,
	}

	// Add format-specific subcommands
	rootCmd.AddCommand(NewEPUBCmd())
	// rootCmd.AddCommand(NewPDFCmd())      // Future
	rootCmd.AddCommand(NewComicInfoCmd())

	// Add general commands
	rootCmd.AddCommand(NewGenerateCmd())

	return rootCmd
}
