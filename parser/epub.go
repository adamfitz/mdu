package parser

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"sort"
)

// Checks the magic bytes of a file to determine if it is an EPUB.
func IsEpub(filePath string) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// EPUB files are ZIP archives starting with "PK\x03\x04"
	header := make([]byte, 4)
	if _, err := f.Read(header); err != nil {
		return false, err
	}
	if !bytes.Equal(header, []byte("PK\x03\x04")) {
		return false, nil
	}

	// Check for "mimetype" file inside the ZIP
	stat, err := f.Stat()
	if err != nil {
		return false, err
	}

	zr, err := zip.NewReader(f, stat.Size())
	if err != nil {
		return false, err
	}

	for _, file := range zr.File {
		if file.Name == "mimetype" {
			rc, err := file.Open()
			if err != nil {
				return false, err
			}
			defer rc.Close()

			buf := make([]byte, file.UncompressedSize64)
			if _, err := rc.Read(buf); err != nil {
				return false, err
			}
			return string(buf) == "application/epub+zip", nil
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
