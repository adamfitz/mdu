package metadata

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// Supported PDF metadata fields for filtering/updating
var PdfSupportedFields = []string{
	"title",
	"author",
	"summary",
	"keywords",
	"creator",
	"calibre:series",
	"calibreSI:series_index",
}

// ---------- Public PDF Functions ----------

// PdfRead reads metadata from a PDF file and returns a map of fields.
func PdfRead(pdfPath string, all bool) (map[string]string, error) {
	ctx, err := api.ReadContextFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF context: %v", err)
	}

	if ctx.XRefTable.Info == nil {
		return nil, nil
	}

	dict, err := ctx.DereferenceDict(*ctx.XRefTable.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to dereference Info dict: %v", err)
	}

	result := make(map[string]string)
	for k, v := range dict {
		if s, ok := v.(types.StringLiteral); ok {
			key := strings.ToLower(strings.TrimSpace(k))
			val := strings.TrimSpace(s.Value())
			result[key] = val
		}
	}

	if !all {
		filtered := make(map[string]string)
		for _, f := range PdfSupportedFields {
			if val, ok := result[f]; ok {
				filtered[f] = val
			}
		}
		result = filtered
	}

	// Sort keys alphabetically
	sortedResult := make(map[string]string)
	keys := make([]string, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sortedResult[k] = result[k]
	}

	return sortedResult, nil
}

// PdfUpdate writes metadata updates to a PDF file.
func PdfUpdate(pdfPath, updatedPath string, updates map[string]string) error {
	// Copy original file
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to read original PDF: %v", err)
	}
	if err := os.WriteFile(updatedPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write updated PDF: %v", err)
	}

	ctx, err := api.ReadContextFile(updatedPath)
	if err != nil {
		return fmt.Errorf("failed to read PDF context: %v", err)
	}

	var dict types.Dict
	if ctx.XRefTable.Info == nil {
		dict = types.NewDict()
		ir, err := ctx.IndRefForNewObject(dict)
		if err != nil {
			return fmt.Errorf("failed to create new Info dict: %v", err)
		}
		ctx.XRefTable.Info = ir
	} else {
		dict, err = ctx.DereferenceDict(*ctx.XRefTable.Info)
		if err != nil {
			return fmt.Errorf("failed to dereference Info dict: %v", err)
		}
	}

	for k, v := range updates {
		dict[k] = types.StringLiteral(v)
	}

	if err := api.WriteContextFile(ctx, updatedPath); err != nil {
		return fmt.Errorf("failed to write updated PDF: %v", err)
	}

	return nil
}
