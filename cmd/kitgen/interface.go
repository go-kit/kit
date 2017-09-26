package main

import "go/ast"

// because "interface" is a keyword...
type iface struct {
	name, stubname, rcvrName *ast.Ident
	methods                  []method
}

func (i iface) stubName() *ast.Ident {
	return i.stubname
}

func (i iface) stubStructDecl() ast.Decl {
	return structDecl(i.stubName(), &ast.FieldList{})
}

func (i iface) endpointsStruct() ast.Decl {
	fl := &ast.FieldList{}
	for _, m := range i.methods {
		fl.List = append(fl.List, &ast.Field{Names: []*ast.Ident{m.name}, Type: sel(id("endpoint"), id("Endpoint"))})
	}
	return structDecl(id("Endpoints"), fl)
}

func (i iface) httpHandler() ast.Decl {
	handlerFn := fetchFuncDecl("NewHTTPHandler")

	// xxx does this "inlining" process merit a helper akin to replaceIdent?
	handleCalls := []ast.Stmt{}
	for _, m := range i.methods {
		handleCall := fetchFuncDecl("inlineHandlerBuilder").Body.List[0].(*ast.ExprStmt).X.(*ast.CallExpr)

		handleCall.Args[0].(*ast.BasicLit).Value = `"` + m.pathName() + `"`

		handleCall.Args[1].(*ast.CallExpr).Args =
			[]ast.Expr{sel(id("endpoints"), m.name), m.decodeFuncName(), m.encodeFuncName()}

		handleCalls = append(handleCalls, &ast.ExprStmt{X: handleCall})
	}

	pasteStmts(handlerFn.Body, 1, handleCalls)

	return handlerFn
}

func (i iface) reciever() *ast.Field {
	return &ast.Field{
		Names: []*ast.Ident{i.receiverName()},
		Type:  i.stubName(),
	}
}

func (i iface) receiverName() *ast.Ident {
	if i.rcvrName != nil {
		return i.rcvrName
	}
	scope := ast.NewScope(nil)
	for _, meth := range i.methods {
		for _, arg := range meth.params {
			if arg.name != nil {
				scope.Insert(ast.NewObj(ast.Var, arg.name.Name))
			}
		}
		for _, arg := range meth.results {
			if arg.name != nil {
				scope.Insert(ast.NewObj(ast.Var, arg.name.Name))
			}
		}
	}
	return id(unexport(inventName(i.name, scope).Name))
}
