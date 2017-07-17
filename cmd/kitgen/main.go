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
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("error while opening %q: %v", filename, err)
	}

	buf, err := process(filename, file)
	if err != nil {
		log.Fatalf(err)
	}

	io.Copy(os.Stdout, buf)
}

func process(filename string, source io.Reader) (io.Reader, error) {
	f, err := parseFile(filename, source)
	if err != nil {
		return nil, err
	}

	context, err := extractContext(f)
	if err != nil {
		return nil, err
	}

	dest, err := transformAST(context)
	if err != nil {
		return nil, err
	}

	buf, err := formatNode(dest)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func parseFile(fname string, source io.Reader) (ast.Node, error) {
	f, err := parser.ParseFile(token.NewFileSet(), fname, source, parser.DeclarationErrors)
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
	tmpl, err := ASTTemplates.Open("full.go")
	if err != nil {
		return nil, err
	}

	root, err := parser.ParseExpr(ioutil.ReadAll(tmpl))
	if err != nil {
		return nil, err
	}

	v := &transformer{src: ctx}
	ast.Walk(v, root)
	return root, v.err()
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
