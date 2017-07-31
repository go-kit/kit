package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode"
)

type (
	sourceContext struct {
		pkg        *ast.Ident
		imports    []*ast.ImportSpec
		interfaces []iface
	}

	iface struct {
		name, stubname, rcvrName *ast.Ident
		methods                  []method
	}

	method struct {
		name            *ast.Ident
		params          []arg
		results         []arg
		structsResolved bool
	}

	arg struct {
		name, asField *ast.Ident
		typ           ast.Expr
	}
)

func id(name string) *ast.Ident {
	return ast.NewIdent(name)
}

func sel(ids ...*ast.Ident) ast.Expr {
	switch len(ids) {
	default:
		return &ast.SelectorExpr{
			X:   sel(ids[:len(ids)-1]...),
			Sel: ids[len(ids)-1],
		}
	case 1:
		return ids[0]
	case 0:
		panic("zero ids to sel()")
	}
}

func fieldList(fn func(arg) *ast.Field, args ...arg) *ast.FieldList {
	fl := &ast.FieldList{List: []*ast.Field{}}
	for _, a := range args {
		fl.List = append(fl.List, fn(a))
	}
	return fl
}

func blockStmt(stmts ...ast.Stmt) *ast.BlockStmt {
	return &ast.BlockStmt{
		List: stmts,
	}
}

func structDecl(name *ast.Ident, fields *ast.FieldList) ast.Decl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: name,
				Type: &ast.StructType{
					Fields: fields,
				},
			},
		},
	}
}

func mustParseExpr(s string) ast.Node {
	n, err := parser.ParseExpr(s)
	if err != nil {
		panic(err)
	}
	return n
}

func pasteStmts(body *ast.BlockStmt, idx int, stmts []ast.Stmt) {
	list := body.List
	prefix := list[:idx]
	suffix := make([]ast.Stmt, len(list)-idx-1)
	copy(suffix, list[idx+1:])

	body.List = append(append(prefix, stmts...), suffix...)
}

func (sc *sourceContext) validate() error {
	if len(sc.interfaces) != 1 {
		return fmt.Errorf("found %d interfaces, expecting exactly 1", len(sc.interfaces))
	}
	for _, i := range sc.interfaces {
		for _, m := range i.methods {
			if len(m.results) < 1 {
				return fmt.Errorf("method %q of interface %q has no result types!", m.name, i.name)
			}
		}
	}
	return nil
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

/*
func NewHTTPHandler(endpoints Endpoints) http.Handler {
	m := http.NewServeMux()
	m.Handle("/bar", httptransport.NewServer(
		endpoints.Bar,
		DecodeBarRequest,
		EncodeBarResponse,
	))
	return m
}
*/
func (i iface) httpHandler() ast.Decl {
	handlerFn := mustParseExpr(`func (endpoints Endpoints) http.Handler {
		m := http.NewServeMux()
		replaceWithHandleCalls()
		return m
	}`).(*ast.FuncLit)

	handleCalls := []ast.Stmt{}
	for _, m := range i.methods {
		handleCall := mustParseExpr(`m.Handle("", httptransport.NewServer())`).(*ast.CallExpr)

		handleCall.Args[0].(*ast.BasicLit).Value = `"` + m.pathName() + `"`

		handleCall.Args[1].(*ast.CallExpr).Args =
			[]ast.Expr{sel(id("endpoints"), m.name), m.decodeFuncName(), m.encodeFuncName()}

		handleCalls = append(handleCalls, &ast.ExprStmt{X: handleCall})
	}

	pasteStmts(handlerFn.Body, 1, handleCalls)

	return &ast.FuncDecl{
		Name: id("NewHTTPHandler"),
		Type: handlerFn.Type,
		Body: handlerFn.Body,
	}
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
	i.rcvrName = id(unexport(inventName(i.name, scope).Name))
	return i.rcvrName
}

func (m method) definition(ifc iface) ast.Decl {
	notImpl := mustParseExpr(`func() {return "", errors.New("not implemented")}`).(*ast.FuncLit)

	parms := m.funcParams()
	if m.hasContext() {
		parms = &ast.FieldList{
			List: append([]*ast.Field{{
				Names: []*ast.Ident{ast.NewIdent("ctx")},
				Type:  sel(id("context"), id("Context")),
			}}, parms.List...),
		}
	}

	return &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{ifc.reciever()}},
		Name: m.name,
		Type: &ast.FuncType{
			Params:  parms,
			Results: m.funcResults(),
		},
		Body: notImpl.Body,
	}
}

