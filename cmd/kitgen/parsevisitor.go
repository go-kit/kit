package main

import (
	"go/ast"

	"github.com/davecgh/go-spew/spew"
)

type (
	parseVisitor struct {
		src *sourceContext
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

func (v *parseVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case *ast.File:
		v.src.pkg = rn.Name
		return v
	case *ast.ImportSpec:
		v.src.imports = append(v.src.imports, rn)
		return nil

	case *ast.TypeSpec:
		return &typeSpecVisitor{src: v.src, node: rn}
	}
}

/*
package foo

type FooService interface {
	Bar(ctx context.Context, i int, s string) (string, error)
}
*/

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
			v.iface.stubname = &(*v.name)
			v.iface.stubname.Name = "stub" + v.name.String()
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
			spew.Dump("results")
			v.results = &[]arg{}
		}
		return &argListVisitor{list: v.results}
	case nil:
		spew.Dump("done", v)
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
		spew.Dump(v.list)
		for _, n := range names {
			spew.Dump(n)
			*v.list = append(*v.list, arg{
				name: n.(*ast.Ident),
				typ:  tp,
			})
		}
		spew.Dump(v.list)
	}
	return nil
}
