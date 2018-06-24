package models

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ryex/go-broadcaster/shared/utils"
)

type LibraryPath struct {
	Path      string
	Added     time.Time
	LastIndex time.Time
	Indexing  bool
}

func (fp LibraryPath) SearchWalk(extensions []string, cb utils.SearchFunc) error {
	rootPath, aerr := filepath.Abs(fp.Path)
	if aerr != nil {
		fmt.Println("could not get an absolute path")
		return aerr
	}

	werr := utils.WalkSearch(rootPath, extensions, cb)
	if werr != nil {
		fmt.Println("error in search walk")
		return werr
	}
	return nil
}
