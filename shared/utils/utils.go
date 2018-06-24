package utils

import (
	"os"
	"path/filepath"

	"github.com/ryex/go-broadcaster/shared/logutils"
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
		logutils.Log.Error("could not get an absolute path for", root)
		return aerr
	}
	logutils.Log.Info("Processing Directory... ", rootPath)
	werr := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if path == rootPath {
			return nil
		}
		if info.IsDir() {
			derr := WalkSearch(path, extensions, cb)
			if derr != nil {
				return derr
			}
			return nil
		}
		logutils.Log.Info("Processing File... ", path)
		logutils.Log.Info("File Extension: ", filepath.Ext(path))
		if StringInSlice(filepath.Ext(path), extensions) {
			cerr := cb(path)
			if cerr != nil {
				logutils.Log.Error("error in search callback", path)
				return cerr
			}
		}

		return nil
	})
	if werr != nil {
		logutils.Log.Error("error walking library path", rootPath)
		return werr
	}
	return nil
}
