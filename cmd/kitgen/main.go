package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
)

// go get github.com/nyarly/inlinefiles
//go:generate inlinefiles --vfs=ASTTemplates --glob=* ./templates ast_templates.go

func usage() string {
	return fmt.Sprintf("Usage: %s <filename>", os.Args[0])
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal(usage())
	}
	filename := os.Args[1]
	buf, err := process(filename)
	if err != nil {
		log.Fatal(err)
	}

	io.Copy(os.Stdout, buf)
}

func process(filename string) (io.Reader, error) {
	f, err := parseFile("main.go")
	if err != nil {
		return nil, err
	}

	context, err := extractContext(f)
	if err != nil {
		return nil, err
	}

	outAST, err := transformAST(context)
	if err != nil {
		return nil, err
	}

	buf, err := formatNode(outAST)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func parseFile(fname string) (ast.Node, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fname, nil, parser.DeclarationErrors)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func extractContext(f ast.Node) (*sourceContext, error) {
	context := &sourceContext{}

	ast.Walk(context, f)

	return context, context.validate()
}

func transformAST(ctx *sourceContext) (ast.Node, error) {
}

func formatNode(node ast.Node) (*bytes.Buffer, error) {
	outfset := token.NewFileSet()
	buf := &bytes.Buffer{}
	err := format.Node(buf, outfset, node)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
