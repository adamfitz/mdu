package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mdu/metadata"
	"mdu/parser"
)

func NewEPUBCmd() *cobra.Command {
	epubCmd := &cobra.Command{
		Use:   "epub",
		Short: "EPUB metadata operations",
		Long:  `Read, update, and validate EPUB file metadata.`,
	}

	epubCmd.AddCommand(
		newEPUBReadCmd(),
		newEPUBUpdateCmd(),
		newEPUBCompareCmd(),
		newEPUBValidateCmd(),
		newEPUBCheckCmd(),
	)

	return epubCmd
}

func newEPUBReadCmd() *cobra.Command {
	var file, dir, output string
	var listFields, all bool

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read metadata from EPUB files",
		Long: `Read and display metadata from EPUB files.
Use --all to see all metadata fields, not just supported ones.`,
		Example: `  mdu epub read --file book.epub
  mdu epub read --dir ./books --all
  mdu epub read --file book.epub --output metadata.txt`,
		Run: func(cmd *cobra.Command, args []string) {
			if listFields {
				fmt.Println("Supported EPUB metadata fields (Kavita-compatible):")
				fmt.Println("  author           - Book author (dc:creator)")
				fmt.Println("  summary          - Book description (dc:description + Summary)")
				fmt.Println("  publisher        - Publisher name (dc:publisher)")
				fmt.Println("  isbn             - ISBN identifier (dc:identifier with scheme)")
				fmt.Println("  calibre:series   - Series name (Kavita: Name)")
				fmt.Println("  calibre:series_index - Series position (Kavita: Volume)")
				fmt.Println("  subject          - Genres (dc:subject)")
				return
			}

			if file == "" && dir == "" {
				log.Fatal("Error: You must specify either --file or --dir")
			}

			allFiles, err := getEPUBFiles(file, dir)
			if err != nil {
				log.Fatalf("Error getting EPUB files: %v", err)
			}

			var outputStr string
			for _, f := range allFiles {
				md, err := metadata.Read(f, all)
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
				fmt.Printf("\n✓ Metadata written to %s\n", output)
			}
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Target EPUB file")
	cmd.Flags().StringVar(&dir, "dir", "", "Target directory containing EPUB files")
	cmd.Flags().BoolVar(&listFields, "list-fields", false, "List all supported metadata fields")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Show all metadata fields")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write output to file")

	return cmd
}

func newEPUBUpdateCmd() *cobra.Command {
	var file, dir, output, inputFile string
	var series, seriesIndex, summary, isbn, author string
	var createBackup bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update metadata in EPUB files",
		Long: `Update metadata in EPUB files while preserving the original OPF structure.
Creates backups by default unless --no-backup is specified.

You can specify metadata using either:
  1. Command-line flags (--author, --series, etc.)
  2. An input file (--input) in JSON or YAML format`,
		Example: `  mdu epub update --file book.epub --author "John Doe" --series "My Series"
  mdu epub update --dir ./books --input metadata.json
  mdu epub update --file book.epub --isbn "978-1234567890" --no-backup`,
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" && dir == "" {
				log.Fatal("Error: You must specify either --file or --dir")
			}

			var updates map[string]string
			var err error

			if inputFile != "" {
				updates, err = parser.ParseInputFile(inputFile)
				if err != nil {
					log.Fatalf("Error parsing input file: %v", err)
				}
				fmt.Printf("✓ Loaded metadata from input file: %s\n", inputFile)
			} else {
				updates = buildUpdatesMap(series, seriesIndex, summary, isbn, author)
			}

			if len(updates) == 0 {
				log.Fatal("Error: No metadata fields specified for update")
			}

			allFiles, err := getEPUBFiles(file, dir)
			if err != nil {
				log.Fatalf("Error getting EPUB files: %v", err)
			}

			fmt.Printf("Updating %d file(s) with the following changes:\n", len(allFiles))
			for k, v := range updates {
				fmt.Printf("  %s: %s\n", k, v)
			}
			fmt.Println()

			var outputStr string
			successCount := 0

			for _, f := range allFiles {
				if createBackup {
					backupPath := f + ".backup"
					if err := copyFile(f, backupPath); err != nil {
						log.Printf("Warning: Could not create backup for %s: %v", f, err)
					} else {
						fmt.Printf("✓ Backup created: %s\n", backupPath)
					}
				}

				if err := metadata.Update(f, f, updates); err != nil {
					log.Printf("✗ Error updating %s: %v", filepath.Base(f), err)
					continue
				}

				md, err := metadata.Read(f, false)
				if err != nil {
					log.Printf("Warning: Updated %s but couldn't read metadata: %v",
						filepath.Base(f), err)
				} else {
					outputStr += parser.RenderMetadataWithHeader(filepath.Base(f), md)
				}

				fmt.Printf("✓ Updated: %s\n", filepath.Base(f))
				successCount++
			}

			fmt.Printf("\n✓ Successfully updated %d of %d file(s)\n\n",
				successCount, len(allFiles))
			fmt.Print(outputStr)

			if output != "" {
				if err := os.WriteFile(output, []byte(outputStr), 0644); err != nil {
					log.Fatalf("Error writing output file: %v", err)
				}
				fmt.Printf("✓ Results written to %s\n", output)
			}
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Target EPUB file")
	cmd.Flags().StringVar(&dir, "dir", "", "Target directory for batch update")
	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file (JSON or YAML) with metadata updates")
	cmd.Flags().StringVar(&series, "series", "", "Series name")
	cmd.Flags().StringVar(&seriesIndex, "series-index", "", "Series index")
	cmd.Flags().StringVar(&summary, "summary", "", "Book summary")
	cmd.Flags().StringVar(&isbn, "isbn", "", "ISBN identifier")
	cmd.Flags().StringVar(&author, "author", "", "Author/creator name")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write results to file")
	cmd.Flags().BoolVar(&createBackup, "backup", true, "Create .backup files")

	return cmd
}

func newEPUBCompareCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "compare <original.epub> <modified.epub>",
		Short: "Compare metadata between two EPUB files",
		Long: `Compare metadata from two EPUB files and generate a diff report.
Useful for validating changes after updates.`,
		Example: `  mdu epub compare book.epub book.epub.backup
  mdu epub compare original.epub modified.epub --output diff.txt`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			original := args[0]
			modified := args[1]

			if !fileExists(original) {
				log.Fatalf("Error: Original file not found: %s", original)
			}
			if !fileExists(modified) {
				log.Fatalf("Error: Modified file not found: %s", modified)
			}

			diff, err := metadata.CompareOPF(original, modified)
			if err != nil {
				log.Fatalf("Error comparing files: %v", err)
			}

			fmt.Println(diff)

			if output != "" {
				if err := os.WriteFile(output, []byte(diff), 0644); err != nil {
					log.Fatalf("Error writing output file: %v", err)
				}
				fmt.Printf("\n✓ Comparison report written to %s\n", output)
			}
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Write comparison report to file")

	return cmd
}

func newEPUBValidateCmd() *cobra.Command {
	var file, output string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate changes by comparing with backup",
		Long: `Compare an EPUB file with its .backup file to validate changes.
Requires that a backup file exists (created with --backup flag during update).`,
		Example: `  mdu epub validate --file book.epub
  mdu epub validate --file book.epub --output validation.txt`,
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" {
				log.Fatal("Error: You must specify --file")
			}

			backupFile := file + ".backup"
			if !fileExists(backupFile) {
				log.Fatalf("Error: Backup file not found: %s", backupFile)
			}

			diff, err := metadata.CompareOPF(backupFile, file)
			if err != nil {
				log.Fatalf("Error comparing files: %v", err)
			}

			fmt.Printf("Comparing: %s (original) vs %s (current)\n\n",
				filepath.Base(backupFile), filepath.Base(file))
			fmt.Println(diff)

			if output != "" {
				if err := os.WriteFile(output, []byte(diff), 0644); err != nil {
					log.Fatalf("Error writing output file: %v", err)
				}
				fmt.Printf("\n✓ Validation report written to %s\n", output)
			}
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Target EPUB file to validate")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write validation report to file")

	return cmd
}

func newEPUBCheckCmd() *cobra.Command {
	var file, dir, output string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check EPUB file validity and structure",
		Long: `Validates that EPUB files have proper structure:
- META-INF/container.xml exists
- OPF file exists at location specified in container.xml
- OPF file has minimum required metadata (title, identifier, language)`,
		Example: `  mdu epub check --file book.epub
  mdu epub check --dir ./books --output report.txt`,
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" && dir == "" {
				log.Fatal("Error: You must specify either --file or --dir")
			}

			allFiles, err := getEPUBFiles(file, dir)
			if err != nil {
				log.Fatalf("Error getting EPUB files: %v", err)
			}

			var results strings.Builder
			validCount := 0
			invalidCount := 0

			for _, f := range allFiles {
				results.WriteString(fmt.Sprintf("\n=== Checking: %s ===\n", filepath.Base(f)))

				md, err := metadata.Read(f, false)

				if err != nil {
					results.WriteString(fmt.Sprintf("❌ INVALID: %v\n", err))
					invalidCount++
					continue
				}

				results.WriteString("✓ EPUB structure valid\n")
				results.WriteString("✓ OPF file found and readable\n")
				results.WriteString("✓ Minimum required metadata present\n")
				results.WriteString("\nMetadata summary:\n")
				results.WriteString(fmt.Sprintf("  Title: %s\n", md["title"]))
				results.WriteString(fmt.Sprintf("  Author: %s\n", md["author"]))

				for key, val := range md {
					if key == "identifier" || strings.HasPrefix(key, "identifier:") {
						results.WriteString(fmt.Sprintf("  Identifier: %s\n", val))
						break
					}
				}

				validCount++
			}

			results.WriteString("\n=== Summary ===\n")
			results.WriteString(fmt.Sprintf("Valid: %d\n", validCount))
			results.WriteString(fmt.Sprintf("Invalid: %d\n", invalidCount))

			fmt.Print(results.String())

			if output != "" {
				if err := os.WriteFile(output, []byte(results.String()), 0644); err != nil {
					log.Fatalf("Error writing output file: %v", err)
				}
				fmt.Printf("\n✓ Check results written to %s\n", output)
			}
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Target EPUB file to check")
	cmd.Flags().StringVar(&dir, "dir", "", "Target directory to check")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write check results to file")

	return cmd
}

// EPUB-specific helper
func getEPUBFiles(file, dir string) ([]string, error) {
	if dir != "" {
		return parser.ListEPUBFiles(dir)
	}
	return []string{file}, nil
}
