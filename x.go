package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/tools/go/types"

	"github.com/davecgh/go-spew/spew"
	_ "golang.org/x/tools/go/gcimporter"
)

func init() {
	flag.StringVar(&gen.pkg, "package", "", "")
	flag.BoolVar(&gen.w, "w", false, "will (over)write files if set. Prints to stdout otherwise.")
	flag.StringVar(&gen.typ, "type", "", "type names; must be set")
}

func main() {
	flag.Parse()
	if gen.pkg == "" {
		oops("-package should not be empty")
	}
	panic(gen.gen())
}

func (gen *g) gen() error {
	// Build the functions map.
	fns := map[string]*types.Signature{}
	for _, f := range flag.Args() {
		fns[f] = nil
	}

	pk, err := build.Default.Import(gen.pkg, "", 0)
	if err != nil {
		return err
	}
	var fset = token.NewFileSet()
	var fs []*ast.File
	// Build AST for all the go files.
	for _, f := range pk.GoFiles {
		full := filepath.Join(pk.Dir, f)
		fmt.Fprintf(os.Stderr, "processing `%s`\n", full)
		r, err := os.Open(full)
		if err != nil {
			return fmt.Errorf("Can not be open `%s` for reading. Error: `%s`\n", full, err)
		}
		defer r.Close()
		f, err := parser.ParseFile(fset, full, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		fs = append(fs, f)
	}
	// Typecheck and collect multiErr.
	var errs multiErr
	conf := types.Config{
		IgnoreFuncBodies: true,
		Error: func(err error) {
			errs.Add(err)
		},
	}
	err = errs.ToErr()
	if err != nil {
		return err
	}
	if len(fs) == 0 {
		return errors.New("no go files")
	}
	info := &types.Info{
		Types:      map[ast.Expr]types.TypeAndValue{},
		Defs:       map[*ast.Ident]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
		Uses:       map[*ast.Ident]types.Object{},
	}
	pkg, err := conf.Check(fs[0].Name.Name, fset, fs, info)
	if err != nil {
		return err
	}
	iface, err := find(pkg, gen.typ)
	if err != nil {
		return err
	}

	// Build the reverse map.
	rev := map[types.Type]ast.Expr{}
	for k, v := range info.Types {
		rev[v.Type] = k
	}
	spew.Dump(iface)

	for i := 0; i < iface.NumMethods(); i++ {
		spew.Dump(iface.Method(i))
	}
	panic(0)

	updateFuncMap(pkg, fns)
	for name, f := range fns {
		if f == nil {
			log.Fatalln(name, "not found")
		}
		// for i := 0; i < f.Params().Len(); i++ {
		// 	pos := pkg.Scope().Innermost(f.Params().At(i).Pos()).Pos()
		// 	spew.Dump(fset.PositionFor(pos, false))
		// 	spew.Dump("~~~~", f.Params().At(i), pos)
		// }

		ps := toList(f.Params())
		var q []*types.Var
		var tags []string
		un := uniqueNames{}
		for _, p := range ps {
			switch p.Type().String() {
			case "golang.org/x/net/context.Context":
				continue
			case "error":
				continue
			}
			n := p.Name()
			if n == "" {
				n = p.Type().String()
				if !isValidIdentifier(n) {
					n = "p"
				}
			}
			if pkger, ok := p.Type().Underlying().(packager); ok {
				fmt.Println(p.Type(), pkger.Pkg(), "::::---")
			}
			spew.Dump(rev[p.Type()])
			imps := map[types.Object]*ast.SelectorExpr{}
			updateDeps(rev[p.Type()], *info, imps)
			fmt.Println("-=----=-\n")
			for k, _ := range imps {
				fmt.Println(k.Pkg())
			}
			// spew.Dump(imps)
			panic(0)

			fmt.Println("vvv", types.ExprString(rev[p.Type()]))
			printer.Fprint(os.Stdout, fset, rev[p.Type()])
			fmt.Println("^^^")

			spew.Dump(p.Type().Underlying(), types.ExprString(rev[p.Type()]), "-=-=-=-")
			n = un.get(capitalize(n))
			q = append(q, types.NewField(0, pkg, n, p.Type(), false))
			tags = append(tags, fmt.Sprintf(`json:"%s"`, toSnake(n)))
		}
		st := types.NewStruct(q, tags)
		s := types.NewTypeName(0, pkg, name+"Request", st)

		fmt.Println("----\n", types.ObjectString(s, RelativeTo(pkg)))
	}
	return nil

	// err = updateFuncMap(fset, f, fns)
	// if err != nil {
	// 	oops(err.Error())
	// }
	// for n, f := range fns {
	// 	if f.f == nil {
	// 		log.Fatalln(n, "not found")
	// 	}
	// 	err = printer.Fprint(os.Stdout, f.fset, f.f)
	// 	fmt.Println("\n")
	// 	// fmt.Println(">>>", name(f.f))
	// 	if err != nil {
	// 		log.Fatalln(err)
	// 	}

	// 	for _, x := range f.f.Params.List {
	// 		fmt.Print("names: ")
	// 		for _, n := range x.Names {
	// 			fmt.Print(n.Name, "")
	// 		}
	// 		fmt.Println("")
	// 		printer.Fprint(os.Stdout, f.fset, x.Type)
	// 		fmt.Println("\n")
	// 		spew.Dump(x.Type)
	// 		if isError(x.Type) {
	// 			fmt.Println("error")
	// 		}
	// 		if isContext(f.fset, x.Type) {
	// 			fmt.Println("it is a context")
	// 		}
	// 		fmt.Println("\n---")
	// 	}

	// 	fmt.Println("out:\n")

	// 	for _, x := range f.f.Results.List {
	// 		fmt.Print("names: ")
	// 		for _, n := range x.Names {
	// 			fmt.Print(n.Name, "")
	// 		}
	// 		fmt.Println("")
	// 		printer.Fprint(os.Stdout, f.fset, x.Type)
	// 		fmt.Println("\n")
	// 		spew.Dump(x.Type)
	// 		if isError(x.Type) {
	// 			fmt.Println("it is an error")
	// 		}
	// 		fmt.Println("\n---")
	// 	}

	// }
}

func isError(n ast.Node) bool {
	switch t := n.(type) {
	case *ast.Ident:
		return t.Name == "error"
	}
	return false
}

func isContext(fset *token.FileSet, n ast.Node) bool {
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, n)
	return buf.String() == "context.Context"
}

