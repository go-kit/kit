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
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/sasha-s/go-inline/goinline"

	"golang.org/x/tools/go/types"

	_ "golang.org/x/tools/go/gcimporter"
)

func main() {
	var gen generator

	flag.StringVar(&gen.pkgName, "package", "", "")
	flag.BoolVar(&gen.w, "w", false, "will (over)write files if set. Prints to stdout otherwise.")
	flag.StringVar(&gen.typ, "type", "", "type name; must be set")
	var binding string
	flag.StringVar(&binding, "binding", "", "comma-separated list of bindings to generate. Bindings to choose from: http,rpc")

	flag.Parse()

	if gen.pkgName == "" {
		oops("-package should not be empty")
	}

	parts := strings.Split(binding, ",")
	gen.bindings = map[string]string{}
	for _, b := range parts {
		if b == "" {
			continue
		}
		v, ok := knownBindings[b]
		if !ok {
			oops(fmt.Sprintf("unknown binding `%s`", b))
		}
		gen.bindings[b] = v
	}

	err := gen.generate()
	if err != nil {
		log.Fatalln(err)
	}
}

func (g *generator) generate() error {
	pk, err := build.Default.Import(g.pkgName, "", 0)
	if err != nil {
		return err
	}
	g.dir = pk.Dir
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
	if len(fs) == 0 {
		return errors.New("no go files")
	}
	// Typecheck and collect multiErr.
	var errs multiErr
	conf := types.Config{
		IgnoreFuncBodies: true,
		Error: func(err error) {
			errs.Add(err)
		},
	}
	g.info = types.Info{
		Types:      map[ast.Expr]types.TypeAndValue{},
		Defs:       map[*ast.Ident]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
		Uses:       map[*ast.Ident]types.Object{},
	}
	g.pkg, err = conf.Check(fs[0].Name.Name, fset, fs, &g.info)
	if err != nil {
		return err
	}
	if len(errs) > 0 {
		return errs.ToErr()
	}
	g.iface, err = find(g.pkg, g.typ)
	if err != nil {
		return err
	}

	// Build the reverse map.
	g.rev = map[types.Type]ast.Expr{}
	for k, v := range g.info.Types {
		g.rev[v.Type] = k
	}

	iface := Interface{
		Name: g.typ,
		Pkg:  g.pkg.Name(),
		Imports: []string{
			"github.com/go-kit/kit/endpoint",
			"golang.org/x/net/context",
		}}
	iface.Receiver = receiver(iface)
	for i := 0; i < g.iface.NumMethods(); i++ {
		errs.Add(g.processFunc(g.iface.Method(i), &iface))
	}
	if len(errs) != 0 {
		return errs.ToErr()
	}

	iface.Imports = uniq(iface.Imports)

	w, close, err := g.open(filepath.Join(g.dir, toSnake(g.typ)+"_endpoints.go"))
	if err != nil {
		return err
	}

	defer close()
	return endpointF.Execute(w, iface)
}

