package main

import (
	"fmt"
	"go/ast"
)

// A Visitor's Visit method is invoked for each node encountered by walkToReplace.
// If the result visitor w is not nil, walkToReplace visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
type Visitor interface {
	Visit(node ast.Node, replace func(ast.Node)) (w Visitor)
}

// Helper functions for common node lists. They may be empty.

func walkIdentList(v Visitor, list []*ast.Ident) {
	for i, x := range list {
		walkToReplace(v, x, func(r ast.Node) {
			list[i] = r.(*ast.Ident)
		})
	}
}

func walkExprList(v Visitor, list []ast.Expr) {
	for i, x := range list {
		walkToReplace(v, x, func(r ast.Node) {
			list[i] = r.(ast.Expr)
		})
	}
}

func walkStmtList(v Visitor, list []ast.Stmt) {
	for i, x := range list {
		walkToReplace(v, x, func(r ast.Node) {
			list[i] = r.(ast.Stmt)
		})
	}
}

func walkDeclList(v Visitor, list []ast.Decl) {
	for i, x := range list {
		walkToReplace(v, x, func(r ast.Node) {
			list[i] = r.(ast.Decl)
		})
	}
}

// TODO(gri): Investigate if providing a closure to walkToReplace leads to
//            simpler use (and may help eliminate Inspect in turn).

// walkToReplace traverses an AST in depth-first order: It starts by calling
// v.Visit(node); node must not be nil. If the visitor w returned by
// v.Visit(node) is not nil, walkToReplace is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Visit(nil).
//

func WalkReplace(v Visitor, node ast.Node) ast.Node {
	return walkToReplace(v, node, func(r ast.Node) {
		panic("tried to replace root node")
	})
}

