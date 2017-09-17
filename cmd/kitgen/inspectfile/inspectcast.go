package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	n, err := parser.ParseFile(token.NewFileSet(), "test", os.Args[1], parser.Trace)
	if err != nil {
		log.Fatal(err)
	}
	ast.Inspect(n, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		return true
	})
}
