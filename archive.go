package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	p "path"
	"path/filepath"
)

func Pack(source string, writers ...io.Writer) error {
	multiWriter := io.MultiWriter(writers...)
	gzipWriter := gzip.NewWriter(multiWriter)
	tarWriter := tar.NewWriter(gzipWriter)

	defer gzipWriter.Close()
	defer tarWriter.Close()

	if err := filepath.Walk(source, addFile(p.Dir(source), tarWriter)); err != nil {
		return err
	}
	return nil
}

func addFile(root string, w *tar.Writer) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		log.Printf("Processing %v\n", path)
		// If we get passed an error, return it immediately.
		if err != nil {
			return err
		}

		// Create header for the current file.
		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}

		// Construct relative file path to keep structure in the tar archive.
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		header.Name = relPath
		log.Printf("Saving in archive with path: %v\n", header.Name)

		if err := w.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			// If we are processing a file (as opposed to a directory) write it's content to the archive.
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// Copy the contents of the file to the archive.
			if _, err := io.Copy(w, file); err != nil {
				return err
			}

		}
		return nil
	}
}
