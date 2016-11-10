package stash

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jinzhu/gorm"
)

// Stash stashes a file or directory by wrapping it into a compressed tar archive.
func Stash(db *gorm.DB, source string, destination string) error {
	if source == "" || destination == "" {
		return fmt.Errorf("\"%v\" or \"%v\" is not a valid path", source, destination)
	}

	if db == nil {
		return fmt.Errorf("DB was nil")
	}

	//NOTE: Destination must always be a directory.

	absPath, err := filepath.Abs(source)
	if err != nil {
		return err
	}

	entry := &Entry{}
	db.Where("path = ?", absPath).First(entry)
	if entry.ID != 0 {
		return errors.New("Entry exist already in database")
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

	entry = &Entry{Path: absPath}
	db.Create(entry)

	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := os.Chdir(filepath.Dir(absPath)); err != nil {
		return err
	}
	if err := os.RemoveAll(source); err != nil {
		return err
	}
	if err := os.Chdir(workingDir); err != nil {
		return err
	}

	return nil
}

// Release gets a file or directory from the stash and writes it to destination.
func Release(db *gorm.DB, source string, target string, destination string) error {
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

	db.Where("path = ?", absPath).Delete(&Entry{})

	return nil
}

func List(db *gorm.DB, source string) error {
	if db == nil {
		return fmt.Errorf("Db was nil")
	}

	absPath, err := filepath.Abs(source)
	if err != nil {
		return err
	}

	entries := []Entry{}
	db.Where("path LIKE ?", fmt.Sprintf("%s%%", absPath)).Find(&entries)

	for _, entry := range entries {
		fmt.Printf("%v\n", entry.Path)
	}
	return nil
}

func Init(path string) error {
	if path == "" {
		return fmt.Errorf("You have specified an invalid directory")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("%v\n", err)
		return fmt.Errorf("Can not find absolute path of \"%v\"", path)
	}

	if err := os.MkdirAll(absPath, 0644); err != nil {
		log.Printf("%v\n", err)
		return fmt.Errorf("Can not create directory")
	}

	return nil
}
