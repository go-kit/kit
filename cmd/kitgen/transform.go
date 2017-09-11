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
	"strings"

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
	}

	for _, file := range out {
		selectify(file, "Endpoints", "endpoints", l.packagePath("endpoints"))
	}

	return formatNodes(out)
}

func selectify(file *ast.File, identName, pkgName, importPath string) {
	if file.Name.Name == pkgName {
		return
	}

	selector := sel(id(pkgName), id(identName))
	if replaceIdent(identName, file, selector) {
		addImport(file, importPath)
	}
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

func formatNodes(nodes outputTree) (files, error) {
	res := files{}
	var err error
	for fn, node := range nodes {
		res[fn], err = formatNode(fn, node)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
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
	root.Decls = append(root.Decls, ifc.httpHandler(sel(id("endpoints"), id("Endpoints"))))
}

func addDecoder(root *ast.File, meth method) {
	root.Decls = append(root.Decls, meth.decoderFunc())
}

func addEncoder(root *ast.File, meth method) {
	root.Decls = append(root.Decls, meth.encoderFunc())
}
