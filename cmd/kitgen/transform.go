package main

import "go/ast"

type transformer struct {
	src *sourceContext
}

func (v *transformer) Visit(ast.Node) ast.Visitor {
	return v
}

func (v *transformer) err() error {
	return nil
}
