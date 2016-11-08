package stash

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

// Stash stashes a file or directory by wrapping it into a compressed tar archive.
func Stash(source string, destination string) error {
	if source == "" || destination == "" {
		return fmt.Errorf("\"%v\" or \"%v\" is not a valid path", source, destination)
	}

	absPath, err := filepath.Abs(source)
	if err != nil {
		return err
	}

	hasher := sha1.New()
	file, err := os.Create(destination + encodeChecksum(computeChecksum(absPath, hasher), base64.URLEncoding) + ".tar.gz")
	if err != nil {
		return err
	}

	if err := Pack(source, file); err != nil {
		return err
	}

	return nil
}
