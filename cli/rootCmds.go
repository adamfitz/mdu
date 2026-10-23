package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mdu",
		Short: "Metadata reader and updater for various formats",
		Long: `A tool for reading and updating metadata in multiple formats.
Currently supports:
  - EPUB files (complete support)
  - PDF files (coming soon)
  - ComicInfo.xml

Preserves original file structure and formatting.`,
	}

	// Add format-specific subcommands
	rootCmd.AddCommand(NewEPUBCmd())
	// rootCmd.AddCommand(NewPDFCmd())      // Future
	rootCmd.AddCommand(NewComicInfoCmd())
	rootCmd.AddCommand(NewMangadexCmd())

	// Add general commands
	rootCmd.AddCommand(NewGenerateCmd())

	if version != "" {
		rootCmd.Version = version
		rootCmd.SetVersionTemplate("mdu {{.Version}}\n")
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			if version != "" {
				fmt.Println(version)
			} else {
				fmt.Println(rootCmd.Version)
			}
		},
	}
	rootCmd.AddCommand(versionCmd)

	return rootCmd
}