/*
	func makeBarEndpoint(s stubFooService) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			req := request.(barrequest)
			s, err := s.bar(ctx, req.i, req.s)
			return barresponse{s: s, err: err}, nil
		}
	}
*/
func (m method) endpointMaker(ifc iface) ast.Decl {
	endpointFn := mustParseExpr(`func() {return func(ctx context.Context, request interface{}) (interface{}, error) { }}`).(*ast.FuncLit)
	scope := ast.NewScope(nil)
	scope.Insert(ast.NewObj(ast.Var, ifc.receiverName().Name))
	scope.Insert(ast.NewObj(ast.Var, "ctx"))
	scope.Insert(ast.NewObj(ast.Var, "req"))

	anonFunc := endpointFn.Body.List[0].(*ast.ReturnStmt).Results[0].(*ast.FuncLit)
	if !m.hasContext() { // is this the right thing?
		anonFunc.Type.Params.List = anonFunc.Type.Params.List[1:]
	}

	castReq := mustParseExpr(`func() {req := request.(NOTTHIS)}`).(*ast.FuncLit).Body.List[0].(*ast.AssignStmt)
	castReq.Rhs[0].(*ast.TypeAssertExpr).Type = m.requestStructName()

	callMethod := m.called(ifc, scope, "ctx", "req")

	returnResponse := &ast.ReturnStmt{
		Results: []ast.Expr{m.wrapResult(callMethod.Lhs), id("nil")},
	}

	anonFunc.Body = blockStmt(
		castReq,
		callMethod,
		returnResponse,
	)

	return &ast.FuncDecl{
		Name: m.endpointMakerName(),
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{ifc.reciever()}},
			Results: &ast.FieldList{List: []*ast.Field{
				&ast.Field{Type: sel(id("endpoint"), id("Endpoint"))},
			}},
		},
		Body: endpointFn.Body,
	}
}

func (m method) pathName() string {
	return "/" + strings.ToLower(m.name.Name)
}

func (m method) encodeFuncName() *ast.Ident {
	return id("Encode" + m.name.Name + "Response")
}

func (m method) decodeFuncName() *ast.Ident {
	return id("Decode" + m.name.Name + "Request")
}

func (m method) resultNames(scope *ast.Scope) []*ast.Ident {
	ids := []*ast.Ident{}
	for _, rz := range m.results {
		ids = append(ids, rz.chooseName(scope))
	}
	return ids
}

func (a arg) chooseName(scope *ast.Scope) *ast.Ident {
	if a.name == nil {
		return inventName(a.typ, scope)
	}
	return a.name
}

func inventName(t ast.Expr, scope *ast.Scope) *ast.Ident {
	n := baseName(t)
	for try := 0; ; try++ {
		nstr := pickName(n, try)
		obj := ast.NewObj(ast.Var, nstr)
		if alt := scope.Insert(obj); alt == nil {
			return ast.NewIdent(nstr)
		}
	}
}

func baseName(t ast.Expr) string {
	switch tt := t.(type) {
	default:
		panic(fmt.Sprintf("don't know how to choose a base name for #t (#v[0])", tt))
	case *ast.Ident:
		return tt.Name
	case *ast.SelectorExpr:
		return tt.Sel.Name
	}
}

func pickName(base string, idx int) string {
	if idx == 0 {
		switch base {
		default:
			return strings.Split(base, "")[0]
		case "error":
			return "err"
		}
	}
	return fmt.Sprintf("%s%d", base, idx)
}

func (m method) called(ifc iface, scope *ast.Scope, ctxName, spreadStruct string) *ast.AssignStmt {
	m.resolveStructNames()

	resNamesExpr := []ast.Expr{}
	for _, r := range m.resultNames(scope) {
		resNamesExpr = append(resNamesExpr, ast.Expr(r))
	}

	arglist := []ast.Expr{}
	if m.hasContext() {
		arglist = append(arglist, id(ctxName))
	}
	ssid := id(spreadStruct)
	for _, f := range m.requestStructFields().List {
		arglist = append(arglist, sel(ssid, f.Names[0]))
	}

	return &ast.AssignStmt{
		Lhs: resNamesExpr,
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun:  sel(ifc.receiverName(), m.name),
				Args: arglist,
			},
		},
	}
}

