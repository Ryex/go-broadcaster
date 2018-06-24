package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

type SearchFunc func(path string) error

func WalkSearch(root string, extensions []string, cb SearchFunc) error {
	rootPath, aerr := filepath.Abs(root)
	if aerr != nil {
		fmt.Println("could not get an absolute path")
		return aerr
	}

	werr := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			derr := WalkSearch(path, extensions, cb)
			if derr != nil {
				return derr
			}
		}
		if StringInSlice(filepath.Ext(path), extensions) {
			cerr := cb(path)
			if cerr != nil {
				fmt.Println("error in search callback")
				return cerr
			}
		}

		return nil
	})
	if werr != nil {
		fmt.Println("error walking library path")
		return werr
	}
	return nil
}
