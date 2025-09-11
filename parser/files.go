package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Supported metadata fields for update
var SupportedUpdateFields = map[string]bool{
	"calibre:series":       true,
	"calibre:series_index": true,
	"summary":              true,
	"author":               true,
}

// ReadJSONMetadata reads a JSON file and returns a map of metadata fields.
func ReadJsonMetadata(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	return raw, nil
}

// Filters out unsupported fields and returns a clean map.
func ValidateMetadataFields(input map[string]string) (map[string]string, []string) {
	valid := make(map[string]string)
	ignored := []string{}
	for k, v := range input {
		if SupportedUpdateFields[k] {
			valid[k] = v
		} else {
			ignored = append(ignored, k)
		}
	}
	return valid, ignored
}
