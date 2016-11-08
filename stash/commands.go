package stash

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Stash stashes a file or directory by wrapping it into a compressed tar archive.
func Stash(source string, destination string) error {
	if source == "" || destination == "" {
		return fmt.Errorf("\"%v\" or \"%v\" is not a valid path", source, destination)
	}

	//NOTE: Destination must always be a directory.

	absPath, err := filepath.Abs(source)
	if err != nil {
		return err
	}

	hasher := sha1.New()
	log.Printf("Using path: %v\n", absPath)
	file, err := os.Create(filepath.Join(destination, encodeChecksum(computeChecksum(absPath, hasher), base64.URLEncoding)) + ".tar.gz")
	if err != nil {
		return err
	}

	if err := Pack(source, file); err != nil {
		return err
	}

	return nil
}

// Release gets a file or directory from the stash and writes it to destination.
func Release(source string, target string, destination string) error {
	if source == "" || destination == "" || target == "" {
		return fmt.Errorf("\"%v\" or \"%v\" or \"%v\" is not a valid path", source, destination, target)
	}

	absPath, err := filepath.Abs(target)
	if err != nil {
		return err
	}

	hasher := sha1.New()
	log.Printf("Using path: %v\n", absPath)
	file, err := os.Open(filepath.Join(source, encodeChecksum(computeChecksum(absPath, hasher), base64.URLEncoding)) + ".tar.gz")
	if err != nil {
		return err
	}

	if err := Unpack(destination, file); err != nil {
		return err
	}

	return nil
}
