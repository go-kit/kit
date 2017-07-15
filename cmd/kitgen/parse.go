package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"

	"go/ast"
	"go/parser"
	"go/token"
)

// Interface is the definition of our service.
type Interface struct {
	Name    string // Service (must be exported)
	Methods []Method
}

func (i Interface) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "type %s interface {\n", i.Name)
	for _, method := range i.Methods {
		fmt.Fprintf(&buf, "\t%s\n", method)
	}
	fmt.Fprintf(&buf, "}\n")
	return buf.String()
}

// Method describes one method of the interface.
type Method struct {
	Name       string // Foo
	Parameters []NamedType
	Results    []NamedType
}

func (m Method) String() string {
	var params []string
	for _, param := range m.Parameters {
		params = append(params, param.String())
	}
	paramStr := strings.Join(params, ", ")

	var results []string
	for _, result := range m.Results {
		results = append(results, result.String())
	}
	var resultStr string
	switch len(m.Results) {
	case 0:
		resultStr = ""
	case 1:
		resultStr = m.Results[0].Type
	default:
		resultStr = "(" + strings.Join(results, ", ") + ")"
	}

	return strings.TrimSpace(fmt.Sprintf("%s(%s) %s", m.Name, paramStr, resultStr))
}

// NamedType is a name and type tuple, e.g. `s string`.
type NamedType struct {
	Name string // "ctx"
	Type string // "context.Context"
}

func (nt NamedType) String() string {
	return fmt.Sprintf("%s %s", nt.Name, nt.Type)
}

func parseFiles(name string, filenames ...string) ([]string, Interface, error) {
	for _, filename := range filenames {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
		if err != nil {
			log.Printf("%s: %v", filename, err)
			continue
		}
		imports, iface, ok := parseFile(name, f)
		if ok {
			Debugf("Successfully parsed %s from %s", name, filename)
			return imports, iface, nil
		}
	}
	return []string{}, Interface{}, errors.New("no suitable interface found")
}

func parseFile(name string, f *ast.File) ([]string, Interface, bool) {
	Debugf("Parsing file %s", f.Name)
	var imports []string
	for i, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			Debugf("%s: skipping Decl %d", f.Name, i)
			continue
		}

		for j, spec := range genDecl.Specs {
			switch s := spec.(type) {
			case *ast.ImportSpec:
				importString := parseImportSpec(s)
				Debugf("Appending import %s", importString)
				imports = append(imports, importString)

			case *ast.TypeSpec:
				if iface, ok := parseTypeSpec(name, s); ok {
					return imports, iface, true
				}
				Debugf("Failed to parse an Interface; continuing")

			default:
				Debugf("%s: Decl %d: skipping Spec %d", f.Name, i, j)
			}
		}
	}
	return []string{}, Interface{}, false
}

func parseImportSpec(s *ast.ImportSpec) string {
	Debugf("Parsing ImportSpec %s", s.Name)

	importAlias, importPath := "", s.Path.Value
	if s.Name != nil {
		importAlias = s.Name.Name
	}
	importString := strings.TrimSpace(fmt.Sprintf("%s %s", importAlias, importPath))
	return importString
}

func parseTypeSpec(name string, s *ast.TypeSpec) (Interface, bool) {
	Debugf("Parsing TypeSpec %s", s.Name)

	t, ok := s.Type.(*ast.InterfaceType)
	if !ok {
		Debugf("Skipping a non-interface type")
		return Interface{}, false
	}
	if s.Name.String() != name {
		Debugf("Found interface %q, but looking for %q; skipping", s.Name, name)
		return Interface{}, false
	}
	return parseInterfaceType(s.Name.Name, t), true
}

func parseInterfaceType(interfaceName string, t *ast.InterfaceType) Interface {
	Debugf("Parsing InterfaceType %s", interfaceName)

	var methods []Method
	for i, field := range t.Methods.List {
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			Debugf("Skipping non-function field %d", i)
			continue
		}

		var methodName string
		for _, ident := range field.Names {
			if !ident.IsExported() {
				continue
			}
			methodName = ident.Name
			break
		}
		if methodName == "" {
			Debugf("Field %d had no exported name; skipping", i)
			continue
		}

		method := parseMethodField(methodName, field, funcType)
		methods = append(methods, method)
	}

	return Interface{
		Name:    interfaceName,
		Methods: methods,
	}
}

func parseMethodField(methodName string, field *ast.Field, funcType *ast.FuncType) Method {
	Debugf("Parsing method %s", methodName)

	params := parseFieldListAsNamedTypes(funcType.Params)
	Debugf("Params: %#+v", params)

	results := parseFieldListAsNamedTypes(funcType.Results)
	Debugf("Results: %#+v", results)

	return Method{
		Name:       methodName,
		Parameters: params,
		Results:    results,
	}
}

func parseFieldListAsNamedTypes(list *ast.FieldList) []NamedType {
	Debugf("Parsing FieldList (%d) as NamedTypes", len(list.List))

	var namedTypes []NamedType
	for _, field := range list.List {
		// Always 1 type
		var typ string
		switch t := field.Type.(type) {
		case *ast.Ident:
			Debugf("Type Ident, i.e. a built-in type")
			typ = t.Name

		case *ast.SelectorExpr:
			Debugf("Type Selector, i.e. a third-party type")
			selectorIdent, ok := t.X.(*ast.Ident)
			if !ok {
				Debugf("Selector X isn't an Ident; very odd, skipping")
				continue
			}
			typ = fmt.Sprintf("%s.%s", selectorIdent.Name, t.Sel.Name)

		default:
			Debugf("Skipping unknown Field Type")
			continue
		}
		Debugf("Type %s", typ)

		// Potentially N names
		var names []string
		for _, ident := range field.Names {
			names = append(names, ident.Name)
		}
		if len(names) == 0 {
			// Anonymous named type, give it a default name
			names = append(names, "somename") // TODO(pb): generator
		}
		for _, name := range names {
			namedType := NamedType{
				Name: name,
				Type: typ,
			}
			Debugf("NamedType %+v", namedType)
			namedTypes = append(namedTypes, namedType)
		}
	}
	return namedTypes
}
