//go:generate go run -tags=dev dist_generate.go
package client

import (
	"os"
	"path/filepath"
)

var root, _ = os.Getwd()
var pathToData = filepath.Join(root, "dist")
