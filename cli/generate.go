package cli

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mdu/parser"
)

func NewGenerateCmd() *cobra.Command {
	var output, format string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate example input files",
		Long: `Generate example input files for batch metadata updates.
Supports multiple formats (EPUB, PDF, ComicInfo).
Output format (JSON/YAML) is determined by file extension.`,
		Example: `  mdu generate --output epub-example.json --format epub
  mdu generate --output comic-example.yaml --format comic
  mdu generate --output metadata.yml --format epub`,
		Run: func(cmd *cobra.Command, args []string) {
			if output == "" {
				log.Fatal("Error: You must specify --output filename")
			}

			ext := strings.ToLower(filepath.Ext(output))
			var err error

			switch ext {
			case ".json":
				err = generateJSONExample(format, output)
			case ".yaml", ".yml":
				err = generateYAMLExample(format, output)
			default:
				log.Fatal("Error: Output file must have .json, .yaml, or .yml extension")
			}

			if err != nil {
				log.Fatalf("Error generating example file: %v", err)
			}

			cmd.Printf("✓ Example %s input file created: %s\n", format, output)
			cmd.Println("\nYou can now edit this file and use it with:")
			cmd.Printf("  mdu %s update --file yourfile --input %s\n", format, output)
			cmd.Printf("  mdu %s update --dir ./files --input %s\n", format, output)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output filename (.json, .yaml, or .yml)")
	cmd.Flags().StringVar(&format, "format", "epub", "Format type (epub, pdf, comic)")
	cmd.MarkFlagRequired("output")

	return cmd
}

func generateJSONExample(format, output string) error {
	switch format {
	case "epub":
		return parser.GenerateExampleJSON(output)
	case "pdf":
		// TODO: Implement PDF JSON generation
		return parser.GenerateExampleJSON(output) // Placeholder
	case "comic":
		// TODO: Implement ComicInfo JSON generation
		return parser.GenerateExampleJSON(output) // Placeholder
	default:
		log.Fatalf("Unknown format: %s", format)
		return nil
	}
}

func generateYAMLExample(format, output string) error {
	switch format {
	case "epub":
		return parser.GenerateExampleYAML(output)
	case "pdf":
		// TODO: Implement PDF YAML generation
		return parser.GenerateExampleYAML(output) // Placeholder
	case "comic":
		// TODO: Implement ComicInfo YAML generation
		return parser.GenerateExampleYAML(output) // Placeholder
	default:
		log.Fatalf("Unknown format: %s", format)
		return nil
	}
}