func walkToReplace(v Visitor, node ast.Node, replace func(ast.Node)) (replacement ast.Node) {
	repl := func(r ast.Node) {
		replacement = r
		replace(r)
	}

	if replacement != nil {
		return replacement
	}

	v = v.Visit(node, repl)

	// walk children
	// (the order of the cases matches the order
	// of the corresponding node types in ast.go)
	switch n := node.(type) {
	// Comments and fields
	case *ast.Comment, *ast.BadExpr, *ast.Ident, *ast.BasicLit:
		cpy := *n
		return &cpy

	case *ast.CommentGroup:
		cpy := *n
		cpy.List = make([]*ast.Comment, len(n.List))
		copy(cpy.List, n.List)
		if v == nil {
			return &cpy
		}

		for i, c := range cpy.List {
			walkToReplace(v, c, func(r ast.Node) {
				cpy.List[i] = r.(*ast.Comment)
			})
		}
		return &cpy

	case *ast.Field:
		cpy := *n

		if cpy.Doc != nil {
			walkToReplace(v, cpy.Doc, func(r ast.Node) {
				cpy.Doc = r.(*ast.CommentGroup)
			})
		}
		walkIdentList(v, cpy.Names)
		walkToReplace(v, cpy.Type, func(r ast.Node) {
			cpy.Type = r.(ast.Expr)
		})
		if cpy.Tag != nil {
			walkToReplace(v, cpy.Tag, func(r ast.Node) {
				cpy.Tag = r.(*ast.BasicLit)
			})
		}
		if cpy.Comment != nil {
			walkToReplace(v, cpy.Comment, func(r ast.Node) {
				cpy.Comment = r.(*ast.CommentGroup)
			})
		}
		return &cpy

	case *ast.FieldList:
		cpy := *n
		cpy.List = make([]*ast.Field, len(n.List))
		copy(cpy.List, n.List)

		for i, f := range cpy.List {
			walkToReplace(v, f, func(r ast.Node) {
				cpy.List[i] = r.(*ast.Field)
			})
		}

		// Expressions
		cpy := *n
		return &cpy

	case *ast.Ellipsis:
		cpy := *n

		if cpy.Elt != nil {
			walkToReplace(v, cpy.Elt, func(r ast.Node) {
				cpy.Elt = r.(ast.Expr)
			})
		}

		return &cpy

	case *ast.FuncLit:
		cpy := *n
		walkToReplace(v, n.Type, func(r ast.Node) {
			n.Type = r.(*ast.FuncType)
		})
		walkToReplace(v, n.Body, func(r ast.Node) {
			n.Body = r.(*ast.BlockStmt)
		})

	case *ast.CompositeLit:
		cpy := *n
		if n.Type != nil {
			walkToReplace(v, n.Type, func(r ast.Node) {
				n.Type = r.(ast.Expr)
			})
		}
		walkExprList(v, n.Elts)

	case *ast.ParenExpr:
		cpy := *n
		walkToReplace(v, n.X, func(r ast.Node) {
			n.X = r.(ast.Expr)
		})

	case *ast.SelectorExpr:
		cpy := *n
		walkToReplace(v, n.X, func(r ast.Node) {
			n.X = r.(ast.Expr)
		})
		walkToReplace(v, n.Sel, func(r ast.Node) {
			n.Sel = r.(*ast.Ident)
		})

	case *ast.IndexExpr:
		cpy := *n
		walkToReplace(v, n.X, func(r ast.Node) {
			n.X = r.(ast.Expr)
		})
		walkToReplace(v, n.Index, func(r ast.Node) {
			n.Index = r.(ast.Expr)
		})

	case *ast.SliceExpr:
		cpy := *n
		walkToReplace(v, n.X, func(r ast.Node) {
			n.X = r.(ast.Expr)
		})
		if n.Low != nil {
			walkToReplace(v, n.Low, func(r ast.Node) {
				n.Low = r.(ast.Expr)
			})
		}
		if n.High != nil {
			walkToReplace(v, n.High, func(r ast.Node) {
				n.High = r.(ast.Expr)
			})
		}
		if n.Max != nil {
			walkToReplace(v, n.Max, func(r ast.Node) {
				n.Max = r.(ast.Expr)
			})
		}

	case *ast.TypeAssertExpr:
		cpy := *n
		walkToReplace(v, n.X, func(r ast.Node) {
			n.X = r.(ast.Expr)
		})
		if n.Type != nil {
			walkToReplace(v, n.Type, func(r ast.Node) {
				n.Type = r.(ast.Expr)
			})
		}

	case *ast.CallExpr:
		cpy := *n
		walkToReplace(v, n.Fun, func(r ast.Node) {
			n.Fun = r.(ast.Expr)
		})
		walkExprList(v, n.Args)

	case *ast.StarExpr:
		cpy := *n
		walkToReplace(v, n.Fun, func(r ast.Node) {
		walkToReplace(v, n.X, func(r ast.Node) {
			n.X = r.(ast.Expr)
		})

	case *ast.UnaryExpr:
		cpy := *n
		walkToReplace(v, n.Fun, func(r ast.Node) {
		walkToReplace(v, n.Fun, func(r ast.Node) {
		walkToReplace(v, n.X, func(r ast.Node) {
			n.X = r.(ast.Expr)
		})

	case *ast.BinaryExpr:
		cpy := *n
		walkToReplace(v, n.Fun, func(r ast.Node) {
		walkToReplace(v, n.X, func(r ast.Node) {
			n.X = r.(ast.Expr)
		})
		walkToReplace(v, n.Y, func(r ast.Node) {
			n.Y = r.(ast.Expr)
		})

	case *ast.KeyValueExpr:
		cpy := *n
		walkToReplace(v, n.Key, func(r ast.Node) {
			n.Key = r.(ast.Expr)
		})
		walkToReplace(v, n.Value, func(r ast.Node) {
			n.Value = r.(ast.Expr)
		})

	// Types
	case *ast.ArrayType:
		cpy := *n
		if n.Len != nil {
			walkToReplace(v, n.Len, func(r ast.Node) {
				n.Len = r.(ast.Expr)
			})
		}
		walkToReplace(v, n.Elt, func(r ast.Node) {
			n.Elt = r.(ast.Expr)
		})

	case *ast.StructType:
		cpy := *n
		walkToReplace(v, n.Fields, func(r ast.Node) {
			n.Fields = r.(*ast.FieldList)
		})

	case *ast.FuncType:
		cpy := *n
		if n.Params != nil {
			walkToReplace(v, n.Params, func(r ast.Node) {
				n.Params = r.(*ast.FieldList)
			})
		}
		if n.Results != nil {
			walkToReplace(v, n.Results, func(r ast.Node) {
				n.Results = r.(*ast.FieldList)
			})
		}

	case *ast.InterfaceType:
		cpy := *n
		walkToReplace(v, n.Methods, func(r ast.Node) {
			n.Methods = r.(*ast.FieldList)
		})

	case *ast.MapType:
		cpy := *n
		walkToReplace(v, n.Key, func(r ast.Node) {
			n.Key = r.(ast.Expr)
		})
		walkToReplace(v, n.Value, func(r ast.Node) {
			n.Value = r.(ast.Expr)
		})

	case *ast.ChanType:
		cpy := *n
		walkToReplace(v, n.Value, func(r ast.Node) {
			n.Value = r.(ast.Expr)
		})

	// Statements
	case *ast.BadStmt:
		// nothing to do

	case *ast.DeclStmt:
		cpy := *n
		walkToReplace(v, n.Decl, func(r ast.Node) {
			n.Decl = r.(ast.Decl)
		})

	case *ast.EmptyStmt:
		// nothing to do

	case *ast.LabeledStmt:
		cpy := *n
		walkToReplace(v, n.Label, func(r ast.Node) {
			n.Label = r.(*ast.Ident)
		})
		walkToReplace(v, n.Stmt, func(r ast.Node) {
			n.Stmt = r.(ast.Stmt)
		})

	case *ast.ExprStmt:
		cpy := *n
		walkToReplace(v, n.X, func(r ast.Node) {
			n.X = r.(ast.Expr)
		})

	case *ast.SendStmt:
		cpy := *n
		walkToReplace(v, n.Chan, func(r ast.Node) {
			n.Chan = r.(ast.Expr)
		})
		walkToReplace(v, n.Value, func(r ast.Node) {
			n.Value = r.(ast.Expr)
		})

	case *ast.IncDecStmt:
		cpy := *n
		walkToReplace(v, n.X, func(r ast.Node) {
			n.X = r.(ast.Expr)
		})

	case *ast.AssignStmt:
		cpy := *n
		walkExprList(v, n.Lhs)
		walkExprList(v, n.Rhs)

	case *ast.GoStmt:
		cpy := *n
		walkToReplace(v, n.Call, func(r ast.Node) {
			n.Call = r.(*ast.CallExpr)
		})

	case *ast.DeferStmt:
		cpy := *n
		walkToReplace(v, n.Call, func(r ast.Node) {
			n.Call = r.(*ast.CallExpr)
		})

	case *ast.ReturnStmt:
		walkExprList(v, n.Results)

	case *ast.BranchStmt:
		if n.Label != nil {
			walkToReplace(v, n.Label, func(r ast.Node) {
				n.Label = r.(*ast.Ident)
			})
		}

	case *ast.BlockStmt:
		walkStmtList(v, n.List)

	case *ast.IfStmt:
		if n.Init != nil {
			walkToReplace(v, n.Init, func(r ast.Node) {
				n.Init = r.(ast.Stmt)
			})
		}
		walkToReplace(v, n.Cond, func(r ast.Node) {
			n.Cond = r.(ast.Expr)
		})
		walkToReplace(v, n.Body, func(r ast.Node) {
			n.Body = r.(*ast.BlockStmt)
		})
		if n.Else != nil {
			walkToReplace(v, n.Else, func(r ast.Node) {
				n.Else = r.(ast.Stmt)
			})
		}

	case *ast.CaseClause:
		walkExprList(v, n.List)
		walkStmtList(v, n.Body)

	case *ast.SwitchStmt:
		if n.Init != nil {
			walkToReplace(v, n.Init, func(r ast.Node) {
				n.Init = r.(ast.Stmt)
			})
		}
		if n.Tag != nil {
			walkToReplace(v, n.Tag, func(r ast.Node) {
				n.Tag = r.(ast.Expr)
			})
		}
		walkToReplace(v, n.Body, func(r ast.Node) {
			n.Body = r.(*ast.BlockStmt)
		})

	case *ast.TypeSwitchStmt:
		if n.Init != nil {
			walkToReplace(v, n.Init, func(r ast.Node) {
				n.Init = r.(ast.Stmt)
			})
		}
		walkToReplace(v, n.Assign, func(r ast.Node) {
			n.Assign = r.(ast.Stmt)
		})
		walkToReplace(v, n.Body, func(r ast.Node) {
			n.Body = r.(*ast.BlockStmt)
		})

	case *ast.CommClause:
		if n.Comm != nil {
			walkToReplace(v, n.Comm, func(r ast.Node) {
				n.Comm = r.(ast.Stmt)
			})
		}
		walkStmtList(v, n.Body)

	case *ast.SelectStmt:
		walkToReplace(v, n.Body, func(r ast.Node) {
			n.Body = r.(*ast.BlockStmt)
		})

	case *ast.ForStmt:
		if n.Init != nil {
			walkToReplace(v, n.Init, func(r ast.Node) {
				n.Init = r.(ast.Stmt)
			})
		}
		if n.Cond != nil {
			walkToReplace(v, n.Cond, func(r ast.Node) {
				n.Cond = r.(ast.Expr)
			})
		}
		if n.Post != nil {
			walkToReplace(v, n.Post, func(r ast.Node) {
				n.Post = r.(ast.Stmt)
			})
		}
		walkToReplace(v, n.Body, func(r ast.Node) {
			n.Body = r.(*ast.BlockStmt)
		})

	case *ast.RangeStmt:
		if n.Key != nil {
			walkToReplace(v, n.Key, func(r ast.Node) {
				n.Key = r.(ast.Expr)
			})
		}
		if n.Value != nil {
			walkToReplace(v, n.Value, func(r ast.Node) {
				n.Value = r.(ast.Expr)
			})
		}
		walkToReplace(v, n.X, func(r ast.Node) {
			n.X = r.(ast.Expr)
		})
		walkToReplace(v, n.Body, func(r ast.Node) {
			n.Body = r.(*ast.BlockStmt)
		})

	// Declarations
	case *ast.ImportSpec:
		if n.Doc != nil {
			walkToReplace(v, n.Doc, func(r ast.Node) {
				n.Doc = r.(*ast.CommentGroup)
			})
		}
		if n.Name != nil {
			walkToReplace(v, n.Name, func(r ast.Node) {
				n.Name = r.(*ast.Ident)
			})
		}
		walkToReplace(v, n.Path, func(r ast.Node) {
			n.Path = r.(*ast.BasicLit)
		})
		if n.Comment != nil {
			walkToReplace(v, n.Comment, func(r ast.Node) {
				n.Comment = r.(*ast.CommentGroup)
			})
		}

	case *ast.ValueSpec:
		if n.Doc != nil {
			walkToReplace(v, n.Doc, func(r ast.Node) {
				n.Doc = r.(*ast.CommentGroup)
			})
		}
		walkIdentList(v, n.Names)
		if n.Type != nil {
			walkToReplace(v, n.Type, func(r ast.Node) {
				n.Type = r.(ast.Expr)
			})
		}
		walkExprList(v, n.Values)
		if n.Comment != nil {
			walkToReplace(v, n.Comment, func(r ast.Node) {
				n.Comment = r.(*ast.CommentGroup)
			})
		}

	case *ast.TypeSpec:
		if n.Doc != nil {
			walkToReplace(v, n.Doc, func(r ast.Node) {
				n.Doc = r.(*ast.CommentGroup)
			})
		}
		walkToReplace(v, n.Name, func(r ast.Node) {
			n.Name = r.(*ast.Ident)
		})
		walkToReplace(v, n.Type, func(r ast.Node) {
			n.Type = r.(ast.Expr)
		})
		if n.Comment != nil {
			walkToReplace(v, n.Comment, func(r ast.Node) {
				n.Comment = r.(*ast.CommentGroup)
			})
		}

	case *ast.BadDecl:
		// nothing to do

	case *ast.GenDecl:
		if n.Doc != nil {
			walkToReplace(v, n.Doc, func(r ast.Node) {
				n.Doc = r.(*ast.CommentGroup)
			})
		}
		for i, s := range n.Specs {
			walkToReplace(v, s, func(r ast.Node) {
				n.Specs[i] = r.(ast.Spec)
			})
		}

	case *ast.FuncDecl:
		if n.Doc != nil {
			walkToReplace(v, n.Doc, func(r ast.Node) {
				n.Doc = r.(*ast.CommentGroup)
			})
		}
		if n.Recv != nil {
			walkToReplace(v, n.Recv, func(r ast.Node) {
				n.Recv = r.(*ast.FieldList)
			})
		}
		walkToReplace(v, n.Name, func(r ast.Node) {
			n.Name = r.(*ast.Ident)
		})
		walkToReplace(v, n.Type, func(r ast.Node) {
			n.Type = r.(*ast.FuncType)
		})
		if n.Body != nil {
			walkToReplace(v, n.Body, func(r ast.Node) {
				n.Body = r.(*ast.BlockStmt)
			})
		}

	// Files and packages
	case *ast.File:
		if n.Doc != nil {
			walkToReplace(v, n.Doc, func(r ast.Node) {
				n.Doc = r.(*ast.CommentGroup)
			})
		}
		walkToReplace(v, n.Name, func(r ast.Node) {
			n.Name = r.(*ast.Ident)
		})
		walkDeclList(v, n.Decls)
		// don't walk n.Comments - they have been
		// visited already through the individual
		// nodes

	case *ast.Package:
		for i, f := range n.Files {
			walkToReplace(v, f, func(r ast.Node) {
				n.Files[i] = r.(*ast.File)
			})
		}

	default:
		panic(fmt.Sprintf("walkToReplace: unexpected node type %T", n))
	}

	v.Visit(nil, func(ast.Node) { panic("can't replace the go-up nil") })
}
