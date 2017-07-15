package main

import "go/ast"

type (
	sourceContext struct {
		ifaceName  string
		ifaceCount int
	}

	typeSpecVisitor struct {
		src  *sourceContext
		node *ast.TypeSpec
	}

	interfaceTypeVisitor struct {
		src *sourceContext
	}
)

func (sc *sourceContext) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return sc
	case *ast.TypeSpec:
		return &typeSpecVisitor{src: sc, node: rn}
	}
}

func (sc *sourceContext) validate() error {
	if sc.ifaceCount != 1 {
		return fmt.Errorf("found %d interfaces, expecting exactly 1", sc.ifaceCount)
	}
	return nil
}

func (sc *sourceContext) foundInterface(name string) {
	sc.ifaceCount++
	if sc.ifaceCount == 1 {
		sc.ifaceName = rn.Name.String()
	}
}

func (v *typeSpecVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case *ast.InterfaceType:
		v.src.foundInterface(v.node.Name)
		return &interfaceTypeVisitor{src: v.src}
	}
}

func (v *interfaceTypeVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case ast.Field:
		spew.Dump(rn)
	}
}
