// +build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"gocloud.dev/blob"
)

func main() {
	opts := []entc.Option{
		entc.Dependency(
			entc.DependencyName("Bucket"),
			entc.DependencyType(&blob.Bucket{}),
		),
	}
	if err := entc.Generate("./schema", &gen.Config{}, opts...); err != nil {
		log.Fatal("running ent codegen:", err)
	}
}
