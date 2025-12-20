package cli

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"mdu/mangasrc"
	"mdu/metadata"
	"mdu/parser"
)

const (
	maxIntegrityRetries = 3
)

func NewComicInfoCmd() *cobra.Command {
	comicInfoCmd := &cobra.Command{
		Use:   "comicinfo",
		Short: "ComicInfo.xml metadata operations",
		Long:  `Read, update, and validate ComicInfo.xml metadata in CBZ files.`,
	}

	comicInfoCmd.AddCommand(
		newComicInfoReadCmd(),
		newComicInfoUpdateCmd(),
		newComicInfoCompareCmd(),
		newComicInfoValidateCmd(),
		newComicInfoCheckCmd(),
		newComicInfoGenerateCmd(),
		newComicInfoSearchCmd(),
	)

	return comicInfoCmd
}

func newComicInfoReadCmd() *cobra.Command {
	var file, dir, output string
	var listFields bool

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read metadata from CBZ files",
		Long: `Read and display ComicInfo.xml metadata from CBZ files.
All ComicInfo.xml fields are displayed by default.`,
		Example: `  mdu comicinfo read --file comic.cbz
  mdu comicinfo read --dir ./comics
  mdu comicinfo read --file comic.cbz --output metadata.txt`,
		Run: func(cmd *cobra.Command, args []string) {
			if listFields {
				fmt.Println("Supported ComicInfo.xml metadata fields (Kavita-compatible):")
				fmt.Println("  Series           - Series name")
				fmt.Println("  Number           - Issue number")
				fmt.Println("  Volume           - Volume number")
				fmt.Println("  Summary          - Issue summary/description")
				fmt.Println("  Notes            - Additional notes")
				fmt.Println("  Writer           - Writer(s)")
				fmt.Println("  Penciller        - Penciller(s)")
				fmt.Println("  Inker            - Inker(s)")
				fmt.Println("  Colorist         - Colorist(s)")
				fmt.Println("  Letterer         - Letterer(s)")
				fmt.Println("  CoverArtist      - Cover artist(s)")
				fmt.Println("  Editor           - Editor(s)")
				fmt.Println("  Publisher        - Publisher name")
				fmt.Println("  Imprint          - Publisher imprint")
				fmt.Println("  Genre            - Genre(s)")
				fmt.Println("  Tags             - Tags")
				fmt.Println("  Web              - Web link")
				fmt.Println("  PageCount        - Number of pages")
				fmt.Println("  LanguageISO      - Language code (e.g., en)")
				fmt.Println("  Format           - Format (e.g., TPB, HC)")
				fmt.Println("  AgeRating        - Age rating")
				fmt.Println("  Year             - Publication year")
				fmt.Println("  Month            - Publication month")
				fmt.Println("  Day              - Publication day")
				return
			}

			if file == "" && dir == "" {
				log.Fatal("Error: You must specify either --file or --dir")
			}

			allFiles, err := getCBZFiles(file, dir)
			if err != nil {
				log.Fatalf("Error getting CBZ files: %v", err)
			}

			var outputStr string
			for _, f := range allFiles {
				md, err := metadata.ReadComicInfo(f)
				if err != nil {
					log.Printf("Error reading %s: %v", f, err)
					continue
				}
				outputStr += parser.RenderComicInfo(md, filepath.Base(f))
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

	cmd.Flags().StringVar(&file, "file", "", "Target CBZ file")
	cmd.Flags().StringVar(&dir, "dir", "", "Target directory containing CBZ files")
	cmd.Flags().BoolVar(&listFields, "list-fields", false, "List all supported metadata fields")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write output to file")

	return cmd
}

func newComicInfoUpdateCmd() *cobra.Command {
	var file, dir, output, inputFile string
	var series, number, volume, summary, writer, publisher string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update metadata in CBZ files",
		Long: `Update ComicInfo.xml metadata in CBZ files while preserving existing data.

You can specify metadata using either:
  1. Command-line flags (--series, --number, --writer, etc.)
  2. An input file (--input) in JSON or YAML format`,
		Example: `  mdu comicinfo update --file comic.cbz --writer "John Doe" --series "Amazing Comics"
  mdu comicinfo update --dir ./comics --input metadata.json
  mdu comicinfo update --file comic.cbz --number "12" --volume "2"`,
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
				updates = buildComicInfoUpdatesMap(series, number, volume, summary, writer, publisher)
			}

			if len(updates) == 0 {
				log.Fatal("Error: No metadata fields specified for update")
			}

			allFiles, err := getCBZFiles(file, dir)
			if err != nil {
				log.Fatalf("Error getting CBZ files: %v", err)
			}

			fmt.Printf("Updating %d file(s) with the following changes:\n", len(allFiles))
			for k, v := range updates {
				fmt.Printf("  %s: %s\n", k, v)
			}
			fmt.Println()

			var outputStr string
			successCount := 0

			for _, f := range allFiles {
				if err := metadata.UpdateComicInfo(f, f, updates); err != nil {
					log.Printf("✗ Error updating %s: %v", filepath.Base(f), err)
					continue
				}

				md, err := metadata.ReadComicInfo(f)
				if err != nil {
					log.Printf("Warning: Updated %s but couldn't read metadata: %v",
						filepath.Base(f), err)
				} else {
					outputStr += parser.RenderComicInfo(md, filepath.Base(f))
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

	cmd.Flags().StringVar(&file, "file", "", "Target CBZ file")
	cmd.Flags().StringVar(&dir, "dir", "", "Target directory for batch update")
	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file (JSON or YAML) with metadata updates")
	cmd.Flags().StringVar(&series, "series", "", "Series name")
	cmd.Flags().StringVar(&number, "number", "", "Issue number")
	cmd.Flags().StringVar(&volume, "volume", "", "Volume number")
	cmd.Flags().StringVar(&summary, "summary", "", "Issue summary")
	cmd.Flags().StringVar(&writer, "writer", "", "Writer name")
	cmd.Flags().StringVar(&publisher, "publisher", "", "Publisher name")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write results to file")

	return cmd
}

func newComicInfoCompareCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "compare <original.cbz> <modified.cbz>",
		Short: "Compare metadata between two CBZ files",
		Long: `Compare ComicInfo.xml metadata from two CBZ files and generate a diff report.
Useful for validating changes after updates.`,
		Example: `  mdu comicinfo compare comic.cbz comic_modified.cbz
  mdu comicinfo compare original.cbz modified.cbz --output diff.txt`,
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

			diff, err := metadata.CompareComicInfo(original, modified)
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

func newComicInfoValidateCmd() *cobra.Command {
	var file, output string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate CBZ file structure and ComicInfo.xml",
		Long:  `Validates that a CBZ file has proper structure and valid ComicInfo.xml.`,
		Example: `  mdu comicinfo validate --file comic.cbz
  mdu comicinfo validate --file comic.cbz --output validation.txt`,
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" {
				log.Fatal("Error: You must specify --file")
			}

			if !fileExists(file) {
				log.Fatalf("Error: File not found: %s", file)
			}

			// Validate integrity
			if err := validateCBZIntegrity(file); err != nil {
				log.Fatalf("✗ Validation failed: %v", err)
			}

			fmt.Printf("✓ File structure is valid: %s\n", filepath.Base(file))

			// Read and display metadata
			md, err := metadata.ReadComicInfo(file)
			if err != nil {
				log.Fatalf("✗ Error reading ComicInfo.xml: %v", err)
			}

			result := fmt.Sprintf("\n✓ ComicInfo.xml is valid\n\n%s", parser.RenderComicInfo(md, filepath.Base(file)))
			fmt.Print(result)

			if output != "" {
				if err := os.WriteFile(output, []byte(result), 0644); err != nil {
					log.Fatalf("Error writing output file: %v", err)
				}
				fmt.Printf("\n✓ Validation report written to %s\n", output)
			}
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Target CBZ file to validate")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write validation report to file")

	return cmd
}

func newComicInfoCheckCmd() *cobra.Command {
	var file, dir, output string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check CBZ file validity and structure",
		Long: `Validates that CBZ files have proper structure:
- File is a valid ZIP archive
- ComicInfo.xml exists (if present)
- ComicInfo.xml has valid XML structure
- Reports whether ComicInfo.xml is present or missing`,
		Example: `  mdu comicinfo check --file comic.cbz
  mdu comicinfo check --dir ./comics --output report.txt`,
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" && dir == "" {
				log.Fatal("Error: You must specify either --file or --dir")
			}

			allFiles, err := getCBZFiles(file, dir)
			if err != nil {
				log.Fatalf("Error getting CBZ files: %v", err)
			}

			var results strings.Builder
			validCount := 0
			invalidCount := 0
			missingCount := 0

			for _, f := range allFiles {
				results.WriteString(fmt.Sprintf("\n=== Checking: %s ===\n", filepath.Base(f)))

				md, err := metadata.ReadComicInfo(f)

				if err != nil {
					if strings.Contains(err.Error(), "ComicInfo.xml not found") {
						results.WriteString("⚠  CBZ structure valid but ComicInfo.xml not found\n")
						missingCount++
					} else {
						results.WriteString(fmt.Sprintf("✗ INVALID: %v\n", err))
						invalidCount++
					}
					continue
				}

				results.WriteString("✓ CBZ structure valid\n")
				results.WriteString("✓ ComicInfo.xml found and readable\n")
				results.WriteString(parser.RenderComicInfo(md, ""))

				validCount++
			}

			results.WriteString("\n=== Summary ===\n")
			results.WriteString(fmt.Sprintf("Valid (with ComicInfo.xml): %d\n", validCount))
			results.WriteString(fmt.Sprintf("Valid (missing ComicInfo.xml): %d\n", missingCount))
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

	cmd.Flags().StringVar(&file, "file", "", "Target CBZ file to check")
	cmd.Flags().StringVar(&dir, "dir", "", "Target directory to check")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write check results to file")

	return cmd
}