var endpointF = template.Must(template.New("endpointf").Parse(`
package {{.Pkg}}

import (
{{range .Imports}}	"{{.}}"
{{end}})

{{$endpoints := print .Name "Endpoints"}}
type {{$endpoints}} struct {
	x {{.Name}}
}

func Make{{$endpoints}} (x {{.Name}}) {{$endpoints}} {
	return {{$endpoints}}{x}
}
{{$client := print .Name "Client"}}
type {{$client}} struct {
	f func(string) endpoint.Endpoint
}

func Make{{$client}} (f func(string) endpoint.Endpoint) {{$client}} {
	return {{$client}}{f}
}
{{$rec := .Receiver}}
{{range .M}}
func ({{$rec}} {{$endpoints}}) {{.Name}} (ctx context.Context, request interface{}) (interface{}, error) {
	select {
	default:
	case <-ctx.Done():
		return nil, endpoint.ErrContextCanceled
	}
	req, ok := request.({{.Req.StructDef.Name}})
	if !ok {
		return nil, endpoint.ErrBadCast
	}
	var err error
	var resp {{.Resp.StructDef.Name}}
	{{range $i,$v:=.Resp.Args}}{{if ne $i 0}}, {{end}}{{call . "resp"}}{{end}} = {{$rec}}.x.{{.Name}}({{range $i,$v:=.Req.Args}}{{if ne $i 0}}, {{end}}{{call . "req"}}{{end}})
	return resp, err
}

func ({{$rec}} {{$client}}) {{.Name}} ({{range $i,$v:=.Req.Args2}}{{if ne $i 0}}, {{end}}{{.Name}} {{.Type}}{{end}}) ({{range $i,$v:=.Resp.Args2}}{{if ne $i 0}}, {{end}}{{.Name}} {{.Type}}{{end}}) {
	{{if not .Req.HasCtx}}{{.Req.CtxName}} := context.TODO(){{end}}
	{{if not .Resp.HasErr}}var {{.Resp.ErrName}} error{{end}}
	var req {{.Req.StructDef.Name}}
	{{range $i,$v:=.Req.Args}}{{if ne $i 0}}, {{end}}{{call . "req"}}{{end}} = {{range $i,$v:=.Req.Args2}}{{if ne $i 0}}, {{end}}{{.Name}}{{end}}
	var raw interface{}
	raw, {{.Resp.ErrName}} = {{$rec}}.f("{{.Name}}")({{.Req.CtxName}} , req)
	if err != nil {
		{{if not .Resp.HasErr}}panic(err){{else}}return{{end}}
	}
	resp, ok := raw.({{.Resp.StructDef.Name}})
	if !ok {
		{{.Resp.ErrName}}  = endpoint.ErrBadCast
		{{if not .Resp.HasErr}}panic(err){{else}}return{{end}}
	}

	{{range $i,$v:=.Resp.Args2}}{{if ne $i 0}}, {{end}}{{.Name}}{{end}} = {{range $i,$v:=.Resp.Args}}{{if ne $i 0}}, {{end}}{{call . "resp"}}{{end}}
	return
}

{{end}}
`))

type structT struct {
	imports []string
	name    string
}

type structDef struct {
	Pkg     string
	Imports []string
	Name    string
	Fields  []struct {
		Name string
		Tag  string
		Type string
	}
}

type argName func(string) string

func (a argName) String() string {
	return fmt.Sprintf("`λ` -> `%s`", a(`λ`))
}

type args struct {
	Args             []argName // For referencing as a field name.
	Args2            []struct{ Name, Type string }
	StructDef        structDef
	ErrName, CtxName string
	HasCtx, HasErr   bool
}

type method struct {
	Name string
	Req  args
	Resp args
}

type Interface struct {
	Pkg      string
	Imports  []string
	Name     string
	M        []method
	Receiver string
}

// t is a tuple for representing parameters or return values of  function.
func (g *generator) parse(name string, t *types.Tuple) *args {
	ps := toList(t)
	var fields []*types.Var
	var tags []string
	imports := map[types.Object]*ast.SelectorExpr{}
	un := uniqueNames{}
	m := &args{}
	for _, p := range ps {
		n := p.Name()
		if n == "" {
			n = p.Type().String()
			if !validIdentifier(n) {
				n = "p"
			}
		}
		n = un.get(capitalize(n))
		t := types.NewField(0, g.pkg, n, p.Type(), false)
		// Filter out context and error.
		switch p.Type().String() {
		case "golang.org/x/net/context.Context":
			ctxName := un.get("ctx")
			m.Args = append(m.Args, func(string) string { return ctxName })
			m.CtxName = ctxName
			m.HasCtx = true
			m.Args2 = append(m.Args2, struct{ Name, Type string }{ctxName, types.TypeString(t.Type(), relativeTo(g.pkg))})
			continue
		case "error":
			errName := un.get("err")
			m.Args = append(m.Args, func(string) string { return errName })
			m.ErrName = errName
			m.HasErr = true
			m.Args2 = append(m.Args2, struct{ Name, Type string }{errName, types.TypeString(t.Type(), relativeTo(g.pkg))})
			continue
		}
		m.Args2 = append(m.Args2, struct{ Name, Type string }{uncapitalize(n), types.TypeString(t.Type(), relativeTo(g.pkg))})
		updateDeps(g.rev[p.Type()], g.info, imports)
		// Make sure all the names are unique.
		m.Args = append(m.Args, func(s string) string { return fmt.Sprintf("%s.%s", s, n) })
		fields = append(fields, t)
		tags = append(tags, fmt.Sprintf(`json:"%s"`, toSnake(n)))
	}
	if !m.HasCtx {
		m.CtxName = un.get("ctx")
	}
	if !m.HasErr {
		m.ErrName = un.get("err")
	}
	imps := cleanImports(imports)
	m.StructDef = structDef{
		Pkg:     g.pkg.Name(),
		Imports: imps,
		Name:    name,
	}
	for i, v := range fields {
		m.StructDef.Fields = append(m.StructDef.Fields, struct {
			Name string
			Tag  string
			Type string
		}{
			Name: v.Name(),
			Type: types.TypeString(v.Type(), relativeTo(g.pkg)),
			Tag:  tags[i],
		})
	}
	return m
}

