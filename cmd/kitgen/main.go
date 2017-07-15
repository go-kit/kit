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

	context := &sourceContext{}

	ast.Walk(context, f)

	buf, err := formatNode(f)
	if err != nil {
		log.Fatal(err)
	}

}

func (s *stripPosVisitor) Visit(n ast.Node) ast.Visitor {
	switch pd := n.(type) {
	case *ast.BlockStmt:
		pd.Lbrace = token.NoPos
		pd.Rbrace = token.NoPos
	}

	return s
}

func parseFile(fname string) (ast.Node, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fname, nil, parser.DeclarationErrors)
	if err != nil {
		return nil, err
	}
	return f, nil
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
