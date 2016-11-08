package main

import (
	"crypto/sha1"
	"encoding/base64"
	"log"
	"os"
	"path"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const (
	ArchiveExtension = ".tar.gz"
	DataStoragePath  = "/.local/share/hidden"
	DatabaseFile     = "index.db"
)

func main() {
	InitEnvironment()

	db, err := InitDB()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	if err := Hide(db, os.Args[1]); err != nil {
		log.Fatalf("%v\n", err)
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

	if !db.HasTable(&Entry{}) {
		db.CreateTable(&Entry{})
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

	if err := Pack(root, file); err != nil {
		return err
	}

	// Add entry to the database.
	entry := Entry{Path: root}
	db.Create(&entry)
	return nil
}

func Show(path string) (err error) {
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
