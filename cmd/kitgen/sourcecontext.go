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
		name, stubname *ast.Ident
		methods        []method
	}

	method struct {
		name    *ast.Ident
		params  []arg
		results []arg
	}

	arg struct {
		name *ast.Ident
		typ  ast.Expr
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
	suffix := list[idx+1:]
	body.List = append(append(prefix, stmts...), suffix...)
}

func (sc *sourceContext) validate() error {
	if len(sc.interfaces) != 1 {
		return fmt.Errorf("found %d interfaces, expecting exactly 1", len(sc.interfaces))
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
			[]ast.Expr{sel(id("endpoints"), m.name), m.encodeFuncName(), m.decodeFuncName()}

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
	r := strings.NewReader(i.name.Name)
	ch, _, err := r.ReadRune()
	if err != nil {
		panic(err)
	}
	return id(string(unicode.ToLower(ch)))
}

func (m method) definition(ifc iface) ast.Decl {
	notImpl := mustParseExpr(`func() {return "", errors.New("not implemented")}`).(*ast.FuncLit)

	return &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{ifc.reciever()}},
		Name: m.name,
		Type: &ast.FuncType{
			Params:  m.funcParams(),
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

	anonFunc := endpointFn.Body.List[0].(*ast.ReturnStmt).Results[0].(*ast.FuncLit)
	if !m.hasContext() { // is this the right thing?
		anonFunc.Type.Params.List = anonFunc.Type.Params.List[1:]
	}

	castReq := mustParseExpr(`func() {req := request.(NOTTHIS)}`).(*ast.FuncLit).Body.List[0].(*ast.AssignStmt)
	castReq.Rhs[0].(*ast.TypeAssertExpr).Type = m.responseStructName()

	callMethod := m.called(ifc, "ctx", "req")

	returnResponse := &ast.ReturnStmt{
		Results: []ast.Expr{m.wrapResult(), id("nil")},
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
				&ast.Field{Type: &ast.InterfaceType{}},
				&ast.Field{Type: id("error")},
			}},
		},
		Body: endpointFn.Body,
	}
}

func (m method) pathName() string {
	return "/" + strings.ToLower(m.name.Name)
}

func (m method) encodeFuncName() *ast.Ident {
	return id("Decode" + m.name.Name + "Request")
}

func (m method) decodeFuncName() *ast.Ident {
	return id("Encode" + m.name.Name + "Response")
}

func (m method) resultNames() []*ast.Ident {
	ids := []*ast.Ident{}
	for _, rz := range m.results {
		ids = append(ids, rz.name)
	}
	return ids
}

func (m method) called(ifc iface, ctxName, spreadStruct string) ast.Stmt {
	resNamesExpr := []ast.Expr{}
	for _, r := range m.resultNames() {
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

func (m method) wrapResult() ast.Expr {
	kvs := []ast.Expr{}

	for _, a := range m.results {
		kvs = append(kvs, &ast.KeyValueExpr{
			Key:   a.exported().Names[0],
			Value: a.name,
		})
	}
	return &ast.CompositeLit{
		Type: m.responseStructName(),
		Elts: kvs,
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
	return structDecl(m.requestStructName(), m.requestStructFields())
}

func (m method) responseStruct() ast.Decl {
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
		Names: []*ast.Ident{},
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
		Names: []*ast.Ident{id(export(a.name.Name))},
		Type:  a.typ,
	}
}
