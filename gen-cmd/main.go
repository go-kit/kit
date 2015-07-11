// generator for the boilerplate code for the endpoint
// to generate the code anotate the endpoint definition with
//		//go:generate go run ../gen-cmd/main.go packageName
//		//go:generate go fmt
// and run `go generate`
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"text/template"
	"unicode"
)

const (
	contextType = "context.Context"
)

func main() {
	pck := os.Args[1]
	f := buildAst(pck)

	vs := &Visitor{
		Package: pck,

		// function helpers
		ReturnExp: ReturnExp,
		ArgsExp:   ArgsExp,
		CompoExp:  CompoExp,
	}

	ast.Walk(vs, f)

	// TODO remove relative path
	mergeTemplate(vs, "definition", "../gen-cmd/endpoint_definition.tmpl")
}

func buildAst(pck string) *ast.File {
	definitions := pck + ".go"
	fset := token.NewFileSet()

	fmt.Println("Reading type definitions from", definitions)
	f, err := parser.ParseFile(fset, definitions, nil, 0)
	if err != nil {
		panic("can not parse file " + err.Error())
	}
	// Useful to debug the ast
	// ast.Print(fset, f)
	return f
}

func mergeTemplate(visitor *Visitor, templateName, templatePath string) {
	fmt.Println("Reading template", templateName, "in the path", templatePath)

	t, e := template.ParseFiles(templatePath)
	if e != nil {
		fmt.Println(e)
		return
	}

	filename := fmt.Sprintf("%s_%s.go", visitor.Package, templateName)
	fmt.Println("Creating file", filename)
	f, err := os.Create(filename)
	if err != nil {
		panic("can not create file " + err.Error())
		return
	}
	defer f.Close()

	fmt.Println("Merge the template", templateName)
	err = t.Execute(f, visitor)
	if err != nil {
		panic("can not create execute the template " + err.Error())
		return
	}
}

type Visitor struct {
	// Data extracted from ast
	Package   string
	Functions []Func

	// functions to help builing the template
	ReturnExp func([]Def) string
	ArgsExp   func(string, []Def) string
	CompoExp  func([]Def) string
}

func (v *Visitor) Visit(node ast.Node) (w ast.Visitor) {
	t, ok := node.(*ast.TypeSpec)
	if ok == false {
		return v
	}
	ft, ok := t.Type.(*ast.FuncType)
	if ok == false {
		return v
	}

	fu := Func{Name: t.Name.String()}

	fu.Args = definitions(ft.Params.List)
	fu.Ret = definitions(ft.Results.List)
	v.Functions = append(v.Functions, fu)

	return v
}

type Func struct {
	Name string
	Args []Def
	Ret  []Def
}

type Def struct {
	Name string
	Type string
}

func (d Def) ToStructDef() string {
	if d.Type == "error" {
		return ""
	}
	if d.Type == contextType {
		return ""
	}
	return fmt.Sprintf("%s %s", ToUpper(d.Name), d.Type)
}

// helper to output a string of return values like: "ret1, ret2, err"
func ReturnExp(defs []Def) string {
	names := []string{}
	for _, d := range defs {
		if d.Type == "error" {
			d.Name = "err"
		}
		names = append(names, d.Name)
	}
	return strings.Join(names, ",")
}

// helper to output a string of argument exp like: "prefix.Arg1, prefix.Arg2, prefix.Arg3"
func ArgsExp(prefix string, defs []Def) string {
	names := []string{}
	for _, d := range defs {
		if d.Type == contextType {
			continue
		}
		names = append(names, fmt.Sprintf("%s.%s", prefix, ToUpper(d.Name)))
	}
	return strings.Join(names, ",")
}

// helper to create composite literals from the definitions like: "Ret: ret, Ret2: ret2"
func CompoExp(defs []Def) string {
	composite := []string{}
	for _, d := range defs {
		if d.Type == "error" {
			continue
		}
		exp := fmt.Sprintf("%s:%s", ToUpper(d.Name), d.Name)
		composite = append(composite, exp)
	}
	return strings.Join(composite, ",")
}

// first letter to upper case
func ToUpper(s string) string {
	a := []rune(s)
	a[0] = unicode.ToUpper(a[0])
	return string(a)
}

// ast to extract definitions
func definitions(fields []*ast.Field) []Def {
	defs := []Def{}
	for index, p := range fields {
		// TODO handle error
		ty, _ := getType(p.Type)
		if len(p.Names) == 0 {
			defs = append(defs, Def{
				Name: fmt.Sprintf("arg%v", index),
				Type: ty,
			})
		}
		for _, n := range p.Names {
			defs = append(defs, Def{Name: n.String(), Type: ty})
		}
	}
	return defs
}

// ast to extract types
func getType(exp ast.Expr) (string, error) {
	id, ok := exp.(*ast.Ident)
	if ok {
		return id.String(), nil
	}

	sel, ok := exp.(*ast.SelectorExpr)
	if ok {
		return fmt.Sprintf("%v.%v", sel.X, sel.Sel), nil
	}
	return "", fmt.Errorf("Can not extract the type from %v type %T", exp, exp)
}