// return a function name for free functions and
// <receiver_type (with start removed)>.<function name>...
func name(f *ast.FuncDecl) string {
	if f.Recv == nil || len(f.Recv.List) == 0 {
		return f.Name.Name
	}
	r := f.Recv.List[0]
	var tn string
	switch t := r.Type.(type) {
	case *ast.StarExpr:
		tn = t.X.(*ast.Ident).Name
	case *ast.Ident:
		tn = t.Name
	}
	return tn + "." + f.Name.Name
}

// oops prints the msg, usage and then exits with error code 2.
func oops(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	_, f := filepath.Split(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage:\n%s ${flags} ${rewrite rules}\n", f)
	fmt.Fprintln(os.Stderr, `Rewrite rules: each rewrite rule is either old->new (will replace old with new), or
old->${imports}::new (will replace old with new, adding imports if needed).
Imports are comma-separated.
`)
	fmt.Fprintln(os.Stderr, "Flags:")
	flag.PrintDefaults()
	os.Exit(2)
}

type g struct {
	pkg string
	w   bool
	typ string

	info types.Info
	rev  map[types.Type]ast.Expr
}

var gen g

type uniqueNames map[string]bool

func (u uniqueNames) get(base string) string {
	c := 0
	n := base
	for u[n] {
		c++
		n = fmt.Sprintf("%s%d", base, c)
	}
	u[n] = true
	return n
}

func filterOut(f []*ast.Field, pred func(ast.Node) bool) []*ast.Field {
	r := []*ast.Field{}
	for _, x := range f {
		if pred(x.Type) {
			continue
		}
		r = append(r, x)
	}
	return r
}

func buildStruct(name string, fset *token.FileSet, fs []*ast.Field) ast.Node {
	fl := []*ast.Field{}
	un := uniqueNames{}
	str := func(node interface{}) string {
		var buf bytes.Buffer
		printer.Fprint(&buf, fset, node)
		return buf.String()
	}
	for _, p := range fs {
		if isContext(fset, p.Type) {
			continue
		}
		if len(p.Names) == 0 {
			name := un.get(capitalize(str(p.Type)))
			field := &ast.Field{
				Type:    p.Type,
				Comment: &ast.CommentGroup{List: []*ast.Comment{&ast.Comment{Text: fmt.Sprintf(" // `json:\"%s\"`", toSnake(name))}}},
				Names:   []*ast.Ident{&ast.Ident{Name: name}},
			}
			fl = append(fl, field)
		}
		for _, name := range p.Names {
			field := &ast.Field{
				Type:    p.Type,
				Comment: &ast.CommentGroup{List: []*ast.Comment{&ast.Comment{Text: fmt.Sprintf(" // `json:\"%s\"`", name.Name)}}},
				Names:   []*ast.Ident{&ast.Ident{Name: capitalize(name.Name)}},
			}
			fl = append(fl, field)
		}
	}

	ts := &ast.TypeSpec{
		Name: &ast.Ident{Name: name},
		Type: &ast.StructType{
			Fields: &ast.FieldList{
				List: fl,
			},
		},
	}
	return ts
}

func x(fset *token.FileSet, name *ast.Ident, f *ast.FuncType) (req ast.Node) {
	return buildStruct(
		name.Name+"Request",
		fset,
		filterOut(f.Params.List, func(n ast.Node) bool {
			return isContext(fset, n)
		}),
	)
}

func updateFuncMap(pkg *types.Package, fns map[string]*types.Signature) error {
	scope := pkg.Scope()
	names := scope.Names()
	for _, n := range names {
		obj := scope.Lookup(n)
		if !obj.Exported() {
			continue
		}
		tn, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}
		if _, ok = fns[tn.Name()]; !ok {
			continue
		}
		named, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}
		sig, ok := named.Underlying().(*types.Signature)
		if !ok {
			continue
		}
		fns[n] = sig
	}
	return nil
}

