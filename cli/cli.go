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
	var file, dir, output string
	var listFields, all bool

	var series, seriesIndex, summary, isbn, author string

	rootCmd := &cobra.Command{
		Use:   "mdu",
		Short: "Metadata reader and updater for EPUB and PDF files",
	}

	// --- Read command ---
	readCmd := &cobra.Command{
		Use:   "read",
		Short: "Read metadata from EPUB or PDF file(s)",
		Run: func(cmd *cobra.Command, args []string) {
			if listFields {
				fmt.Println("Supported metadata fields:")
				for _, f := range metadata.SupportedFields {
					fmt.Printf("  %s\n", f)
				}
				for _, f := range metadata.PdfSupportedFields {
					fmt.Printf("  %s\n", f)
				}
				return
			}

			if file == "" && dir == "" {
				log.Fatal("You must specify either --file or --dir")
			}

			var allFiles []string
			if dir != "" {
				epubs, _ := parser.ListEpubFiles(dir)
				pdfs, _ := parser.ListPdfFiles(dir)
				allFiles = append(epubs, pdfs...) // need to expand the second slice to append each element in the slice
			} else {
				allFiles = []string{file}
			}

			var outputStr string
			for _, f := range allFiles {
				var md map[string]string
				var err error

				if ok, _ := parser.IsEpub(f); ok {
					md, err = metadata.Read(f, all)
				} else if ok, _ := parser.IsPDF(f); ok {
					md, err = metadata.PdfRead(f, all)
				} else {
					log.Printf("Unsupported file format: %s", f)
					continue
				}

				if err != nil {
					log.Printf("Error reading %s: %v", f, err)
					continue
				}
				outputStr += parser.RenderMetadataWithHeader(filepath.Base(f), md)
			}

			fmt.Print(outputStr)

			if output != "" {
				if err := os.WriteFile(output, []byte(outputStr), 0644); err != nil {
					log.Fatalf("Error writing output file: %v", err)
				}
				fmt.Printf("Metadata written to %s\n", output)
			}
		},
	}

	readCmd.Flags().StringVar(&file, "file", "", "Target EPUB or PDF file")
	readCmd.Flags().StringVar(&dir, "dir", "", "Target directory containing EPUB or PDF files")
	readCmd.Flags().BoolVar(&listFields, "list-fields", false, "List all supported metadata fields")
	readCmd.Flags().BoolVarP(&all, "all", "a", false, "Return all metadata fields, not just known ones")
	readCmd.Flags().StringVarP(&output, "output", "o", "", "Write output to file instead of printing to console")

	// --- Update command ---
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update metadata fields in EPUB or PDF file(s)",
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
				epubs, _ := parser.ListEpubFiles(dir)
				pdfs, _ := parser.ListPdfFiles(dir)
				allFiles = append(epubs, pdfs...) // need to expand the second slice to append each element in the slice
			} else {
				allFiles = []string{file}
			}

			var outputStr string
			for _, f := range allFiles {
				var err error
				if ok, _ := parser.IsEpub(f); ok {
					err = metadata.Update(f, f, updates)
				} else if ok, _ := parser.IsPDF(f); ok {
					err = metadata.PdfUpdate(f, f, updates)
				} else {
					log.Printf("Unsupported file format: %s", f)
					continue
				}

				if err != nil {
					log.Printf("Error updating %s: %v", f, err)
					continue
				}

				var md map[string]string
				if ok, _ := parser.IsEpub(f); ok {
					md, _ = metadata.Read(f, false)
				} else {
					md, _ = metadata.PdfRead(f, false)
				}
				outputStr += parser.RenderMetadataWithHeader(filepath.Base(f), md)
			}

			fmt.Print(outputStr)

			if output != "" {
				if err := os.WriteFile(output, []byte(outputStr), 0644); err != nil {
					log.Fatalf("Error writing to output file: %v", err)
				}
				fmt.Printf("Metadata written to %s\n", output)
			}
		},
	}

	updateCmd.Flags().StringVar(&file, "file", "", "Target EPUB or PDF file")
	updateCmd.Flags().StringVar(&dir, "dir", "", "Target directory containing EPUB or PDF files")
	updateCmd.Flags().StringVar(&series, "series", "", "Series name")
	updateCmd.Flags().StringVar(&seriesIndex, "series-index", "", "Series index")
	updateCmd.Flags().StringVar(&summary, "summary", "", "Book summary")
	updateCmd.Flags().StringVar(&isbn, "isbn", "", "ISBN identifier")
	updateCmd.Flags().StringVar(&author, "author", "", "Author/creator name")
	updateCmd.Flags().StringVarP(&output, "output", "o", "", "Optional output file to write update results")

	rootCmd.AddCommand(readCmd, updateCmd)
	return rootCmd
}
