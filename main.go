package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Needed for GORM
	homedir "github.com/mitchellh/go-homedir"
)

var errNoLog = errors.New("Could not enable logging. Redirecting log output to stdout.")

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		return
	}

	if strings.ToLower(flag.Arg(0)) == "init" {
		if len(flag.Args()) < 2 {
			flag.Usage()
			return
		}
		target := flag.Arg(1)

		if err := Init(target); err != nil {
			fmt.Printf("%v.\n", err)
		}
		return
	}

	homePath, err := homedir.Dir()
	if err != nil {
		fmt.Printf("Can not retrieve home path for current user.\n")
		return
	}

	configPath := filepath.Join(homePath, ".config/stash/stash.conf")
	config, err := readConfig(configPath)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	if err := initLogging(config.LogDir); err != nil {
		fmt.Printf("%v\n", errNoLog)
	}

	dbPath := filepath.Join(config.DataDir, "index.db")
	db, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Main: %v at path %v\n", err, dbPath)
		fmt.Printf("Can not connect to database.\n")
		return
	}

	source := flag.Arg(0)
	switch strings.ToLower(flag.Arg(0)) {
	case "list":
		source = "."
		if len(flag.Args()) >= 2 {
			source = flag.Arg(1)
		}
		if err := List(db, source); err != nil {
			fmt.Printf("Could not list entries.\nMore info: %v\n", err)
			log.Printf("%v\n", err)
		}
	case "release":
		source = config.DataDir
		target := flag.Arg(1)
		destination := "."

		if len(flag.Args()) >= 3 {
			destination = flag.Arg(2)
		}
		if err := Release(db, source, target, destination); err != nil {
			fmt.Printf("Could not release.\nMore info: %v\n", err)
			log.Printf("%v\n", err)
		}
	case "stash":
		source = flag.Arg(1)
		fallthrough
	default:
		if err := Stash(db, source, config.DataDir); err != nil {
			fmt.Printf("Could not stash files.\nMore info: %v\n", err)
			log.Printf("%v\n", err)
		}
	}
}

func readConfig(path string) (*Config, error) {
	config := new(Config)
	if _, err := toml.DecodeFile(path, config); err != nil {
		return nil, fmt.Errorf("Could not read configuration file. Did you call \"stash init\"?")
	}
	return config, nil
}

func initLogging(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	path = filepath.Join(absPath, string(time.Now().Local().Format("2006-02-01")))
	path += ".log"

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("Can not create log file.")
	}

	log.SetOutput(file)
	log.Printf("Started logging to file %v\n", absPath)
	return nil
}
