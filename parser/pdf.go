package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
)

// isPDF checks the magic bytes of a file to determine if it is a PDF.
func IsPDF(filePath string) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	header := make([]byte, 5)
	if _, err := f.Read(header); err != nil {
		return false, err
	}

	// PDF files start with "%PDF-"
	return bytes.Equal(header, []byte("%PDF-")), nil
}

// ListPListPdfFilesDFFiles returns all valid PDF files in the given directory, sorted alphabetically.
func ListPdfFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		ok, err := IsPDF(path)
		if err != nil {
			continue
		}
		if ok {
			files = append(files, path)
		}
	}

	sort.Strings(files)
	return files, nil
}