func (g *generator) writeType(name string, t *types.Tuple) (*args, error) {
	m := g.parse(name, t)

	w, close, err := g.open(filepath.Join(g.dir, toSnake(name)+".go"))
	if err != nil {
		return nil, err
	}
	defer close()

	return m, structF.Execute(w, m.StructDef)
}

// TODO: do not generate import() if import list is empty.
var structF = template.Must(template.New("structf").Parse(`
package {{.Pkg}}

import (
{{range .Imports}}	"{{.}}"
{{end}})

type {{.Name}} struct {
{{range .Fields}}    {{.Name}} {{.Type}} // ` + "`" + "{{.Tag}}`" + `
{{end}}}
`))

func (g *generator) processFunc(f *types.Func, iface *Interface) error {
	sig := f.Type().(*types.Signature)
	req := fmt.Sprintf("%s%sRequest", g.typ, f.Name())
	resp := fmt.Sprintf("%s%sResponse", g.typ, f.Name())
	reqS, err := g.writeType(req, sig.Params())
	if err != nil {
		return err
	}
	respS, err := g.writeType(resp, sig.Results())
	if err != nil {
		return err
	}
	iface.M = append(iface.M, method{Name: f.Name(), Req: *reqS, Resp: *respS})
	// Make sure the reciever, the params and return values do not overlap.
	un := uniqueNames{}
	// TODO: add imported packages.
	for _, x := range []string{iface.Receiver, "req", "resp", "raw", "ok", `break`, `default`, `func`, `interface`, `select`, `case`, `defer`, `go`, `map`, `struct`, `chan`, `else`, `goto`, `package`, `switch`, `const`, `fallthrough`, `if`, `range`, `type`, `continue`, `for`, `import`, `return`, `varbool`, `true`, `false`, `uint8`, `uint16`, `uint32`, `uint64`, `int8`, `int16`, `int32`, `int64`, `float32`, `float64`, `complex64`, `complex`, `128`, `string`, `int`, `uint`, `uintptr`, `byte`, `rune`, `iota`, `nil`, `appnd`, `copy`, `delete`, `len`, `cap`, `make`, `new`, `complex`, `real`, `imag`, `close`, `panic`, `recover`, `print`, `println`, `error`} {
		un.get(x)
	}
	for i, a := range reqS.Args2 {
		reqS.Args2[i].Name = un.get(a.Name)
	}
	for i, a := range respS.Args2 {
		respS.Args2[i].Name = un.get(a.Name)
	}

	imports := map[types.Object]*ast.SelectorExpr{}
	for _, p := range toList(sig.Params()) {
		updateDeps(g.rev[p.Type()], g.info, imports)
	}
	for _, p := range toList(sig.Results()) {
		updateDeps(g.rev[p.Type()], g.info, imports)
	}
	imp := cleanImports(imports)
	iface.Imports = append(iface.Imports, imp...)

	im := map[string]goinline.Target{
		"FunT":            goinline.Target{Ident: f.Name(), Imports: nil},
		"RequestT":        goinline.Target{Ident: req, Imports: nil},
		"ResponseT":       goinline.Target{Ident: resp, Imports: nil},
		"makeHTTPBinding": goinline.Target{Ident: fmt.Sprintf("Make%s%sHTTPBinding", g.typ, f.Name()), Imports: nil},
		"NetrpcBinding":   goinline.Target{Ident: fmt.Sprintf("%s%sNetrpcBinding", g.typ, f.Name()), Imports: nil, NoFiltering: true},
		"NewHTTPClient":   goinline.Target{Ident: fmt.Sprintf("New%s%sHTTPClient", g.typ, f.Name()), Imports: nil},
		"NewRPCClient":    goinline.Target{Ident: fmt.Sprintf("New%s%sRPCClient", g.typ, f.Name()), Imports: nil},
	}

	for _, pkg := range g.bindings {
		pk, err := build.Default.Import(pkg, "", 0)
		if err != nil {
			return err
		}

		for _, fn := range pk.GoFiles {
			full := filepath.Join(pk.Dir, fn)

			var fset = token.NewFileSet()
			ff, err := parser.ParseFile(fset, full, nil, parser.ParseComments)
			if err != nil {
				return err
			}

			err = goinline.Inline(fset, ff, im)
			if err != nil {
				return err
			}
			// Change the package name.
			ff.Name = &ast.Ident{Name: g.pkg.Name()}
			target := filepath.Join(g.dir, toSnake(fmt.Sprintf("%s%s_%s", g.typ, f.Name(), fn)))
			w, close, err := g.open(target)
			if err != nil {
				return err
			}
			defer close()
			err = printer.Fprint(w, fset, ff)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generator) open(name string) (w io.Writer, close func() error, err error) {
	if g.w == false {
		return os.Stdout, func() error { return nil }, nil
	}
	err = canWriteSafely(name)
	if err != nil {
		return nil, nil, err
	}
	f, err := os.Create(name)
	if err != nil {
		return nil, nil, err
	}
	_, err = fmt.Fprintln(f, string(preamble))
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	return f, f.Close, nil
}

// oops prints the msg, usage and then exits with error code 2.
func oops(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	_, f := filepath.Split(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage:\n%s ${flags}\n", f)
	fmt.Fprintln(os.Stderr, "Flags:")
	flag.PrintDefaults()
	os.Exit(2)
}

// canWriteSafely checks whether we can write the file safely.
// It is safe to write if either
// * fn does not exists.
// * fn exists and starts with the preamble.
func canWriteSafely(fn string) error {
	_, err := os.Stat(fn)
	if os.IsNotExist(err) {
		return nil
	}
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := make([]byte, len(preamble))
	_, err = io.ReadAtLeast(f, buf, len(buf))
	if err != nil {
		return fmt.Errorf("failed to read %d bytes: `%s`", len(buf), err)
	}
	if !bytes.Equal(buf, []byte(preamble)) {
		return fmt.Errorf("the preamble does not match. Exptected `%s`, got `%s`", string(preamble), string(buf))
	}
	return nil
}

type generator struct {
	pkgName string
	w       bool
	typ     string

	info     types.Info
	rev      map[types.Type]ast.Expr
	iface    *types.Interface
	pkg      *types.Package
	dir      string
	bindings map[string]string
}

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
	return nil, fmt.Errorf("%s not found in %s", iface, pkg.Name())
}

func receiver(iface Interface) string {
	if len(iface.Name) == 0 {
		return "x"
	}
	return string([]rune{unicode.ToLower(([]rune(iface.Name))[0])})
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	rs := []rune(s)
	rs[0] = unicode.ToUpper(rs[0])
	return string(rs)
}

func uncapitalize(s string) string {
	if s == "" {
		return s
	}
	rs := []rune(s)
	rs[0] = unicode.ToLower(rs[0])
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

func relativeTo(pkg *types.Package) types.Qualifier {
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

func validIdentifier(id string) bool {
	if id == "" || id == "_" {
		return false
	}
	for i, r := range id {
		if !unicode.IsLetter(r) && (i == 0 || !unicode.IsDigit(r)) {
			return false
		}
	}
	return true
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

func (m *multiErr) Add(err ...error) {
	for _, e := range err {
		if e == nil {
			continue
		}
		*m = append(*m, e)
	}
}

func (m multiErr) ToErr() error {
	if len(m) == 0 {
		return nil
	}
	return m
}

func (m multiErr) Error() string {
	s := make([]string, 0, len(m))
	for _, err := range m {
		s = append(s, err.Error())
	}
	return strings.Join(s, " && ")
}

func cleanImports(imports map[types.Object]*ast.SelectorExpr) []string {
	var imps []string
	for k, _ := range imports {
		imps = append(imps, k.Pkg().Path())
	}
	return uniq(imps)
}

func uniq(imps []string) []string {
	sort.Strings(imps)
	// Remove dups.
	to := 0
	for from, s := range imps {
		if from > 0 && s == imps[from-1] {
			continue
		}
		imps[to] = s
		to++
	}
	imps = imps[:to]
	// TODO: sort imports properly.
	return imps
}

var knownBindings map[string]string = map[string]string{
	"rpc":  "github.com/go-kit/kit/codegen/blueprints/rpc",
	"http": "github.com/go-kit/kit/codegen/blueprints/http",
}

const preamble = `// Do not edit! Generated by gokit-generate`
