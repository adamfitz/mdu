package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"

	"mdu/metadata"
	"mdu/parser"
	"mdu/ranobedb"
)

func NewEPUBCmd() *cobra.Command {
	epubCmd := &cobra.Command{
		Use:   "epub",
		Short: "EPUB metadata operations",
		Long:  `Read, update, search and validate EPUB file metadata.`,
	}

	epubCmd.AddCommand(
		newEPUBReadCmd(),
		newEPUBUpdateCmd(),
		newEPUBCompareCmd(),
		newEPUBValidateCmd(),
		newEPUBCheckCmd(),
		newRanobeSearchCmd(),
		newEpubGenerateCmd(),
	)

	return epubCmd
}

func newEPUBReadCmd() *cobra.Command {
	var file, dir, output string
	var listFields, all bool

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read metadata from EPUB files",
		Long:  `Read and display metadata from EPUB files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listFields {
				fmt.Println("Supported EPUB metadata fields (Kavita-compatible):")
				fmt.Println("  author, summary, publisher, isbn, calibre:series, calibre:series_index, subject")
				return nil
			}

			if file == "" && dir == "" {
				return fmt.Errorf("you must specify either --file or --dir")
			}

			allFiles, err := getEPUBFiles(file, dir)
			if err != nil {
				return fmt.Errorf("getting EPUB files: %w", err)
			}

			var outputStr string
			for _, f := range allFiles {
				md, err := metadata.Read(f, all)
				if err != nil {
					fmt.Printf("Error reading %s: %v\n", f, err)
					continue
				}
				outputStr += parser.RenderMetadataWithHeader(filepath.Base(f), md)
			}

			fmt.Print(outputStr)

			if output != "" {
				if err := os.WriteFile(output, []byte(outputStr), 0644); err != nil {
					return fmt.Errorf("writing output file: %w", err)
				}
				fmt.Printf("\n✓ Metadata written to %s\n", output)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Target EPUB file")
	cmd.Flags().StringVar(&dir, "dir", "", "Target directory")
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" && dir == "" {
				return fmt.Errorf("you must specify either --file or --dir")
			}

			var updates map[string]string
			var err error

			if inputFile != "" {
				updates, err = parser.ParseInputFile(inputFile)
				if err != nil {
					return fmt.Errorf("parsing input file: %w", err)
				}
				fmt.Printf("✓ Loaded metadata from input file: %s\n", inputFile)
			} else {
				updates = buildUpdatesMap(series, seriesIndex, summary, isbn, author)
			}

			if len(updates) == 0 {
				return fmt.Errorf("no metadata fields specified for update")
			}

			allFiles, err := getEPUBFiles(file, dir)
			if err != nil {
				return fmt.Errorf("getting EPUB files: %w", err)
			}

			fmt.Printf("Updating %d file(s) with changes:\n", len(allFiles))
			for k, v := range updates {
				fmt.Printf("  %s: %s\n", k, v)
			}
			fmt.Println()

			for _, f := range allFiles {
				if createBackup {
					backupPath := f + ".backup"
					if err := copyFile(f, backupPath); err == nil {
						fmt.Printf("✓ Backup created: %s\n", backupPath)
					}
				}

				if err := metadata.Update(f, f, updates); err != nil {
					fmt.Printf("✗ Error updating %s: %v\n", filepath.Base(f), err)
					continue
				}
				fmt.Printf("✓ Updated: %s\n", filepath.Base(f))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Target EPUB file")
	cmd.Flags().StringVar(&dir, "dir", "", "Target directory")
	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file with metadata updates")
	cmd.Flags().StringVar(&series, "series", "", "Series name")
	cmd.Flags().StringVar(&seriesIndex, "series-index", "", "Series index")
	cmd.Flags().StringVar(&summary, "summary", "", "Book summary")
	cmd.Flags().StringVar(&isbn, "isbn", "", "ISBN identifier")
	cmd.Flags().StringVar(&author, "author", "", "Author name")
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" && dir == "" {
				return fmt.Errorf("you must specify either --file or --dir")
			}

			allFiles, err := getEPUBFiles(file, dir)
			if err != nil {
				return fmt.Errorf("getting EPUB files: %w", err)
			}

			var results strings.Builder
			validCount, invalidCount := 0, 0

			for _, f := range allFiles {
				results.WriteString(fmt.Sprintf("\n=== Checking: %s ===\n", filepath.Base(f)))
				md, err := metadata.Read(f, false)
				if err != nil {
					results.WriteString(fmt.Sprintf("❌ INVALID: %v\n", err))
					invalidCount++
					continue
				}

				results.WriteString("✓ EPUB structure valid\n")
				results.WriteString(fmt.Sprintf("  Title: %s\n", md["title"]))
				results.WriteString(fmt.Sprintf("  Author: %s\n", md["author"]))
				validCount++
			}

			results.WriteString("\n=== Summary ===\n")
			results.WriteString(fmt.Sprintf("Valid: %d\nInvalid: %d\n", validCount, invalidCount))
			fmt.Print(results.String())

			if output != "" {
				if err := os.WriteFile(output, []byte(results.String()), 0644); err != nil {
					return fmt.Errorf("writing output file: %w", err)
				}
				fmt.Printf("\n✓ Check results written to %s\n", output)
			}

			return nil
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

// Search RanobeDB API for novel titles and display results with relevance scores.
func newRanobeSearchCmd() *cobra.Command {
	var showURL bool

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search RanobeDB for a novel title and return best matches",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")

			fmt.Printf("\n🔍 Searching RanobeDB for: %s\n\n", query)

			db := ranobedb.RanobeDB{}
			search, err := db.SearchNovel(query)
			if err != nil {
				return fmt.Errorf("failed to search RanobeDB: %w", err)
			}

			if len(search.Series) == 0 {
				fmt.Printf("No results found for: %s\n", query)
				return nil
			}

			// ---- scoring ----
			type scoredResult struct {
				series ranobedb.Series
				score  float64
			}

			scoredResults := make([]scoredResult, 0, len(search.Series))

			for _, s := range search.Series {
				score := parser.ScoreTitleTokens(query, s.Title)
				scoredResults = append(scoredResults, scoredResult{
					series: s,
					score:  score,
				})
			}

			// ---- sort ----
			sort.Slice(scoredResults, func(i, j int) bool {
				return scoredResults[i].score > scoredResults[j].score
			})

			// ---- calculate display width (FIXED UTF-8 ALIGNMENT) ----
			maxTitleWidth := 20
			for _, r := range scoredResults {
				if w := runewidth.StringWidth(r.series.Title); w > maxTitleWidth {
					maxTitleWidth = w
				}
			}

			// optional clamp (prevents stupid long titles breaking layout)
			if maxTitleWidth > 60 {
				maxTitleWidth = 60
			}

			// ---- printer ----
			printEntry := func(r scoredResult, highlight bool) {
				title := r.series.Title

				width := runewidth.StringWidth(title)
				padding := maxTitleWidth - width
				if padding < 0 {
					padding = 0
				}

				fmt.Printf("%s%s | ID: %d\n",
					title,
					strings.Repeat(" ", padding),
					r.series.ID,
				)

				if showURL {
					fmt.Printf("  https://ranobedb.org/series/%d\n", r.series.ID)
				}

				if highlight {
					fmt.Printf("  Best Match (Score: %.2f)\n", r.score)
				}
			}

			// ---- best match ----
			fmt.Println("Best Match:")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			printEntry(scoredResults[0], true)

			// ---- others ----
			if len(scoredResults) > 1 {
				count := len(scoredResults) - 1
				if count > 9 {
					count = 9
				}

				fmt.Println("\nOther Matches:")
				fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

				for i := 1; i <= count; i++ {
					printEntry(scoredResults[i], false)
				}
			}

			fmt.Println("\nUse with: mdu epub generate --ranobedb <id>")

			return nil
		},
	}

	cmd.Flags().BoolVar(&showURL, "url", false, "Show RanobeDB URL")

	return cmd
}

// Fetch full novel info from RanobeDB by ID, including metadata and book list and update file metadata
func newEpubGenerateCmd() *cobra.Command {
	var ranobeID int
	var filePath string
	var dirPath string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate EPUB metadata using RanobeDB",
		Long: `Fetch metadata from RanobeDB and generate OPF metadata.

Usage:
  mdu epub generate --ranobedb <id> <file|directory>
  mdu epub generate --ranobedb <id> --file <file>
  mdu epub generate --ranobedb <id> --dir <directory>`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ranobeID == 0 {
				return fmt.Errorf("--ranobedb is required")
			}

			// resolve target
			if len(args) > 0 {
				filePath = args[0]
			}

			if filePath == "" && dirPath == "" {
				return fmt.Errorf("must provide a file or directory")
			}

			db := ranobedb.RanobeDB{}

			fmt.Printf("Fetching metadata for series ID: %d\n", ranobeID)

			novel, err := db.GetNovelInfo(strconv.Itoa(ranobeID))
			if err != nil {
				return fmt.Errorf("failed to fetch novel info: %w", err)
			}

			fmt.Println("Building OPF metadata...")
			opf := metadata.BuildMetadataMap(novel)

			// Apply to file
			if filePath != "" {
				fmt.Printf("Updating EPUB file: %s\n", filePath)

				// Write to a temp file first
				tmpFile, err := os.CreateTemp(filepath.Dir(filePath), "*.epub.tmp")
				if err != nil {
					return fmt.Errorf("failed to create temp file: %w", err)
				}
				tmpPath := tmpFile.Name()
				tmpFile.Close()

				// Write updated metadata to temp
				if err := metadata.WriteOPFToFile(filePath, opf, tmpPath); err != nil {
					os.Remove(tmpPath)
					return fmt.Errorf("failed to write metadata to temp file: %w", err)
				}

				// Verify the temp file is readable and metadata is not corrupted
				if _, err := metadata.Read(tmpPath, false); err != nil {
					os.Remove(tmpPath)
					return fmt.Errorf("metadata verification failed, original file untouched: %w", err)
				}

				// Replace original with temp
				if err := os.Rename(tmpPath, filePath); err != nil {
					os.Remove(tmpPath)
					return fmt.Errorf("failed to replace original file: %w", err)
				}

				fmt.Println("EPUB metadata updated successfully.")
				return nil
			}

			// Apply to directory
			if dirPath != "" {
				fmt.Printf("Updating EPUB files in directory: %s\n", dirPath)

				entries, err := os.ReadDir(dirPath)
				if err != nil {
					return fmt.Errorf("failed to read directory: %w", err)
				}

				for _, entry := range entries {
					if entry.IsDir() || filepath.Ext(entry.Name()) != ".epub" {
						continue
					}

					filePath := filepath.Join(dirPath, entry.Name())

					// Write to temp first
					tmpFile, err := os.CreateTemp(dirPath, "*.epub.tmp")
					if err != nil {
						return fmt.Errorf("failed to create temp file for %s: %w", entry.Name(), err)
					}
					tmpPath := tmpFile.Name()
					tmpFile.Close()

					if err := metadata.WriteOPFToFile(filePath, opf, tmpPath); err != nil {
						os.Remove(tmpPath)
						fmt.Printf("✗ Failed to update %s: %v\n", entry.Name(), err)
						continue
					}

					// Verify
					if _, err := metadata.Read(tmpPath, false); err != nil {
						os.Remove(tmpPath)
						fmt.Printf("✗ Verification failed for %s, original untouched: %v\n", entry.Name(), err)
						continue
					}

					// Replace original
					if err := os.Rename(tmpPath, filePath); err != nil {
						os.Remove(tmpPath)
						fmt.Printf("✗ Failed to replace %s: %v\n", entry.Name(), err)
						continue
					}

					fmt.Printf("✓ Updated: %s\n", entry.Name())
				}

				fmt.Println("EPUB metadata updated successfully.")
				return nil
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&ranobeID, "ranobedb", 0, "RanobeDB series ID (required)")
	cmd.Flags().StringVar(&filePath, "file", "", "Target EPUB file")
	cmd.Flags().StringVar(&dirPath, "dir", "", "Target directory containing EPUB files")

	return cmd
}