func newComicInfoGenerateCmd() *cobra.Command {
	var file, dir, output string
	var mangadexID string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate ComicInfo.xml in CBZ files from MangaDex metadata",
		Long: `Generate and add ComicInfo.xml to CBZ files using MangaDex API metadata.

This command:
  1. Fetches manga metadata from MangaDex using the provided manga ID
  2. Fetches author/artist names from MangaDex
  3. Extracts chapter numbers from CBZ filenames
  4. Creates ComicInfo.xml for each chapter
  5. Repackages CBZ files with the new ComicInfo.xml
  6. Validates integrity using checksum verification
  7. Replaces original files only after successful validation`,
		Example: `  mdu comicinfo generate --mangadex-id "abc123-def456" --file ch001.cbz
  mdu comicinfo generate --mangadex-id "abc123-def456" --dir ./chapters`,
		Run: func(cmd *cobra.Command, args []string) {
			if mangadexID == "" {
				log.Fatal("Error: --mangadex-id is required")
			}
			if file == "" && dir == "" {
				log.Fatal("Error: You must specify either --file or --dir")
			}
			if file != "" && dir != "" {
				log.Fatal("Error: You cannot specify both --file and --dir")
			}

			// Step 1: Fetch MangaDex metadata
			fmt.Printf("🔍 Fetching metadata from MangaDex (ID: %s)...\n", mangadexID)
			mangadexMetadata, err := mangasrc.TitleMetadata(mangadexID)
			if err != nil {
				log.Fatalf("Error fetching MangaDex metadata: %v", err)
			}
			fmt.Println("✓ Metadata fetched successfully")

			// Step 2: Convert to ComicInfo struct and fetch author names
			baseComicInfo, err := parser.MangaDexToComicInfo(mangadexMetadata)
			if err != nil {
				log.Fatalf("Error converting MangaDex metadata to ComicInfo: %v", err)
			}

			// Fetch actual author/artist names
			if err := enrichComicInfoWithAuthors(baseComicInfo, mangadexMetadata); err != nil {
				log.Printf("⚠️  Warning: Could not fetch all author names: %v", err)
			}

			fmt.Println("✓ Metadata converted to ComicInfo format")

			// Step 3: Get all CBZ files to process
			allFiles, err := resolveCBZFiles(file, dir)
			if err != nil {
				log.Fatalf("Error resolving CBZ files: %v", err)
			}

			if len(allFiles) == 0 {
				log.Fatal("Error: No CBZ files found to process")
			}

			fmt.Printf("\n📚 Processing %d file(s)...\n\n", len(allFiles))

			// Step 4: Process each file
			successCount := 0
			skippedCount := 0
			failedCount := 0
			var outputStr string

			for i, cbzPath := range allFiles {
				fileName := filepath.Base(cbzPath)
				fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(allFiles), fileName)

				// Extract chapter number from filename
				chapterNum := parser.ExtractChapterNumber(fileName)
				if chapterNum == "" {
					log.Printf("⚠️  Could not extract chapter number from '%s', skipping\n\n", fileName)
					skippedCount++
					continue
				}
				fmt.Printf("  📄 Extracted chapter number: %s\n", chapterNum)

				// Create a copy of the base ComicInfo and update with chapter number
				chapterComicInfo := *baseComicInfo
				chapterComicInfo.Number = chapterNum
				chapterComicInfo.Title = fmt.Sprintf("Chapter %s", chapterNum)

				// Process this chapter with integrity validation
				if err := processChapterWithIntegrity(cbzPath, &chapterComicInfo); err != nil {
					log.Printf("✗ Error processing %s: %v\n\n", fileName, err)
					failedCount++
					continue
				}

				// Read back the metadata for output
				md, err := metadata.ReadComicInfo(cbzPath)
				if err != nil {
					log.Printf("⚠️  Processed %s but couldn't read metadata: %v\n", fileName, err)
				} else {
					outputStr += parser.RenderComicInfo(md, fileName)
				}

				fmt.Printf("✓ Successfully processed: %s\n\n", fileName)
				successCount++
			}

			// Final summary
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("✓ Successfully processed: %d\n", successCount)
			if skippedCount > 0 {
				fmt.Printf("⚠️  Skipped (no chapter number): %d\n", skippedCount)
			}
			if failedCount > 0 {
				fmt.Printf("✗ Failed: %d\n", failedCount)
			}
			fmt.Printf("Total files: %d\n\n", len(allFiles))
			fmt.Print(outputStr)

			if output != "" {
				if err := os.WriteFile(output, []byte(outputStr), 0644); err != nil {
					log.Fatalf("Error writing output file: %v", err)
				}
				fmt.Printf("✓ Results written to %s\n", output)
			}
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Target CBZ file (absolute or relative path)")
	cmd.Flags().StringVar(&dir, "dir", "", "Target directory containing CBZ files")
	cmd.Flags().StringVar(&mangadexID, "mangadex-id", "", "MangaDex manga ID to fetch metadata (required)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write results to file")

	return cmd
}

func newComicInfoSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <title>",
		Short: "Search MangaDex for a manga title and return the best match",
		Long: `Search MangaDex by title (multiple words) and use token match scoring to find the best match.
Returns the MangaDex ID and title information.

The search uses English titles only and performs fuzzy matching to find the closest result.`,
		Example: `  mdu comicinfo search One Piece
  mdu comicinfo search "Attack on Titan"
  mdu comicinfo search Berserk`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Join all positional args into a single title
			title := strings.Join(args, " ")

			fmt.Printf("\n🔍 Searching MangaDex for: %s\n\n", title)

			// Search MangaDex
			mdTitles, err := mangasrc.MangadexTitleSearch(title)
			if err != nil {
				log.Fatalf("Error searching for title '%s': %v", title, err)
			}

			if len(mdTitles) == 0 {
				fmt.Printf("No results found for: %s\n", title)
				return
			}

			// Extract all the returned titles
			searchResults := parser.ExtractEnglishTitles(mdTitles)

			// Search for the best match
			nameMatch, score := parser.BestTokenMatch(title, searchResults)

			result := parser.FindEntryByTitle(mdTitles, nameMatch)
			if result == nil {
				log.Fatalf("Error: Could not find matching entry")
			}

			// Print results
			fmt.Printf("Best Match (Score: %.2f):\n", score)
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			parser.PrintTitleSearchResults([]mangasrc.MangadexTitleSearchResponse{*result})
			fmt.Printf("\n💡 Use this ID with: mdu comicinfo generate --mangadex-id %s --dir <path>\n", result.ID)
		},
	}

	return cmd
}

