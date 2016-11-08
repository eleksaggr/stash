package stash

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

		if err := w.WriteHeader(header); err != nil {
			return err
		}

		//TODO: Revert this logic, ie. check for isDir and return
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

func Unpack(destination string, r io.Reader) error {
	log.Printf("Extracting to: %v\n", destination)
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	if err != nil {
		return err
	}

	for {
		header, err := tarReader.Next()

		// If we hit EOF, the archive is empty.
		if err == io.EOF {
			return nil
		}

		if err != nil {
			// Skip this file since the header is corrupt.
			log.Printf("Header corrupt or read error. Skipping file...\n")
			log.Panicf("%v\n", err)
			continue
		}

		target := filepath.Join(destination, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); os.IsNotExist(err) {
				if err = os.MkdirAll(target, 0755|os.ModeDir); err != nil {
					log.Printf("Error creating directory during extraction. Skipping...\n")
					log.Panicf("%v\n", err)
					continue
				}
			}
		case tar.TypeReg:
			file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, header.FileInfo().Mode())
			if err != nil {
				log.Printf("Error during file creation. Skipping...\n")
				log.Panicf("%v\n", err)
				continue
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				log.Printf("Error during writing to file. Skipping...\n")
				log.Panicf("%v\n", err)
				continue
			}
		default:
			log.Printf("Wrong header type flag. Skipping...\n")
			continue
		}

	}
}
