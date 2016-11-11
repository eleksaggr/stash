package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/zillolo/stash/stash"
)

const (
	ArchiveExtension = ".tar.gz"
	DataStoragePath  = "/.local/share/hidden"
	DatabaseFile     = "index.db"
)

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		return
	}

	homePath, err := homedir.Dir()
	if err != nil {
		fmt.Printf("Can not retrieve home path for current user.\n")
		return
	}

	configPath := filepath.Join(homePath, "/.local/share/stash/stash.conf")
	config, err := readConfig(configPath)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	source := flag.Arg(0)
	switch strings.ToLower(flag.Arg(0)) {
	case "init":
		if len(flag.Args()) < 2 {
			flag.Usage()
			return
		}
		target := flag.Arg(1)

		if err := stash.Init(target); err != nil {
			fmt.Printf("%v.\n", err)
		}
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

func readConfig(path string) (*stash.Config, error) {
	config := new(stash.Config)
	if _, err := toml.DecodeFile(path, config); err != nil {
		return nil, fmt.Errorf("Could not read configuration file. Did you call \"stash init\"?")
	}
	return config, nil
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
