package cli

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"mdu/metadata"
)

func NewRootCmd() *cobra.Command {
	var file string
	var listFields bool

	rootCmd := &cobra.Command{
		Use:   "mdu",
		Short: "EPUB metadata reader and updater",
	}

	// --- Read command ---
	readCmd := &cobra.Command{
		Use:   "read",
		Short: "Read metadata from an EPUB file",
		Run: func(cmd *cobra.Command, args []string) {
			if listFields {
				fmt.Println("Supported metadata fields:")
				for _, f := range metadata.SupportedFields {
					fmt.Printf("  %s\n", f)
				}
				return
			}

			if file == "" {
				log.Fatal("You must specify --file")
			}
			md, err := metadata.Read(file)
			if err != nil {
				log.Fatalf("Error reading metadata: %v", err)
			}
			for k, v := range md {
				fmt.Printf("%s: %s\n", k, v)
			}
		},
	}
	readCmd.Flags().StringVar(&file, "file", "", "Target EPUB file")
	readCmd.Flags().BoolVar(&listFields, "list-fields", false, "List all supported metadata fields")

	// --- Update command ---
	var series, seriesIndex, summary, isbn string

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update metadata fields in an EPUB file",
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" {
				log.Fatal("You must specify --file")
			}

			updates := make(map[string]string)
			if series != "" {
				updates["calibre:series"] = series
			}
			if seriesIndex != "" {
				updates["calibre:series_index"] = seriesIndex
			}
			if summary != "" {
				updates["summary"] = summary
			}
			if isbn != "" {
				updates["isbn"] = isbn
			}

			if len(updates) == 0 {
				log.Fatal("No metadata fields specified for update")
			}

			if err := metadata.Update(file, file, updates); err != nil {
				log.Fatalf("Error updating EPUB: %v", err)
			}
			fmt.Println("Metadata updated successfully")
		},
	}

	updateCmd.Flags().StringVar(&file, "file", "", "Target EPUB file")
	updateCmd.Flags().StringVar(&series, "series", "", "Series name")
	updateCmd.Flags().StringVar(&seriesIndex, "series-index", "", "Series index")
	updateCmd.Flags().StringVar(&summary, "summary", "", "Book summary")
	updateCmd.Flags().StringVar(&isbn, "isbn", "", "ISBN identifier")

	rootCmd.AddCommand(readCmd, updateCmd)
	return rootCmd
}
