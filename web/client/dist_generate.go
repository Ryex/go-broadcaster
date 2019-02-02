// +build ignore

package main

import (
	"log"

	"github.com/ryex/go-broadcaster/web/client"

	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(client.Dist, vfsgen.Options{
		PackageName:  "client",
		BuildTags:    "!dev",
		VariableName: "Dist",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
