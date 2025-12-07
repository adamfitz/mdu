package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mdu/mangasrc"
	"mdu/metadata"
	"mdu/parser"
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
		newComicInfoMangadexIdSearchCmd(),
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
	var createBackup bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update metadata in CBZ files",
		Long: `Update ComicInfo.xml metadata in CBZ files while preserving existing data.
Creates backups by default unless --no-backup is specified.

You can specify metadata using either:
  1. Command-line flags (--series, --number, --writer, etc.)
  2. An input file (--input) in JSON or YAML format`,
		Example: `  mdu comicinfo update --file comic.cbz --writer "John Doe" --series "Amazing Comics"
  mdu comicinfo update --dir ./comics --input metadata.json
  mdu comicinfo update --file comic.cbz --number "12" --volume "2" --no-backup`,
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
				if createBackup {
					backupPath := f + ".backup"
					if err := copyFile(f, backupPath); err != nil {
						log.Printf("Warning: Could not create backup for %s: %v", f, err)
					} else {
						fmt.Printf("✓ Backup created: %s\n", backupPath)
					}
				}

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
	cmd.Flags().BoolVar(&createBackup, "backup", true, "Create .backup files")

	return cmd
}

func newComicInfoCompareCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "compare <original.cbz> <modified.cbz>",
		Short: "Compare metadata between two CBZ files",
		Long: `Compare ComicInfo.xml metadata from two CBZ files and generate a diff report.
Useful for validating changes after updates.`,
		Example: `  mdu comicinfo compare comic.cbz comic.cbz.backup
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
		Short: "Validate changes by comparing with backup",
		Long: `Compare a CBZ file with its .backup file to validate changes.
Requires that a backup file exists (created with --backup flag during update).`,
		Example: `  mdu comicinfo validate --file comic.cbz
  mdu comicinfo validate --file comic.cbz --output validation.txt`,
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" {
				log.Fatal("Error: You must specify --file")
			}

			backupFile := file + ".backup"
			if !fileExists(backupFile) {
				log.Fatalf("Error: Backup file not found: %s", backupFile)
			}

			diff, err := metadata.CompareComicInfo(backupFile, file)
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
						results.WriteString("⚠ CBZ structure valid but ComicInfo.xml not found\n")
						missingCount++
					} else {
						results.WriteString(fmt.Sprintf("✗ INVALID: %v\n", err))
						invalidCount++
					}
					continue
				}

				results.WriteString("✓ CBZ structure valid\n")
				results.WriteString("✓ ComicInfo.xml found and readable\n")
				results.WriteString(parser.RenderComicInfo(md))

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
	var file, dir, output, inputFile string
	var series, number, volume, summary, writer, publisher string
	var mangadexID string
	var parseFilename bool

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate ComicInfo.xml in CBZ files",
		Long: `Generate and add ComicInfo.xml to CBZ files that don't have one.
Creates backups by default.

You can specify metadata using:
  1. Command-line flags (--series, --number, --writer, etc.)
  2. An input file (--input) in JSON or YAML format
  3. MangaDex API (--mangadex-id) to fetch metadata
  4. Filename parsing (enabled by default) to extract chapter numbers

Filename parsing is ENABLED by default and will attempt to extract chapter numbers 
from filenames like: ch0001.cbz, chapter 1.cbz, chapter-01.cbz, ch1.cbz, etc.
Use --parse-filename=false to disable this behavior.

