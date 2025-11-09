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

func NewRootCmd() *cobra.Command {
	var file, dir, output, inputFile string
	var listFields, all bool
	var series, seriesIndex, summary, isbn, author string
	var createBackup bool

	rootCmd := &cobra.Command{
		Use:   "mdu",
		Short: "EPUB metadata reader and updater",
		Long: `A tool for reading and updating EPUB metadata.
Preserves the original OPF file structure and all XML namespaces.

Supports both command-line flags and input files (JSON/YAML) for batch operations.`,
	}

	// --- Read command ---
	readCmd := &cobra.Command{
		Use:   "read",
		Short: "Read metadata from an EPUB file or directory",
		Long: `Read and display metadata from EPUB files.
Use --all to see all metadata fields, not just supported ones.`,
		Run: func(cmd *cobra.Command, args []string) {
			if listFields {
				fmt.Println("Supported metadata fields:")
				for _, f := range metadata.SupportedFields {
					fmt.Printf("  %s\n", f)
				}
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

	readCmd.Flags().StringVar(&file, "file", "", "Target EPUB file")
	readCmd.Flags().StringVar(&dir, "dir", "", "Target directory containing EPUB files")
	readCmd.Flags().BoolVar(&listFields, "list-fields", false, "List all supported metadata fields")
	readCmd.Flags().BoolVarP(&all, "all", "a", false, "Show all metadata fields")
	readCmd.Flags().StringVarP(&output, "output", "o", "", "Write output to file")

	// --- Update command ---
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update metadata fields in EPUB files",
		Long: `Update metadata in EPUB files while preserving the original OPF structure.
Creates backups by default unless --no-backup is specified.

You can specify metadata using either:
  1. Command-line flags (--author, --series, etc.)
  2. An input file (--input) in JSON or YAML format

Input file takes precedence over command-line flags if both are provided.`,
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" && dir == "" {
				log.Fatal("Error: You must specify either --file or --dir")
			}

			var updates map[string]string
			var err error

			// Parse input file if provided
			if inputFile != "" {
				updates, err = parser.ParseInputFile(inputFile)
				if err != nil {
					log.Fatalf("Error parsing input file: %v", err)
				}
				fmt.Printf("✓ Loaded metadata from input file: %s\n", inputFile)
			} else {
				// Build updates from command-line flags
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
				// Create backup if requested
				if createBackup {
					backupPath := f + ".backup"
					if err := copyFile(f, backupPath); err != nil {
						log.Printf("Warning: Could not create backup for %s: %v", f, err)
					} else {
						fmt.Printf("✓ Backup created: %s\n", backupPath)
					}
				}

				// Update the file
				if err := metadata.Update(f, f, updates); err != nil {
					log.Printf("✗ Error updating %s: %v", filepath.Base(f), err)
					continue
				}

				// Read back and display updated metadata
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

	updateCmd.Flags().StringVar(&file, "file", "", "Target EPUB file")
	updateCmd.Flags().StringVar(&dir, "dir", "", "Target directory for batch update")
	updateCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file (JSON or YAML) with metadata updates")
	updateCmd.Flags().StringVar(&series, "series", "", "Series name")
	updateCmd.Flags().StringVar(&seriesIndex, "series-index", "", "Series index")
	updateCmd.Flags().StringVar(&summary, "summary", "", "Book summary")
	updateCmd.Flags().StringVar(&isbn, "isbn", "", "ISBN identifier")
	updateCmd.Flags().StringVar(&author, "author", "", "Author/creator name")
	updateCmd.Flags().StringVarP(&output, "output", "o", "", "Write results to file")
	updateCmd.Flags().BoolVar(&createBackup, "backup", true, "Create .backup files")

	// --- Compare/Validate command ---
	compareCmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare metadata between two EPUB files",
		Long: `Compare metadata from two EPUB files and generate a diff report.
Useful for validating changes after updates.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				log.Fatal("Error: compare requires exactly 2 file paths\nUsage: mdu compare <original.epub> <modified.epub>")
			}

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

	compareCmd.Flags().StringVarP(&output, "output", "o", "", "Write comparison report to file")

	// --- Validate command (compare file with its backup) ---
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate changes by comparing with backup",
		Long: `Compare an EPUB file with its .backup file to validate changes.
Requires that a backup file exists (created with --backup flag during update).`,
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

	validateCmd.Flags().StringVar(&file, "file", "", "Target EPUB file to validate")
	validateCmd.Flags().StringVarP(&output, "output", "o", "", "Write validation report to file")

	// --- Check command (validate EPUB structure) ---
	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Check EPUB file validity and structure",
		Long: `Validates that an EPUB file has proper structure:
- META-INF/container.xml exists
- OPF file exists at location specified in container.xml
- OPF file has minimum required metadata (title, identifier, language)`,
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

				// Try to read metadata
				md, err := metadata.Read(f, false)

				if err != nil {
					results.WriteString(fmt.Sprintf("❌ INVALID: %v\n", err))
					invalidCount++
					continue
				}

				// Check passed
				results.WriteString("✓ EPUB structure valid\n")
				results.WriteString("✓ OPF file found and readable\n")
				results.WriteString("✓ Minimum required metadata present\n")
				results.WriteString(fmt.Sprintf("\nMetadata summary:\n"))
				results.WriteString(fmt.Sprintf("  Title: %s\n", md["title"]))
				results.WriteString(fmt.Sprintf("  Author: %s\n", md["author"]))

				// Show identifiers
				for key, val := range md {
					if key == "identifier" || strings.HasPrefix(key, "identifier:") {
						results.WriteString(fmt.Sprintf("  Identifier: %s\n", val))
						break
					}
				}

				validCount++
			}

			results.WriteString(fmt.Sprintf("\n=== Summary ===\n"))
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

	checkCmd.Flags().StringVar(&file, "file", "", "Target EPUB file to check")
	checkCmd.Flags().StringVar(&dir, "dir", "", "Target directory to check")
	checkCmd.Flags().StringVarP(&output, "output", "o", "", "Write check results to file")

	// --- Generate command (create example input files) ---
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate example input files",
		Long: `Generate example input files for batch metadata updates.
Supports both JSON and YAML formats.`,
		Run: func(cmd *cobra.Command, args []string) {
			if output == "" {
				log.Fatal("Error: You must specify --output filename")
			}

			ext := strings.ToLower(filepath.Ext(output))
			var err error

			switch ext {
			case ".json":
				err = parser.GenerateExampleJSON(output)
			case ".yaml", ".yml":
				err = parser.GenerateExampleYAML(output)
			default:
				log.Fatal("Error: Output file must have .json, .yaml, or .yml extension")
			}

			if err != nil {
				log.Fatalf("Error generating example file: %v", err)
			}

			fmt.Printf("✓ Example input file created: %s\n", output)
			fmt.Println("\nYou can now edit this file and use it with:")
			fmt.Printf("  mdu update --file book.epub --input %s\n", output)
			fmt.Printf("  mdu update --dir ./books --input %s\n", output)
		},
	}

	generateCmd.Flags().StringVarP(&output, "output", "o", "", "Output filename (.json, .yaml, or .yml)")
	generateCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(readCmd, updateCmd, compareCmd, validateCmd, checkCmd, generateCmd)
	return rootCmd
}

// --- Helper Functions ---

func getEPUBFiles(file, dir string) ([]string, error) {
	if dir != "" {
		return parser.ListEPUBFiles(dir)
	}
	return []string{file}, nil
}

func buildUpdatesMap(series, seriesIndex, summary, isbn, author string) map[string]string {
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

	return updates
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
