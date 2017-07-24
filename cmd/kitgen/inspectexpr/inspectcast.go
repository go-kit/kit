package main

import (
	"go/ast"
	"go/parser"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	n, err := parser.ParseExpr(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	ast.Inspect(n, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		spew.Dump(n)
		return true
	})
}
