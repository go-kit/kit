package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	"golang.org/x/tools/imports"
)

type (
	files  map[string]io.Reader
	layout interface {
		transformAST(ctx *sourceContext) (files, error)
	}

	flat      struct{}
	deflayout struct {
		targetDir string
	}

	outputTree map[string]*ast.File
)

func (ot outputTree) addFile(path, pkgname string) *ast.File {
	file := &ast.File{
		Name:  id(pkgname),
		Decls: []ast.Decl{},
	}
	ot[path] = file
	return file
}

func getGopath() string {
	gopath, set := os.LookupEnv("GOPATH")
	if !set {
		return filepath.Join(os.Getenv("HOME"), "go")
	}
	return gopath
}

func importPath(targetDir, gopath string) (string, error) {
	if !filepath.IsAbs(targetDir) {
		return "", fmt.Errorf("%q is not an absolute path", targetDir)
	}

	for _, dir := range filepath.SplitList(gopath) {
		abspath, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		srcPath := filepath.Join(abspath, "src")

		res, err := filepath.Rel(srcPath, targetDir)
		if err != nil {
			continue
		}
		if strings.Index(res, "..") == -1 {
			return res, nil
		}
	}
	return "", fmt.Errorf("%q is not in GOPATH (%s)", targetDir, gopath)

}

func (l deflayout) packagePath(sub string) string {
	return filepath.Join(l.targetDir, sub)
}

func (l deflayout) transformAST(ctx *sourceContext) (files, error) {
	out := make(outputTree)

	endpoints := out.addFile("endpoints/endpoints.go", "endpoints")
	http := out.addFile("http/http.go", "http")
	service := out.addFile("service/service.go", "service")

	addImports(endpoints, ctx)
	addImports(http, ctx)
	addImports(service, ctx)

	for _, iface := range ctx.interfaces { //only one...
		addStubStruct(service, iface)

		for _, meth := range iface.methods {
			addMethod(service, iface, meth)
			addRequestStruct(endpoints, meth)
			addResponseStruct(endpoints, meth)
			addEndpointMaker(endpoints, iface, meth)
		}

		addEndpointsStruct(endpoints, iface)
		addHTTPHandler(http, iface)

		for _, meth := range iface.methods {
			addDecoder(http, meth)
			addEncoder(http, meth)
		}

		for _, file := range out {
			selectify(file, "service", iface.stubName().Name, l.packagePath("service"))
			selectify(file, "endpoints", "Endpoints", l.packagePath("endpoints"))
			for _, meth := range iface.methods {
				selectify(file, "endpoints", meth.requestStructName().Name, l.packagePath("endpoints"))
			}
		}
	}

	return formatNodes(out)
}

func selectify(file *ast.File, pkgName, identName, importPath string) {
	if file.Name.Name == pkgName {
		return
	}

	selector := sel(id(pkgName), id(identName))
	if selectifyIdent(identName, file, selector) {
		addImport(file, importPath)
	}
}

type selIdentFn func(ast.Node, func(ast.Node)) Visitor

func (f selIdentFn) Visit(node ast.Node, r func(ast.Node)) Visitor {
	return f(node, r)
}

func selectifyIdent(identName string, file *ast.File, selector ast.Expr) (replaced bool) {
	var r selIdentFn
	r = selIdentFn(func(node ast.Node, replaceWith func(ast.Node)) Visitor {
		switch id := node.(type) {
		case *ast.SelectorExpr:
			return nil
		case *ast.Ident:
			if id.Name == identName {
				replaced = true
				replaceWith(selector)
			}
		}
		return r
	})
	WalkReplace(r, file)
	return
}

func (f flat) transformAST(ctx *sourceContext) (files, error) {
	root := &ast.File{
		Name:  ctx.pkg,
		Decls: []ast.Decl{},
	}

	addImports(root, ctx)

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

	return formatNodes(outputTree{"gokit.go": root})
}

func formatNode(fname string, node ast.Node) (*bytes.Buffer, error) {
	outfset := token.NewFileSet()
	buf := &bytes.Buffer{}
	err := format.Node(buf, outfset, node)
	if err != nil {
		return nil, err
	}
	imps, err := imports.Process(fname, buf.Bytes(), nil)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(imps), nil
}

type sortableDecls []ast.Decl

func (sd sortableDecls) Len() int {
	return len(sd)
}

func (sd sortableDecls) Less(i int, j int) bool {
	switch left := sd[i].(type) {
	case *ast.GenDecl:
		switch right := sd[j].(type) {
		default:
			return left.Tok == token.IMPORT
		case *ast.GenDecl:
			return left.Tok == token.IMPORT && right.Tok != token.IMPORT
		}
	}
	return false
}

func (sd sortableDecls) Swap(i int, j int) {
	sd[i], sd[j] = sd[j], sd[i]
}

func formatNodes(nodes outputTree) (files, error) {
	res := files{}
	var err error
	for fn, node := range nodes {
		sort.Stable(sortableDecls(node.Decls))
		res[fn], err = formatNode(fn, node)
		if err != nil {
			return nil, errors.Wrapf(err, "formatNodes")
		}
	}
	return res, nil
}

// XXX debug
func spewDecls(f *ast.File) {
	for _, d := range f.Decls {
		switch dcl := d.(type) {
		default:
			spew.Dump(dcl)
		case *ast.GenDecl:
			spew.Dump(dcl.Tok)
		case *ast.FuncDecl:
			spew.Dump(dcl.Name.Name)
		}
	}
}

func addImports(root *ast.File, ctx *sourceContext) {
	root.Decls = append(root.Decls, ctx.importDecls()...)
}

func addImport(root *ast.File, path string) {
	root.Decls = append(root.Decls, importFor(importSpec(path)))
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
