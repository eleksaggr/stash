package stash

import "github.com/jinzhu/gorm"

// Entry represents a file/directory in the database.
type Entry struct {
	gorm.Model

	IsDir bool
	Path  string
}
