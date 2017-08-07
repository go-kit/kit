package main

import (
	"go/ast"
)

func transformAST(ctx *sourceContext) (ast.Node, error) {
	root := &ast.File{
		Name:    ctx.pkg,
		Imports: ctx.imports,
		Decls:   []ast.Decl{},
	}

	for _, iface := range ctx.interfaces { //only one...
		addStubStruct(root, iface)

		for _, meth := range iface.methods {
			addMethod(root, iface, meth)
			addRequestStruct(root, meth)
			addResponseStruct(root, meth)
			addEndpointMaker(root, iface, meth)
		}

		addEndpointsStruct(root, iface)
		addHTTPHandler(root, iface)

		for _, meth := range iface.methods {
			addDecoder(root, meth)
			addEncoder(root, meth)
		}
	}
	return root, nil
}

func addStubStruct(root *ast.File, iface iface) {
	root.Decls = append(root.Decls, iface.stubStructDecl())
}

func addMethod(root *ast.File, iface iface, meth method) {
	def := meth.definition(iface)
	root.Decls = append(root.Decls, def)
}

func addRequestStruct(root *ast.File, meth method) {
	root.Decls = append(root.Decls, meth.requestStruct())
}

func addResponseStruct(root *ast.File, meth method) {
	root.Decls = append(root.Decls, meth.responseStruct())
}

func addEndpointMaker(root *ast.File, ifc iface, meth method) {
	root.Decls = append(root.Decls, meth.endpointMaker(ifc))
}

func addEndpointsStruct(root *ast.File, ifc iface) {
	root.Decls = append(root.Decls, ifc.endpointsStruct())
}

func addHTTPHandler(root *ast.File, ifc iface) {
	root.Decls = append(root.Decls, ifc.httpHandler())
}

func addDecoder(root *ast.File, meth method) {
	root.Decls = append(root.Decls, meth.decoderFunc())
}

func addEncoder(root *ast.File, meth method) {
	root.Decls = append(root.Decls, meth.encoderFunc())
}