// ============================================================================
// Helper Functions
// ============================================================================

// enrichComicInfoWithAuthors fetches actual author and artist names from MangaDex
// and updates the ComicInfo struct with real names instead of placeholder IDs
func enrichComicInfoWithAuthors(comic *parser.ComicInfo, md *mangasrc.MangadexTitleMetadata) error {
	var authors []string
	var artists []string

	for _, rel := range md.Relationships {
		relMap, ok := rel.(map[string]interface{})
		if !ok {
			continue
		}

		relType, ok := relMap["type"].(string)
		if !ok {
			continue
		}

		relID, ok := relMap["id"].(string)
		if !ok {
			continue
		}

		// Fetch author/artist name
		name, err := parser.MangaAuthorName(relID)
		if err != nil {
			log.Printf("⚠️  Warning: Could not fetch name for %s %s: %v", relType, relID, err)
			continue
		}

		switch relType {
		case "author":
			authors = append(authors, name)
		case "artist":
			artists = append(artists, name)
		}
	}

	// Update ComicInfo with fetched names
	if len(authors) > 0 {
		comic.Writer = strings.Join(authors, ", ")
	}
	if len(artists) > 0 {
		comic.Penciller = strings.Join(artists, ", ")
	}

	return nil
}

// resolveCBZFiles resolves file or directory to a list of CBZ files
func resolveCBZFiles(file, dir string) ([]string, error) {
	if file != "" {
		absPath, err := resolveFilePath(file)
		if err != nil {
			return nil, fmt.Errorf("error resolving file path: %w", err)
		}
		if !strings.HasSuffix(strings.ToLower(absPath), ".cbz") {
			return nil, fmt.Errorf("file must be a .cbz file")
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("file does not exist: %s", absPath)
		}
		return []string{absPath}, nil
	}

	// Handle directory
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("error resolving directory path: %w", err)
	}
	return getCBZFiles("", absDir)
}

