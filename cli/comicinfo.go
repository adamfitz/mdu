package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"strings"

	"mdu/mangasrc"
	"mdu/parser"
)

var MangadexId string // mangadex id used to retrieve title metadata

func NewMangaInfoCmd() *cobra.Command {
	mangaCmd := &cobra.Command{
		Use:   "manga",
		Short: "ComicInfo.xml metadata operations",
		Long:  `Read, update, and validate ComicInfo.xml metadata in CBZ/CBR files.`,
	}

	mangaCmd.AddCommand(
		newMangaReadCmd(),
		newMangaUpdateCmd(),
		newMangaGenerateCmd(),
		retrieveMangaMetadataCmd(MangadexId),
		retrieveMangadexIdCmd(),
	)

	return mangaCmd
}

func newMangaReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read ComicInfo.xml from manga archives",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement manga reading
			cmd.Println("ComicInfo.xml support coming soon!")
		},
	}
	return cmd
}

func newMangaUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update ComicInfo.xml in manga archives",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement manga updates
			cmd.Println("ComicInfo.xml support coming soon!")
		},
	}
	return cmd
}

func newMangaGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate mangaInfo.xml for manga archives",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement manga generation
			cmd.Println("ComicInfo.xml generation coming soon!")
		},
	}
	return cmd
}

// retrieveMangadexIdCmd searches Mangadex for a title and prints the results.
func retrieveMangadexIdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [title words...]",
		Short: "Search Mangadex for a title and return all matches (title name and Mangadex ID)",
		Args:  cobra.MinimumNArgs(1), // Require at least one positional arg
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

func retrieveMangaMetadataCmd(MangadexId string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metadata",
		Short: "Query Mangadex metatdata to populate the ComicInfo.xml",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement manga generation
			cmd.Println("Retrieve metadata from Mangadex!")
			titleMetadata, metadataErr := mangasrc.TitleMetadata(MangadexId)
			if metadataErr != nil {
				log.Fatalf("could not retrieve metadata for id: %s, %v", MangadexId, metadataErr)
			}
			parser.PrintMangaDexMetadata(titleMetadata)

		},
	}
	return cmd
}