func find(pkg *types.Package, iface string) (*types.Interface, error) {
	scope := pkg.Scope()
	names := scope.Names()
	for _, n := range names {
		obj := scope.Lookup(n)

		tn, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}
		if tn.Name() != iface {
			continue
		}
		if !obj.Exported() {
			return nil, fmt.Errorf("%s should exported", iface)
		}
		t := tn.Type().Underlying()
		i, ok := t.(*types.Interface)
		if !ok {
			return nil, fmt.Errorf("exptected interface, got %s for %s", t, iface)
		}
		return i, nil
	}
	return nil, errors.New("not found")
}

func uncapitalize(s string) string {
	if s == "" {
		return s
	}
	rs := []rune(s)
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	rs := []rune(s)
	rs[0] = unicode.ToUpper(rs[0])
	return string(rs)
}

func toSnake(s string) string {
	if s == "" {
		return s
	}
	parts := []string{}
	rs := []rune(s)
	start := 0
	for i, r := range rs {
		if i > 0 && unicode.IsUpper(r) {
			parts = append(parts, string(rs[start:i]))
			start = i
		}
		rs[i] = unicode.ToLower(r)
	}
	parts = append(parts, string(rs[start:len(rs)]))
	return strings.Join(parts, "_")
}

func toList(t *types.Tuple) []*types.Var {
	var r []*types.Var
	for i := 0; i < t.Len(); i++ {
		r = append(r, t.At(i))
	}
	return r
}

func RelativeTo(pkg *types.Package) types.Qualifier {
	if pkg == nil {
		return nil
	}
	return func(other *types.Package) string {
		if pkg == other {
			return "" // same package; unqualified
		}
		return other.Name()
	}
}

func isValidIdentifier(id string) bool {
	if id == "" || id == "_" {
		return false
	}
	for i, r := range id {
		if !isLetter(r) && (i == 0 || !isDigit(r)) {
			return false
		}
	}
	return true
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= 0x80 && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}

type packager interface {
	Pkg() *types.Package
}

func updateDeps(node ast.Node, info types.Info, imps map[types.Object]*ast.SelectorExpr) {
	ast.Inspect(node, func(n ast.Node) bool {
		s, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true // Keep going.
		}
		obj := info.Uses[s.Sel]
		imps[obj] = s
		return false // Do not go deeper.
	})
}

type multiErr []error

func (e *multiErr) Add(err error) {
	if err == nil {
		return
	}
	*e = append(*e, err)
}

func (e multiErr) ToErr() error {
	if len(e) == 0 {
		return nil
	}
	return e
}

func (e multiErr) Error() string {
	s := make([]string, 0, len(e))
	for _, err := range e {
		s = append(s, err.Error())
	}
	return strings.Join(s, " && ")
}
