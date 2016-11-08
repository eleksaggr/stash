package main

import (
	"flag"
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

	_, err := InitDB()
	if err != nil {
		log.Panicf("%v\n", err)
	}

	if len(flag.Args()) < 2 {
		flag.Usage()
	}

	source := flag.Arg(0)
	switch strings.ToLower(flag.Arg(0)) {
	case "release":
		source = "/home/alex/.local/share/hidden"
		target := flag.Arg(1)
		destination := "."

		if len(flag.Args()) >= 3 {
			destination = flag.Arg(2)
		}
		if err := stash.Release(source, target, destination); err != nil {
			log.Panicf("%v\n", err)
		}
	case "stash":
		source = flag.Arg(1)
		fallthrough
	default:
		if err := stash.Stash(source, "/home/alex/.local/share/hidden"); err != nil {
			log.Panicf("%v\n", err)
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
