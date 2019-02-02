// +build dev

package client

import "net/http"

// Assets contains project assets.
var Dist http.FileSystem = http.Dir(pathToData)