When multiple sources are used, they are applied in order:
  1. Manual flags/input file (base metadata)
  2. MangaDex API (overwrites with fetched data)
  3. Filename parsing (sets/overwrites Number field only)`,
		Example: `  mdu comicinfo generate --file comic.cbz --series "Amazing Comics" --number "1"
  mdu comicinfo generate --dir ./comics --input metadata.json
  mdu comicinfo generate --file comic.cbz --mangadex-id "abc123-def456"
  mdu comicinfo generate --dir ./manga --mangadex-id "abc123"
  mdu comicinfo generate --file ch0001.cbz --series "One Piece"
  mdu comicinfo generate --file chapter-05.cbz --series "Naruto" --parse-filename=false`,
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" && dir == "" {
				log.Fatal("Error: You must specify either --file or --dir")
			}

			// Start with base metadata from flags or input file
			var baseUpdates map[string]string
			var err error

			if inputFile != "" {
				baseUpdates, err = parser.ParseInputFile(inputFile)
				if err != nil {
					log.Fatalf("Error parsing input file: %v", err)
				}
				fmt.Printf("✓ Loaded metadata from input file: %s\n", inputFile)
			} else {
				baseUpdates = buildComicInfoUpdatesMap(series, number, volume, summary, writer, publisher)
			}

			// Fetch MangaDex metadata if ID provided
			var mangadexMetadata *mangasrc.MangadexTitleMetadata
			if mangadexID != "" {
				fmt.Printf("🔍 Fetching metadata from MangaDex (ID: %s)...\n", mangadexID)
				mangadexMetadata, err = mangasrc.TitleMetadata(mangadexID)
				if err != nil {
					log.Fatalf("Error fetching MangaDex metadata: %v", err)
				}
				fmt.Printf("✓ Successfully fetched MangaDex metadata\n")

				// Display what was fetched
				if mangadexMetadata != nil && len(mangadexMetadata.Attributes) > 0 {
					fmt.Println("  Fetched fields:")
					if title, ok := mangadexMetadata.Attributes["title"].(map[string]any); ok {
						if enTitle, ok := title["en"].(string); ok {
							fmt.Printf("    Series: %s\n", enTitle)
						}
					}
					if desc, ok := mangadexMetadata.Attributes["description"].(map[string]any); ok {
						if enDesc, ok := desc["en"].(string); ok {
							displayVal := enDesc
							if len(displayVal) > 60 {
								displayVal = displayVal[:57] + "..."
							}
							fmt.Printf("    Summary: %s\n", displayVal)
						}
					}
				}
			}

			allFiles, err := getCBZFiles(file, dir)
			if err != nil {
				log.Fatalf("Error getting CBZ files: %v", err)
			}

			fmt.Printf("\nGenerating ComicInfo.xml for %d file(s)\n", len(allFiles))
			if len(baseUpdates) > 0 {
				fmt.Println("Base metadata:")
				for k, v := range baseUpdates {
					fmt.Printf("  %s: %s\n", k, v)
				}
			}
			if !parseFilename {
				fmt.Println("⚠️  Filename parsing disabled - chapter numbers will not be extracted")
			} else {
				fmt.Println("✓ Filename parsing enabled - will extract chapter numbers")
			}
			fmt.Println()

			var outputStr string
			successCount := 0

			for _, f := range allFiles {
				// Build the updates map for this file
				updates := make(map[string]string)

				// 1. Start with base metadata (flags or input file)
				for k, v := range baseUpdates {
					updates[k] = v
				}

				// 2. Apply MangaDex metadata (overwrites base)
				if mangadexMetadata != nil {
					// Extract Series from title
					if title, ok := mangadexMetadata.Attributes["title"].(map[string]any); ok {
						if enTitle, ok := title["en"].(string); ok {
							updates["Series"] = enTitle
						}
					}

					// Extract Summary from description
					if desc, ok := mangadexMetadata.Attributes["description"].(map[string]any); ok {
						if enDesc, ok := desc["en"].(string); ok {
							updates["Summary"] = enDesc
						}
					}

					// Extract tags/genres
					if tags, ok := mangadexMetadata.Attributes["tags"].([]any); ok {
						var genres []string
						for _, tag := range tags {
							if tagMap, ok := tag.(map[string]any); ok {
								if attrs, ok := tagMap["attributes"].(map[string]any); ok {
									if name, ok := attrs["name"].(map[string]any); ok {
										if enName, ok := name["en"].(string); ok {
											genres = append(genres, enName)
										}
									}
								}
							}
						}
						if len(genres) > 0 {
							updates["Genre"] = strings.Join(genres, ", ")
						}
					}
				}

				// 3. Parse filename for chapter number if enabled (overwrites Number field)
				if parseFilename {
					chapterNum := parser.ExtractChapterNumber(filepath.Base(f))
					if chapterNum != "" {
						updates["Number"] = chapterNum
						fmt.Printf("📄 Extracted chapter number '%s' from filename: %s\n",
							chapterNum, filepath.Base(f))
					} else {
						log.Printf("⚠️  Filename parsing failed: no chapter number pattern found in '%s'",
							filepath.Base(f))
					}
				}

				if len(updates) == 0 {
					log.Printf("⚠️  No metadata to generate for %s, skipping", filepath.Base(f))
					continue
				}

				// Create backup
				backupPath := f + ".backup"
				if err := copyFile(f, backupPath); err != nil {
					log.Printf("Warning: Could not create backup for %s: %v", f, err)
				} else {
					fmt.Printf("✓ Backup created: %s\n", backupPath)
				}

				if err := metadata.GenerateComicInfo(f, f, updates); err != nil {
					log.Printf("✗ Error generating ComicInfo.xml for %s: %v", filepath.Base(f), err)
					continue
				}

				md, err := metadata.ReadComicInfo(f)
				if err != nil {
					log.Printf("Warning: Generated ComicInfo.xml for %s but couldn't read metadata: %v",
						filepath.Base(f), err)
				} else {
					outputStr += parser.RenderComicInfo(md, filepath.Base(f))
				}

				fmt.Printf("✓ Generated: %s\n", filepath.Base(f))
				successCount++
			}

			fmt.Printf("\n✓ Successfully generated ComicInfo.xml for %d of %d file(s)\n\n",
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
	cmd.Flags().StringVar(&dir, "dir", "", "Target directory for batch generation")
	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file (JSON or YAML) with metadata")
	cmd.Flags().StringVar(&series, "series", "", "Series name")
	cmd.Flags().StringVar(&number, "number", "", "Issue/chapter number")
	cmd.Flags().StringVar(&volume, "volume", "", "Volume number")
	cmd.Flags().StringVar(&summary, "summary", "", "Issue summary")
	cmd.Flags().StringVar(&writer, "writer", "", "Writer name")
	cmd.Flags().StringVar(&publisher, "publisher", "", "Publisher name")
	cmd.Flags().StringVar(&mangadexID, "mangadex-id", "", "MangaDex manga ID to fetch metadata")
	cmd.Flags().BoolVar(&parseFilename, "parse-filename", true, "Parse chapter number from filename (enabled by default)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write results to file")

	return cmd
}

func newComicInfoMangadexIdSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "search Mangadex for specific title by name and return id",
		Long: `Title search (multipe words) and token match scoring to find the best match by name (english only).
		Returns the Mangadex Id as a string.  The manga name can contain spaces, however puncuation that causes the 
		shell to evaluate as an expression will not work `,
		Example: `  mdu comicinfo search The name of the manga I am searching for`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Join all positional args into a single title
			title := strings.Join(args, " ")

			fmt.Printf("\nResults for Title: %s\n\n", title)

			// Search Mangadex
			mdTitles, searchErr := mangasrc.MangadexTitleSearch(title)
			if searchErr != nil {
				log.Fatalf("error searching for title: %s, %v", title, searchErr)
			}

			// Extract all the returned titles
			searchResults := parser.ExtractEnglishTitles(mdTitles) // extract all the english titles from name or alt name

			// search for the best match
			nameMatch, _ := parser.BestTokenMatch(title, searchResults)

			result := parser.FindEntryByTitle(mdTitles, nameMatch)

			// Print results
			parser.PrintTitleSearchResults([]mangasrc.MangadexTitleSearchResponse{*result})
		},
	}

	return cmd
}

// CBZ-specific helper
func getCBZFiles(file, dir string) ([]string, error) {
	if dir != "" {
		return parser.ListCBZFiles(dir)
	}
	return []string{file}, nil
}

// Helper function to build updates map for ComicInfo fields
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
