package main

type transformer struct {
	src *sourceContext
}

func (v *transformer) Visit(ast.Node) ast.Visitor {
	return v
}
