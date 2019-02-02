// +build dev

package filesystem

import "net/http"

// Assets contains project assets.
var Dist http.FileSystem = http.Dir(PathToData)