// resolveFilePath resolves a file path to absolute path
func resolveFilePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	if strings.Contains(path, string(filepath.Separator)) {
		return filepath.Abs(path)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Join(cwd, path), nil
}

// processChapterWithIntegrity processes a CBZ file with integrity validation and retry logic
func processChapterWithIntegrity(cbzPath string, comicInfo *parser.ComicInfo) error {
	fileName := filepath.Base(cbzPath)
	chapterName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// Create temp directory for this operation
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("mdu_%s_%d", chapterName, time.Now().Unix()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	fmt.Printf("  📁 Created temp directory: %s\n", tempDir)

	// Extract original CBZ contents
	fmt.Printf("  📦 Extracting CBZ contents...\n")
	if err := extractCBZ(cbzPath, tempDir); err != nil {
		return fmt.Errorf("failed to extract CBZ: %w", err)
	}

	// Write ComicInfo.xml to temp directory
	comicInfoPath := filepath.Join(tempDir, "ComicInfo.xml")
	fmt.Printf("  📝 Writing ComicInfo.xml...\n")
	if _, err := parser.WriteComicInfo(comicInfo, comicInfoPath); err != nil {
		return fmt.Errorf("failed to write ComicInfo.xml: %w", err)
	}

	// Attempt to create and validate new CBZ with retries
	var lastErr error
	for attempt := 1; attempt <= maxIntegrityRetries; attempt++ {
		if attempt > 1 {
			fmt.Printf("  🔄 Retry attempt %d/%d...\n", attempt, maxIntegrityRetries)
		}

		// Create new CBZ
		newCBZPath := cbzPath + ".new"
		fmt.Printf("  🗜️  Creating new CBZ file...\n")
		if err := createCBZ(tempDir, newCBZPath); err != nil {
			lastErr = fmt.Errorf("failed to create CBZ: %w", err)
			os.Remove(newCBZPath)
			continue
		}

		// Validate integrity with checksum
		fmt.Printf("  🔐 Validating CBZ integrity...\n")
		if err := validateCBZIntegrity(newCBZPath); err != nil {
			lastErr = fmt.Errorf("integrity validation failed: %w", err)
			os.Remove(newCBZPath)
			continue
		}

		// Additional checksum verification
		if err := verifyCBZChecksum(newCBZPath, tempDir); err != nil {
			lastErr = fmt.Errorf("checksum verification failed: %w", err)
			os.Remove(newCBZPath)
			continue
		}

		fmt.Printf("  ✓ Integrity check passed\n")

		// Replace original file with new file
		fmt.Printf("  🔄 Replacing original file...\n")
		if err := os.Remove(cbzPath); err != nil {
			os.Remove(newCBZPath)
			return fmt.Errorf("failed to remove original file: %w", err)
		}
		if err := os.Rename(newCBZPath, cbzPath); err != nil {
			return fmt.Errorf("failed to rename new file: %w", err)
		}

		return nil
	}

	return fmt.Errorf("failed after %d attempts: %v", maxIntegrityRetries, lastErr)
}

// extractCBZ extracts a CBZ (zip) file to a directory
func extractCBZ(cbzPath, destDir string) error {
	r, err := zip.OpenReader(cbzPath)
	if err != nil {
		return fmt.Errorf("failed to open CBZ: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		// Skip ComicInfo.xml if it already exists (we'll write a new one)
		if f.Name == "ComicInfo.xml" {
			continue
		}

		fpath := filepath.Join(destDir, f.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("failed to open zip entry: %w", err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("failed to copy file content: %w", err)
		}
	}
	return nil
}

// createCBZ creates a CBZ (zip) file from a directory
func createCBZ(sourceDir, cbzPath string) error {
	file, err := os.Create(cbzPath)
	if err != nil {
		return fmt.Errorf("failed to create CBZ file: %w", err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Use forward slashes in zip file
		relPath = filepath.ToSlash(relPath)

		zipFile, err := w.Create(relPath)
		if err != nil {
			return err
		}

		fsFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fsFile.Close()

		_, err = io.Copy(zipFile, fsFile)
		return err
	})
}

// validateCBZIntegrity validates that a CBZ file can be opened and all files are readable
func validateCBZIntegrity(cbzPath string) error {
	r, err := zip.OpenReader(cbzPath)
	if err != nil {
		return fmt.Errorf("cannot open CBZ: %w", err)
	}
	defer r.Close()

	// Check that we can read all files
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("cannot open file %s in CBZ: %w", f.Name, err)
		}

		// Try to read the entire file to verify it's not corrupted
		if _, err := io.Copy(io.Discard, rc); err != nil {
			rc.Close()
			return fmt.Errorf("cannot read file %s in CBZ: %w", f.Name, err)
		}
		rc.Close()
	}

	// Verify ComicInfo.xml exists
	hasComicInfo := false
	for _, f := range r.File {
		if f.Name == "ComicInfo.xml" {
			hasComicInfo = true
			break
		}
	}
	if !hasComicInfo {
		return fmt.Errorf("ComicInfo.xml not found in CBZ")
	}

	return nil
}

// verifyCBZChecksum performs additional checksum verification on the CBZ contents
func verifyCBZChecksum(cbzPath, sourceDir string) error {
	r, err := zip.OpenReader(cbzPath)
	if err != nil {
		return fmt.Errorf("cannot open CBZ for checksum: %w", err)
	}
	defer r.Close()

	// Verify each file's checksum against the source
	for _, f := range r.File {
		// Read file from CBZ
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("cannot open %s in CBZ: %w", f.Name, err)
		}

		cbzHash := sha256.New()
		if _, err := io.Copy(cbzHash, rc); err != nil {
			rc.Close()
			return fmt.Errorf("cannot hash %s in CBZ: %w", f.Name, err)
		}
		rc.Close()
		cbzSum := cbzHash.Sum(nil)

		// Read original file from source directory
		sourcePath := filepath.Join(sourceDir, f.Name)
		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return fmt.Errorf("cannot open source file %s: %w", sourcePath, err)
		}

		sourceHash := sha256.New()
		if _, err := io.Copy(sourceHash, sourceFile); err != nil {
			sourceFile.Close()
			return fmt.Errorf("cannot hash source file %s: %w", sourcePath, err)
		}
		sourceFile.Close()
		sourceSum := sourceHash.Sum(nil)

		// Compare checksums
		if string(cbzSum) != string(sourceSum) {
			return fmt.Errorf("checksum mismatch for %s", f.Name)
		}
	}

	return nil
}

// getCBZFiles returns a list of CBZ files from file or directory
func getCBZFiles(file, dir string) ([]string, error) {
	if dir != "" {
		return parser.ListCBZFiles(dir)
	}
	return []string{file}, nil
}

// buildComicInfoUpdatesMap creates a map of ComicInfo field updates
func buildComicInfoUpdatesMap(series, number, volume, summary, writer, publisher string) map[string]string {
	updates := make(map[string]string)

	if series != "" {
		updates["Series"] = series
	}
	if number != "" {
		updates["Number"] = number
	}
	if volume != "" {
		updates["Volume"] = volume
	}
	if summary != "" {
		updates["Summary"] = summary
	}
	if writer != "" {
		updates["Writer"] = writer
	}
	if publisher != "" {
		updates["Publisher"] = publisher
	}

	return updates
}
