// +build ignore

package main

import (
	"log"

	"github.com/ryex/go-broadcaster/web/filesystem"

	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(filesystem.Dist, vfsgen.Options{
		PackageName:  "filesystem",
		BuildTags:    "!dev",
		VariableName: "Dist",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
