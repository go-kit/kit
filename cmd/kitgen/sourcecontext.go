package main

import (
	"fmt"
	"go/ast"
)

type (
	sourceContext struct {
		imports    []*ast.ImportSpec
		interfaces []iface
	}

	iface struct {
		name    *ast.Ident
		methods []method
	}

	method struct {
		name    *ast.Ident
		params  []arg
		results []arg
	}

	arg struct {
		name *ast.Ident
		typ  ast.Expr
	}

	typeSpecVisitor struct {
		src   *sourceContext
		node  *ast.TypeSpec
		iface *iface
		name  *ast.Ident
	}

	interfaceTypeVisitor struct {
		node    *ast.TypeSpec
		ts      *typeSpecVisitor
		methods []method
	}

	methodVisitor struct {
		node            *ast.TypeSpec
		list            *[]method
		name            *ast.Ident
		params, results *[]arg
		isMethod        bool
	}

	argListVisitor struct {
		list *[]arg
	}

	argVisitor struct {
		node  *ast.TypeSpec
		parts []ast.Expr
		list  *[]arg
	}
)

func (sc *sourceContext) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return sc
	case *ast.ImportSpec:
		sc.imports = append(sc.imports, rn)
		return nil

	case *ast.TypeSpec:
		return &typeSpecVisitor{src: sc, node: rn}
	}
}

func (sc *sourceContext) validate() error {
	if len(sc.interfaces) != 1 {
		return fmt.Errorf("found %d interfaces, expecting exactly 1", len(sc.interfaces))
	}
	return nil
}

func (v *typeSpecVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case *ast.Ident:
		if v.name == nil {
			v.name = rn
		}
		return v
	case *ast.InterfaceType:
		return &interfaceTypeVisitor{ts: v, methods: []method{}}
	case nil:
		if v.iface != nil {
			v.iface.name = v.name
			v.src.interfaces = append(v.src.interfaces, *v.iface)
		}
		return nil
	}
}

func (v *interfaceTypeVisitor) Visit(n ast.Node) ast.Visitor {
	switch n.(type) {
	default:
		return v
	case *ast.Field:
		return &methodVisitor{list: &v.methods}
	case nil:
		v.ts.iface = &iface{methods: v.methods}
		return nil
	}
}

func (v *methodVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case *ast.Ident:
		if rn.IsExported() {
			v.name = rn
		}
		return v
	case *ast.FuncType:

		v.isMethod = true
		return v
	case *ast.FieldList:
		if v.params == nil {
			v.params = &[]arg{}
			return &argListVisitor{list: v.params}
		}
		if v.results == nil {
			v.results = &[]arg{}
		}
		return &argListVisitor{list: v.results}
	case nil:
		if v.isMethod && v.name != nil {
			*v.list = append(*v.list, method{name: v.name, params: *v.params, results: *v.results})
		}
		return nil
	}
}

func (v *argListVisitor) Visit(n ast.Node) ast.Visitor {
	switch n.(type) {
	default:
		return nil
	case *ast.Field:
		return &argVisitor{list: v.list}
	}
}

func (v *argVisitor) Visit(n ast.Node) ast.Visitor {
	switch t := n.(type) {
	case *ast.CommentGroup, *ast.BasicLit:
		return nil
	case *ast.Ident: //Expr -> everything, but clarity
		v.parts = append(v.parts, t)
	case ast.Expr:
		v.parts = append(v.parts, t)
	case nil:
		names := v.parts[:len(v.parts)-1]
		tp := v.parts[len(v.parts)-1]
		for _, n := range names {
			*v.list = append(*v.list, arg{
				name: n.(*ast.Ident),
				typ:  tp,
			})
		}
	}
	return nil
}
