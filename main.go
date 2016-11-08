package main

import (
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/zillolo/stash/stash"
)

const (
	ArchiveExtension = ".tar.gz"
	DataStoragePath  = "/.local/share/hidden"
	DatabaseFile     = "index.db"
)

var restoreFlag = flag.Bool("restore", false, "Show a hidden file.")

func main() {
	flag.Parse()
	InitEnvironment()

	db, err := InitDB()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	log.Printf("File: %v\n", flag.Arg(0))
	if *restoreFlag {
		if err := Restore(flag.Arg(0)); err != nil {
			log.Fatalf("%v\n", err)
		}
	} else {
		if err := Hide(db, flag.Arg(0)); err != nil {
			log.Fatalf("%v\n", err)
		}
	}
}
func InitEnvironment() {
	path := path.Join(getHome(), DataStoragePath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.Mkdir(path, 0755|os.ModeDir); err != nil {
			panic(err)
		}
	}
}

func InitDB() (db *gorm.DB, err error) {
	db, err = gorm.Open("sqlite3", path.Join(getHome(), DataStoragePath, DatabaseFile))
	if err != nil {
		return nil, err
	}

	if !db.HasTable(&stash.Entry{}) {
		db.CreateTable(&stash.Entry{})
	}
	return db, nil
}

func Hide(db *gorm.DB, root string) (err error) {
	// Compute the filename for the archive.
	name := path.Join(getDataPath(), computeFilename(root))
	name += ArchiveExtension

	// Create the file for the archive.
	file, err := os.Create(name)
	if err != nil {
		return err
	}

	if err := stash.Pack(root, file); err != nil {
		return err
	}

	// Add entry to the database.
	absPath, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	entry := stash.Entry{Path: absPath}
	db.Create(&entry)
	return nil
}

func Restore(path string) (err error) {
	name := filepath.Join(getDataPath(), computeFilename(path))
	name += ArchiveExtension

	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}

	if err := stash.Unpack(filepath.Dir(path), file); err != nil {
		return err
	}
	return nil
}

func getHome() string {
	homePath := os.Getenv("HOME")
	if homePath == "" {
		panic("Hide may not be used by non-humans.")
	}
	return homePath
}

func getDataPath() string {
	return path.Join(getHome(), DataStoragePath)
}

func computeFilename(path string) string {
	hasher := sha1.New()
	hasher.Write([]byte(path))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}
