package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// MetadataInput represents the structure of metadata updates from input file
type MetadataInput struct {
	Author      string `json:"author" yaml:"author"`
	Summary     string `json:"summary" yaml:"summary"`
	ISBN        string `json:"isbn" yaml:"isbn"`
	Series      string `json:"series" yaml:"series"`
	SeriesIndex string `json:"series_index" yaml:"series_index"`

	// Allow arbitrary additional fields
	Additional map[string]string `json:"additional,omitempty" yaml:"additional,omitempty"`
}

// ParseInputFile reads a JSON or YAML file and returns metadata updates
func ParseInputFile(filePath string) (map[string]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	var input MetadataInput

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &input); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file format: %s (use .json, .yaml, or .yml)", ext)
	}

	return inputToMap(input), nil
}

// inputToMap converts MetadataInput struct to map[string]string
func inputToMap(input MetadataInput) map[string]string {
	updates := make(map[string]string)

	if input.Author != "" {
		updates["author"] = input.Author
	}
	if input.Summary != "" {
		updates["summary"] = input.Summary
	}
	if input.ISBN != "" {
		updates["isbn"] = input.ISBN
	}
	if input.Series != "" {
		updates["calibre:series"] = input.Series
	}
	if input.SeriesIndex != "" {
		updates["calibre:series_index"] = input.SeriesIndex
	}

	// Add any additional fields
	for key, value := range input.Additional {
		if value != "" {
			updates[key] = value
		}
	}

	return updates
}

// ValidateInputFile checks if an input file is valid without applying updates
func ValidateInputFile(filePath string) error {
	updates, err := ParseInputFile(filePath)
	if err != nil {
		return err
	}

	if len(updates) == 0 {
		return fmt.Errorf("input file contains no metadata updates")
	}

	return nil
}

// GenerateExampleJSON creates an example JSON input file
func GenerateExampleJSON(outputPath string) error {
	example := MetadataInput{
		Author:      "Author Name",
		Summary:     "Book description or summary",
		ISBN:        "978-1234567890",
		Series:      "Series Name",
		SeriesIndex: "1",
		Additional: map[string]string{
			"publisher": "Publisher Name",
			"language":  "en",
		},
	}

	data, err := json.MarshalIndent(example, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

// GenerateExampleYAML creates an example YAML input file
func GenerateExampleYAML(outputPath string) error {
	yamlContent := `# EPUB Metadata Update Template
# All fields are optional - only include fields you want to update

# Standard metadata fields
author: "Author Name"
summary: "Book description or summary"
isbn: "978-1234567890"

# Series information (for Kavita)
series: "Series Name"
series_index: "1"

# Additional metadata fields (optional)
additional:
  publisher: "Publisher Name"
  language: "en-US"
  # Add any other custom fields here
`

	return os.WriteFile(outputPath, []byte(yamlContent), 0644)
}
