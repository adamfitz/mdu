package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"mdu/metadata"
	"mdu/parser"
)

func NewRootCmd() *cobra.Command {
	var file string
	var listFields bool
	var all bool
	var output string

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

			md, err := metadata.Read(file, all)
			if err != nil {
				log.Fatalf("Error reading metadata\n%v\n", err)
			}

			rendered := parser.RenderMetadataOutput(md)
			fmt.Printf("\n%s", rendered)

			if output != "" {
				if err := os.WriteFile(output, []byte(rendered), 0644); err != nil {
					log.Fatalf("Error writing output file: %v", err)
				}
				fmt.Printf("\nMetadata written to %s\n", output)
			}
		},
	}

	readCmd.Flags().StringVar(&file, "file", "", "Target EPUB file")
	readCmd.Flags().BoolVar(&listFields, "list-fields", false, "List all supported metadata fields")
	readCmd.Flags().BoolVarP(&all, "all", "a", false, "Return all metadata fields, not just known ones")
	readCmd.Flags().StringVarP(&output, "output", "o", "", "Write output to file instead of printing to console")

	// --- Update command ---
	var series, seriesIndex, summary, isbn, author, dir string

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update metadata fields in an EPUB file",
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" && dir == "" {
				log.Fatal("You must specify either --file or --dir")
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
			if author != "" {
				updates["author"] = author
			}

			if len(updates) == 0 {
				log.Fatal("No metadata fields specified for update")
			}

			var allFiles []string
			if dir != "" {
				epubs, err := parser.ListEPUBFiles(dir)
				if err != nil {
					log.Fatalf("Error reading directory: %v", err)
				}
				allFiles = epubs
			} else {
				allFiles = []string{file}
			}

			for _, f := range allFiles {
				if err := metadata.Update(f, f, updates); err != nil {
					log.Printf("Error updating %s: %v", f, err)
					continue
				}
				md, _ := metadata.Read(f, false) // read updated fields (known only)
				outputStr := parser.RenderMetadataWithHeader(filepath.Base(f), md)
				fmt.Print(outputStr)
			}

			if len(allFiles) > 1 {
				fmt.Printf("Updated %d EPUB files in directory.\n", len(allFiles))
			} else {
				fmt.Println("Metadata updated successfully")
			}
		},
	}

	updateCmd.Flags().StringVar(&file, "file", "", "Target EPUB file")
	updateCmd.Flags().StringVar(&series, "series", "", "Series name")
	updateCmd.Flags().StringVar(&seriesIndex, "series-index", "", "Series index")
	updateCmd.Flags().StringVar(&summary, "summary", "", "Book summary")
	updateCmd.Flags().StringVar(&isbn, "isbn", "", "ISBN identifier")
	updateCmd.Flags().StringVar(&author, "author", "", "Author/creator name")
	updateCmd.Flags().StringVar(&dir, "dir", "", "Target directory containing EPUB files for batch update")

	rootCmd.AddCommand(readCmd, updateCmd)
	return rootCmd
}