func (m method) wrapResult(results []ast.Expr) ast.Expr {
	kvs := []ast.Expr{}
	m.resolveStructNames()

	for i, a := range m.results {
		kvs = append(kvs, &ast.KeyValueExpr{
			Key:   ast.NewIdent(export(a.asField.Name)),
			Value: results[i],
		})
	}
	return &ast.CompositeLit{
		Type: m.responseStructName(),
		Elts: kvs,
	}
}

func (m method) resolveStructNames() {
	if m.structsResolved {
		return
	}
	m.structsResolved = true
	scope := ast.NewScope(nil)
	for i, p := range m.params {
		p.asField = p.chooseName(scope)
		m.params[i] = p
	}
	scope = ast.NewScope(nil)
	for i, r := range m.results {
		r.asField = r.chooseName(scope)
		m.results[i] = r
	}
}

func (m method) decoderFunc() ast.Decl {
	fn := mustParseExpr(`
		func (_ context.Context, r *http.Request) (interface{}, error) {
			var req ReqStructName
			err := json.NewDecoder(r.Body).Decode(&req)
			return req, err
		}
	`).(*ast.FuncLit)

	fn.Body.List[0].(*ast.DeclStmt).Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Type = m.requestStructName()

	return &ast.FuncDecl{
		Name: m.decodeFuncName(),
		Type: fn.Type,
		Body: fn.Body,
	}
}

func (m method) encoderFunc() ast.Decl {
	fn := mustParseExpr(`
		func (_ context.Context, w http.ResponseWriter, response interface{}) error {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			return json.NewEncoder(w).Encode(response)
		}
	`).(*ast.FuncLit)

	return &ast.FuncDecl{
		Name: m.encodeFuncName(),
		Type: fn.Type,
		Body: fn.Body,
	}
}

func (m method) endpointMakerName() *ast.Ident {
	return id("make" + m.name.Name + "Endpoint")
}

func (m method) requestStruct() ast.Decl {
	m.resolveStructNames()
	return structDecl(m.requestStructName(), m.requestStructFields())
}

func (m method) responseStruct() ast.Decl {
	m.resolveStructNames()
	return structDecl(m.responseStructName(), m.responseStructFields())
}

func (m method) hasContext() bool {
	if len(m.params) < 1 {
		return false
	}
	carg := m.params[0].typ
	// ugh. this is maybe okay for the one-off, but a general case for matching
	// types would be helpful
	if sel, is := carg.(*ast.SelectorExpr); is && sel.Sel.Name == "Context" {
		if id, is := sel.X.(*ast.Ident); is && id.Name == "context" {
			return true
		}
	}
	return false
}

func (m method) nonContextParams() []arg {
	if m.hasContext() {
		return m.params[1:]
	}
	return m.params
}

func (m method) funcParams() *ast.FieldList {
	return fieldList(func(a arg) *ast.Field {
		return a.field()
	}, m.nonContextParams()...)
}

func (m method) funcResults() *ast.FieldList {
	return fieldList(func(a arg) *ast.Field {
		return a.result()
	}, m.results...)
}

func (m method) requestStructName() *ast.Ident {
	return id(export(m.name.Name) + "Request")
}

func (m method) requestStructFields() *ast.FieldList {
	return fieldList(func(a arg) *ast.Field {
		return a.exported()
	}, m.nonContextParams()...)
}

func (m method) responseStructName() *ast.Ident {
	return id(export(m.name.Name) + "Response")
}

func (m method) responseStructFields() *ast.FieldList {
	return fieldList(func(a arg) *ast.Field {
		return a.exported()
	}, m.results...)
}

func (a arg) field() *ast.Field {
	return &ast.Field{
		Names: []*ast.Ident{a.name},
		Type:  a.typ,
	}
}

func (a arg) result() *ast.Field {
	return &ast.Field{
		Names: nil,
		Type:  a.typ,
	}
}

func export(s string) string {
	return strings.Title(s)
}

func unexport(s string) string {
	first := true
	return strings.Map(func(r rune) rune {
		if first {
			first = false
			return unicode.ToLower(r)
		}
		return r
	}, s)
}

func (a arg) exported() *ast.Field {
	return &ast.Field{
		Names: []*ast.Ident{id(export(a.asField.Name))},
		Type:  a.typ,
	}
}
