package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/zillolo/stash/stash"
)

const (
	ArchiveExtension = ".tar.gz"
	DataStoragePath  = "/.local/share/hidden"
	DatabaseFile     = "index.db"
)

func main() {
	flag.Parse()

	InitEnvironment()
	InitLogging()

	db, err := InitDB()
	if err != nil {
		fmt.Printf("There was an error creating the database.\nMore info: %v\n", err)
		log.Printf("%v\n", err)
	}

	if len(flag.Args()) < 2 {
		flag.Usage()
	}

	source := flag.Arg(0)
	switch strings.ToLower(flag.Arg(0)) {
	case "list":
		source = "."
		if len(flag.Args()) >= 2 {
			source = flag.Arg(1)
		}
		if err := stash.List(db, source); err != nil {
			fmt.Printf("Could not list entries.\nMore info: %v\n", err)
			log.Printf("%v\n", err)
		}
	case "release":
		source = "/home/alex/.local/share/hidden"
		target := flag.Arg(1)
		destination := "."

		if len(flag.Args()) >= 3 {
			destination = flag.Arg(2)
		}
		if err := stash.Release(db, source, target, destination); err != nil {
			fmt.Printf("Could not release.\nMore info: %v\n", err)
			log.Printf("%v\n", err)
		}
	case "stash":
		source = flag.Arg(1)
		fallthrough
	default:
		if err := stash.Stash(db, source, "/home/alex/.local/share/hidden"); err != nil {
			fmt.Printf("Could not stash files.\nMore info: %v\n", err)
			log.Printf("%v\n", err)
		}
	}
}

func InitLogging() {
	file, err := os.OpenFile("stash.log", os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		//TODO: Handle error here.
	}
	defer file.Close()

	log.SetOutput(file)
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

//func Restore(path string) (err error) {
//	name := filepath.Join(getDataPath(), computeFilename(path))
//	name += ArchiveExtension
//
//	file, err := os.Open(name)
//	if err != nil {
//		return err
//	}
//	defer file.Close()
//
//	path, err = filepath.Abs(path)
//	if err != nil {
//		return err
//	}
//
//	if err := stash.Unpack(filepath.Dir(path), file); err != nil {
//		return err
//	}
//	return nil
//}

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
