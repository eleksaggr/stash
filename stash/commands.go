package stash

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Needed for GORM
	homedir "github.com/mitchellh/go-homedir"
)

const (
	confFile = "stash.conf"
	dbFile   = "index.db"

	logDir = ".cache/stash/logs"
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

// List lists all stashed objects that match the source path.
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

// Init initalizes the environment by creating directories, databases and configuration files
// needed by the application.
func Init(path string) error {
	if path == "" {
		return fmt.Errorf("You have specified an invalid directory")
	}

	// Get the absolute path to our target directory and create any directories that are missing.
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("%v\n", err)
		return fmt.Errorf("Can not find absolute path of \"%v\"", path)
	}

	if err := os.MkdirAll(absPath, 0644); err != nil {
		log.Printf("%v\n", err)
		return fmt.Errorf("Can not create directory")
	}

	// Create the database in the newly created directory.
	dbPath := filepath.Join(absPath, dbFile)
	db, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Init: %v\n", err)
		return fmt.Errorf("Can not create database file")
	}

	if !db.HasTable(&Entry{}) {
		db.CreateTable(&Entry{})
	}

	// Create the log file directory
	homePath, err := homedir.Dir()
	logPath := filepath.Join(homePath, logDir)
	if err != nil {
		log.Printf("Init: %v\n", err)
		return fmt.Errorf("Can not retrieve home directory for current user")
	}

	confPath := filepath.Join(absPath, confFile)
	config := Config{
		DataDir:  absPath,
		ConfPath: confPath,
		LogDir:   logPath,
	}

	var buffer bytes.Buffer
	encoder := toml.NewEncoder(&buffer)
	if err := encoder.Encode(&config); err != nil {
		log.Printf("Init: %v\n", err)
		return fmt.Errorf("Could not encode configuration file to TOML.")
	}

	if err := ioutil.WriteFile(confPath, buffer.Bytes(), 0644); err != nil {
		log.Printf("Init: %v\n", err)
		return fmt.Errorf("Can not write to configuration file")
	}

	return nil
}
