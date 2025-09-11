package parser

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// Checks if a file is a valid EPUB by inspecting its ZIP header and mimetype.
func IsEpub(filePath string) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Check ZIP magic bytes
	header := make([]byte, 4)
	if _, err := f.ReadAt(header, 0); err != nil {
		return false, err
	}
	if !bytes.Equal(header, []byte("PK\x03\x04")) {
		return false, nil
	}

	// Get file size
	stat, err := f.Stat()
	if err != nil {
		return false, err
	}

	// Create zip reader
	zr, err := zip.NewReader(f, stat.Size())
	if err != nil {
		return false, err
	}

	// Look for "mimetype" file
	for _, file := range zr.File {
		if file.Name == "mimetype" {
			rc, err := file.Open()
			if err != nil {
				return false, err
			}
			defer rc.Close()

			buf := make([]byte, 20)
			n, err := io.ReadFull(rc, buf)
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				return false, err
			}
			return string(buf[:n]) == "application/epub+zip", nil
		}
	}

	return false, nil
}

// ListEpubFiles returns all valid EPUB files in the given directory, sorted alphabetically.
func ListEpubFiles(dir string) ([]string, error) {
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
		ok, err := IsEpub(path)
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
